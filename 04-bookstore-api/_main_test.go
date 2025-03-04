package bookstore

import (
	bytes "bytes"
	sql "database/sql"
	json "encoding/json"
	errors "errors"
	fmt "fmt"
	ioutil "io/ioutil"
	log "log"
	http "net/http"
	os "os"
	testing "testing"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	httptest "net/http/httptest"
	debug "runtime/debug"
	strings "strings"
	sync "sync"
	mock "github.com/stretchr/testify/mock"
	require "github.com/stretchr/testify/require"
	assert "github.com/stretchr/testify/assert"
)



var mockSaveBooks func([]Book) error
var mockGetBookById = func(id string) (Book, bool, error) {
	if id == "valid-id" {
		return Book{Id: "valid-id", Title: "Test Book", Author: "Test Author", Price: "10.99", Imageurl: "test-url"}, true, nil
	}
	if id == "error-id" {
		return Book{}, false, errors.New("database error")
	}
	return Book{}, false, nil
}
var mockGetBooks func() ([]Book, error)

type MockBookStore struct {
	mock.Mock
}


/*
ROOST_METHOD_HASH=checkError_c8f50cac3c
ROOST_METHOD_SIG_HASH=checkError_45ba6f7a64

FUNCTION_DEF=func checkError(err error) // print logs in console


*/
func TestCheckError(t *testing.T) {

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	tests := []struct {
		name       string
		err        error
		wantOutput string
	}{
		{
			name:       "Logging an Error Message when an Error Occurs",
			err:        errors.New("sample error"),
			wantOutput: "Error - sample error",
		},
		{
			name:       "No Logging when No Error is Passed",
			err:        nil,
			wantOutput: "",
		},
		{
			name:       "Logging a Complex Error Object",
			err:        fmt.Errorf("wrapped error: %w", errors.New("base error")),
			wantOutput: "Error - wrapped error: base error",
		},
		{
			name: "Logging a JSON Parsing Error",
			err: func() error {
				var book Book
				data := []byte(`{"id": "1", "title": "Go Programming", "author": "John Doe", "price": "invalid_price"}`)
				return json.Unmarshal(data, &book)
			}(),
			wantOutput: "Error - json: cannot unmarshal string into Go struct field Book.price of type string",
		},
		{
			name: "Logging an HTTP Request Failure",
			err: func() error {
				_, err := http.Get("http://invalid.url")
				return err
			}(),
			wantOutput: "Error - Get \"http://invalid.url\":",
		},
		{
			name: "Logging an I/O Operation Failure",
			err: func() error {
				_, err := ioutil.ReadFile("non_existent_file.txt")
				return err
			}(),
			wantOutput: "Error - open non_existent_file.txt:",
		},
		{
			name: "Logging a Custom Struct Error",
			err: func() error {
				type CustomError struct {
					msg string
				}

				func (e *CustomError) Error() string { return e.msg }

				return &CustomError{msg: "custom error occurred"}
			}(),
			wantOutput: "Error - custom error occurred",
		},
		{
			name: "Logging an Error from a Failed Database Operation",
			err: func() error {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("error initializing sqlmock: %v", err)
				}
				defer db.Close()
				mock.ExpectQuery("SELECT * FROM books").WillReturnError(errors.New("database connection failed"))
				_, err = db.Query("SELECT * FROM books")
				return err
			}(),
			wantOutput: "Error - database connection failed",
		},
		{
			name: "Logging an Error from an External API Call",
			err: func() error {
				return errors.New("failed to fetch data from API")
			}(),
			wantOutput: "Error - failed to fetch data from API",
		},
		{
			name: "Logging an Error from a Failed Marshaling Operation",
			err: func() error {
				ch := make(chan int)
				_, err := json.Marshal(ch)
				return err
			}(),
			wantOutput: "Error - json: unsupported type: chan int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			buf.Reset()

			checkError(tt.err)

			loggedOutput := buf.String()

			if tt.wantOutput != "" && !bytes.Contains(buf.Bytes(), []byte(tt.wantOutput)) {
				t.Errorf("Expected log output to contain: %q, but got: %q", tt.wantOutput, loggedOutput)
			} else if tt.wantOutput == "" && loggedOutput != "" {
				t.Errorf("Expected no log output, but got: %q", loggedOutput)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=handleAddBook_1affcd2057
ROOST_METHOD_SIG_HASH=handleAddBook_244c083dc8

FUNCTION_DEF=func handleAddBook(w http.ResponseWriter, r *http.Request) // to add new book


*/
func TestHandleAddBook(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		requestBody    string
		mockSaveError  error
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "Scenario 1: Successfully Add a New Book",
			method:         "POST",
			requestBody:    `[{"id":"1","title":"Go Lang","author":"John Doe","price":"10.99","image_url":"http://example.com/book.jpg"}]`,
			mockSaveError:  nil,
			expectedStatus: http.StatusOK,
			expectedMsg:    "New book added successfully",
		},
		{
			name:           "Scenario 2: Reject Non-POST Requests",
			method:         "GET",
			requestBody:    ``,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedMsg:    "GET - Method not allowed",
		},
		{
			name:           "Scenario 3: Handle Malformed JSON in Request Body",
			method:         "POST",
			requestBody:    `{"id": "1", "title": "Invalid JSON"`,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Bad Request",
		},
		{
			name:           "Scenario 4: Handle Empty Request Body",
			method:         "POST",
			requestBody:    ``,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Bad Request",
		},
		{
			name:           "Scenario 5: Handle Internal Server Error on Save Failure",
			method:         "POST",
			requestBody:    `[{"id":"2","title":"Go Advanced","author":"Jane Doe","price":"15.99","image_url":"http://example.com/book2.jpg"}]`,
			mockSaveError:  fmt.Errorf("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "Internal server error",
		},
		{
			name:           "Scenario 6: Handle Multiple Books in a Single Request",
			method:         "POST",
			requestBody:    `[{"id":"3","title":"Book 1","author":"Author 1","price":"9.99","image_url":"http://example.com/book1.jpg"},{"id":"4","title":"Book 2","author":"Author 2","price":"12.99","image_url":"http://example.com/book2.jpg"}]`,
			mockSaveError:  nil,
			expectedStatus: http.StatusOK,
			expectedMsg:    "New book added successfully",
		},
		{
			name:           "Scenario 7: Handle JSON Parsing Failure",
			method:         "POST",
			requestBody:    `invalid_json`,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "Bad Request",
		},
		{
			name:           "Scenario 8: Ensure Book List is Persisted Correctly",
			method:         "POST",
			requestBody:    `[{"id":"5","title":"Persisted Book","author":"Persistent Author","price":"20.99","image_url":"http://example.com/book3.jpg"}]`,
			mockSaveError:  nil,
			expectedStatus: http.StatusOK,
			expectedMsg:    "New book added successfully",
		},
		{
			name:           "Scenario 9: Handle Large Request Body",
			method:         "POST",
			requestBody:    strings.Repeat(`{"id":"6","title":"Large Book","author":"Big Author","price":"25.99","image_url":"http://example.com/book4.jpg"},`, 1000),
			mockSaveError:  nil,
			expectedStatus: http.StatusOK,
			expectedMsg:    "New book added successfully",
		},
		{
			name:           "Scenario 10: Handle Duplicate Book Entries",
			method:         "POST",
			requestBody:    `[{"id":"1","title":"Go Lang","author":"John Doe","price":"10.99","image_url":"http://example.com/book.jpg"}]`,
			mockSaveError:  nil,
			expectedStatus: http.StatusOK,
			expectedMsg:    "New book added successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			mockSaveBooks = func(books []Book) error {
				return tt.mockSaveError
			}

			req := httptest.NewRequest(tt.method, "/addBook", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handleAddBook(rec, req)

			resp := rec.Result()
			body, _ := ioutil.ReadAll(resp.Body)

			var msg Message
			_ = json.Unmarshal(body, &msg)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if msg.Msg != tt.expectedMsg {
				t.Errorf("Expected message '%s', got '%s'", tt.expectedMsg, msg.Msg)
			}
		})
	}
}

func saveBooks(books []Book) error {
	if mockSaveBooks != nil {
		return mockSaveBooks(books)
	}

	return nil
}


/*
ROOST_METHOD_HASH=handleDeleteBookById_29d0101c0b
ROOST_METHOD_SIG_HASH=handleDeleteBookById_27357b82d7

FUNCTION_DEF=func handleDeleteBookById(w http.ResponseWriter, r *http.Request) 

*/
func (m *MockBookStore) getBookById(bookId string) (Book, int, error) {
	args := m.Called(bookId)
	return args.Get(0).(Book), args.Int(1), args.Error(2)
}

func (m *MockBookStore) getBooks() ([]Book, error) {
	args := m.Called()
	return args.Get(0).([]Book), args.Error(1)
}

func (m *MockBookStore) saveBooks(books []Book) {
	m.Called(books)
}

func TestHandleDeleteBookById(t *testing.T) {
	mockStore := new(MockBookStore)

	tests := []struct {
		name             string
		bookID           string
		mockGetBookById  func()
		mockGetBooks     func()
		mockSaveBooks    func()
		expectedStatus   int
		expectedResponse string
	}{
		{
			name:   "Successfully Delete an Existing Book",
			bookID: "123",
			mockGetBookById: func() {
				mockStore.On("getBookById", "123").Return(Book{Id: "123", Title: "Go Lang"}, 0, nil)
			},
			mockGetBooks: func() {
				mockStore.On("getBooks").Return([]Book{{Id: "123", Title: "Go Lang"}}, nil)
			},
			mockSaveBooks: func() {
				mockStore.On("saveBooks", []Book{}).Return()
			},
			expectedStatus:   200,
			expectedResponse: `{"msg":"Book deleted successfully"}`,
		},
		{
			name:   "Attempt to Delete a Non-Existent Book",
			bookID: "999",
			mockGetBookById: func() {
				mockStore.On("getBookById", "999").Return(Book{}, -1, nil)
			},
			expectedStatus:   200,
			expectedResponse: `{"msg":"Book Not found"}`,
		},
		{
			name:   "Handle Internal Server Error from getBookById",
			bookID: "error",
			mockGetBookById: func() {
				mockStore.On("getBookById", "error").Return(Book{}, -1, errors.New("DB error"))
			},
			expectedStatus:   500,
			expectedResponse: `{"msg":"Internal server error"}`,
		},
		{
			name:             "Handle Case Where No ID is Provided in Request",
			bookID:           "",
			expectedStatus:   200,
			expectedResponse: `{"msg":"Book Not found"}`,
		},
		{
			name:   "Verify Correct Book is Deleted from the List",
			bookID: "456",
			mockGetBookById: func() {
				mockStore.On("getBookById", "456").Return(Book{Id: "456", Title: "Python"}, 1, nil)
			},
			mockGetBooks: func() {
				mockStore.On("getBooks").Return([]Book{
					{Id: "123", Title: "Go Lang"},
					{Id: "456", Title: "Python"},
					{Id: "789", Title: "Rust"},
				}, nil)
			},
			mockSaveBooks: func() {
				mockStore.On("saveBooks", []Book{
					{Id: "123", Title: "Go Lang"},
					{Id: "789", Title: "Rust"},
				}).Return()
			},
			expectedStatus:   200,
			expectedResponse: `{"msg":"Book deleted successfully"}`,
		},
		{
			name:   "Deleting the Only Book in the List",
			bookID: "only",
			mockGetBookById: func() {
				mockStore.On("getBookById", "only").Return(Book{Id: "only", Title: "Solo Book"}, 0, nil)
			},
			mockGetBooks: func() {
				mockStore.On("getBooks").Return([]Book{{Id: "only", Title: "Solo Book"}}, nil)
			},
			mockSaveBooks: func() {
				mockStore.On("saveBooks", []Book{}).Return()
			},
			expectedStatus:   200,
			expectedResponse: `{"msg":"Book deleted successfully"}`,
		},
		{
			name:   "Handle Concurrent Deletion Requests",
			bookID: "concurrent",
			mockGetBookById: func() {
				mockStore.On("getBookById", "concurrent").Return(Book{Id: "concurrent", Title: "Concurrency"}, 0, nil)
			},
			mockGetBooks: func() {
				mockStore.On("getBooks").Return([]Book{{Id: "concurrent", Title: "Concurrency"}}, nil)
			},
			mockSaveBooks: func() {
				mockStore.On("saveBooks", []Book{}).Return()
			},
			expectedStatus:   200,
			expectedResponse: `{"msg":"Book deleted successfully"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered, failing test: %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			if tt.mockGetBookById != nil {
				tt.mockGetBookById()
			}
			if tt.mockGetBooks != nil {
				tt.mockGetBooks()
			}
			if tt.mockSaveBooks != nil {
				tt.mockSaveBooks()
			}

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/delete?id=%s", tt.bookID), nil)
			rec := httptest.NewRecorder()

			handleDeleteBookById(rec, req)

			res := rec.Result()
			body, _ := ioutil.ReadAll(res.Body)

			require.Equal(t, tt.expectedStatus, res.StatusCode)
			require.JSONEq(t, tt.expectedResponse, string(body))

			mockStore.AssertExpectations(t)
		})
	}

	t.Run("Concurrent Deletion Requests", func(t *testing.T) {
		var wg sync.WaitGroup
		successCount := 0
		failureCount := 0
		mu := sync.Mutex{}

		mockStore.On("getBookById", "concurrent").Return(Book{Id: "concurrent", Title: "Concurrency"}, 0, nil).Once()
		mockStore.On("getBooks").Return([]Book{{Id: "concurrent", Title: "Concurrency"}}, nil).Once()
		mockStore.On("saveBooks", []Book{}).Return().Once()

		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				req := httptest.NewRequest(http.MethodDelete, "/delete?id=concurrent", nil)
				rec := httptest.NewRecorder()
				handleDeleteBookById(rec, req)

				res := rec.Result()
				body, _ := ioutil.ReadAll(res.Body)

				mu.Lock()
				if strings.Contains(string(body), "Book deleted successfully") {
					successCount++
				} else if strings.Contains(string(body), "Book Not found") {
					failureCount++
				}
				mu.Unlock()
			}()
		}

		wg.Wait()

		require.Equal(t, 1, successCount, "Only one request should succeed")
		require.Equal(t, 4, failureCount, "Remaining requests should return 'Book Not found'")
	})
}


/*
ROOST_METHOD_HASH=handleGetBookById_0c3df003d4
ROOST_METHOD_SIG_HASH=handleGetBookById_8376f0df3c

FUNCTION_DEF=func handleGetBookById(w http.ResponseWriter, r *http.Request) // get book by id handler


*/
func TestHandleGetBookById(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockReturnBook Book
		mockReturnErr  error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Successfully Retrieve a Book by ID",
			queryParams:    "id=valid-id",
			mockReturnBook: Book{Id: "valid-id", Title: "Test Book", Author: "Test Author", Price: "10.99", Imageurl: "test-url"},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":"valid-id","title":"Test Book","author":"Test Author","price":"10.99","image_url":"test-url"}`,
		},
		{
			name:           "Book Not Found for Given ID",
			queryParams:    "id=missing-id",
			mockReturnBook: Book{},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"Msg":"Book Not found"}`,
		},
		{
			name:           "Missing Book ID in Request",
			queryParams:    "",
			mockReturnBook: Book{},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"Msg":"Book Not found"}`,
		},
		{
			name:           "Internal Server Error from getBookById",
			queryParams:    "id=error-id",
			mockReturnBook: Book{},
			mockReturnErr:  errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"Msg":"Internal server error"}`,
		},
		{
			name:           "Malformed JSON Response Handling",
			queryParams:    "id=valid-id",
			mockReturnBook: Book{Id: "valid-id", Title: "Test Book", Author: "Test Author", Price: "10.99", Imageurl: "test-url"},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":"valid-id","title":"Test Book","author":"Test Author","price":"10.99","image_url":"test-url"}`,
		},
		{
			name:           "Case Sensitivity in Query Parameters",
			queryParams:    "ID=valid-id",
			mockReturnBook: Book{},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"Msg":"Book Not found"}`,
		},
		{
			name:           "Multiple Query Parameters in Request",
			queryParams:    "id=valid-id&extra_param=123",
			mockReturnBook: Book{Id: "valid-id", Title: "Test Book", Author: "Test Author", Price: "10.99", Imageurl: "test-url"},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":"valid-id","title":"Test Book","author":"Test Author","price":"10.99","image_url":"test-url"}`,
		},
		{
			name:           "Handling Special Characters in Book ID",
			queryParams:    "id=book-123%3Ftest",
			mockReturnBook: Book{Id: "book-123?test", Title: "Special Book", Author: "Special Author", Price: "15.99", Imageurl: "special-url"},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":"book-123?test","title":"Special Book","author":"Special Author","price":"15.99","image_url":"special-url"}`,
		},
		{
			name:           "Large Book ID Input",
			queryParams:    "id=" + strings.Repeat("a", 1000),
			mockReturnBook: Book{},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"Msg":"Book Not found"}`,
		},
		{
			name: "Concurrent Requests to the Handler",
			mockReturnBook: Book{
				Id:       "valid-id",
				Title:    "Concurrent Book",
				Author:   "Concurrent Author",
				Price:    "12.99",
				Imageurl: "concurrent-url",
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":"valid-id","title":"Concurrent Book","author":"Concurrent Author","price":"12.99","image_url":"concurrent-url"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			req := httptest.NewRequest("GET", "/books?"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handleGetBookById(w, req)

			resp := w.Result()
			body, _ := ioutil.ReadAll(resp.Body)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if strings.TrimSpace(string(body)) != tt.expectedBody {
				t.Errorf("expected body %s, got %s", tt.expectedBody, string(body))
			}
		})
	}

	t.Run("Concurrent Requests to the Handler", func(t *testing.T) {
		var wg sync.WaitGroup
		const numRequests = 10
		wg.Add(numRequests)

		for i := 0; i < numRequests; i++ {
			go func(i int) {
				defer wg.Done()
				req := httptest.NewRequest("GET", fmt.Sprintf("/books?id=valid-id-%d", i), nil)
				w := httptest.NewRecorder()

				handleGetBookById(w, req)

				resp := w.Result()
				body, _ := ioutil.ReadAll(resp.Body)

				expectedBody := `{"id":"valid-id","title":"Concurrent Book","author":"Concurrent Author","price":"12.99","image_url":"concurrent-url"}`
				if strings.TrimSpace(string(body)) != expectedBody {
					t.Errorf("expected body %s, got %s", expectedBody, string(body))
				}
			}(i)
		}

		wg.Wait()
	})
}

func captureOutput(f func()) string {
	var buf bytes.Buffer
	stdout := ioutil.NopCloser(bytes.NewBuffer(nil))
	fmt.Fprintf(stdout, "%s", &buf)
	f()
	return buf.String()
}

func jsonMessageByte(msg string) []byte {
	response, _ := json.Marshal(Message{Msg: msg})
	return response
}


/*
ROOST_METHOD_HASH=handleGetBooks_d7ed706f3d
ROOST_METHOD_SIG_HASH=handleGetBooks_6511204350

FUNCTION_DEF=func handleGetBooks(w http.ResponseWriter, r *http.Request) // List all the books handler


*/
func TestHandleGetBooks(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   func() ([]Book, error)
		expectedStatus int
		expectedBody   string
		expectLog      bool
	}{
		{
			name: "Successfully Retrieve a List of Books",
			mockResponse: func() ([]Book, error) {
				return []Book{
					{Id: "1", Title: "Go Programming", Author: "John Doe", Price: "29.99", Imageurl: "image1.jpg"},
					{Id: "2", Title: "Advanced Go", Author: "Jane Smith", Price: "39.99", Imageurl: "image2.jpg"},
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":"1","title":"Go Programming","author":"John Doe","price":"29.99","image_url":"image1.jpg"},{"id":"2","title":"Advanced Go","author":"Jane Smith","price":"39.99","image_url":"image2.jpg"}]`,
			expectLog:      false,
		},
		{
			name: "Handle Error When getBooks Fails",
			mockResponse: func() ([]Book, error) {
				return nil, fmt.Errorf("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"Msg":"Internal server error"}`,
			expectLog:      true,
		},
		{
			name: "Handle Empty Book List",
			mockResponse: func() ([]Book, error) {
				return []Book{}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[]`,
			expectLog:      false,
		},
	}

	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer log.SetOutput(os.Stdout)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			mockGetBooks = tt.mockResponse

			req := httptest.NewRequest("GET", "/books", nil)
			rr := httptest.NewRecorder()

			handleGetBooks(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			body, _ := ioutil.ReadAll(rr.Body)
			if string(body) != tt.expectedBody {
				t.Errorf("Expected body %s, got %s", tt.expectedBody, string(body))
			}

			loggedOutput := logBuffer.String()
			if tt.expectLog && loggedOutput == "" {
				t.Errorf("Expected error log but got none")
			} else if !tt.expectLog && loggedOutput != "" {
				t.Errorf("Unexpected log output: %s", loggedOutput)
			}

			if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", contentType)
			}
		})
	}

	t.Run("Test Concurrent Requests", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
				t.Fail()
			}
		}()

		mockGetBooks = func() ([]Book, error) {
			return []Book{
				{Id: "1", Title: "Go Programming", Author: "John Doe", Price: "29.99", Imageurl: "image1.jpg"},
			}, nil
		}

		var wg sync.WaitGroup
		numRequests := 10
		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				req := httptest.NewRequest("GET", "/books", nil)
				rr := httptest.NewRecorder()
				handleGetBooks(rr, req)

				if rr.Code != http.StatusOK {
					t.Errorf("Unexpected status code: %d", rr.Code)
				}
			}()
		}
		wg.Wait()
	})
}

func getBooks() ([]Book, error) {
	return mockGetBooks()
}


/*
ROOST_METHOD_HASH=handleUpdateBook_998bf9ccd9
ROOST_METHOD_SIG_HASH=handleUpdateBook_d4f51d5735

FUNCTION_DEF=func handleUpdateBook(w http.ResponseWriter, r *http.Request) 

*/
func TestHandleUpdateBook(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		requestBody    string
		mockGetBook    func(id string) (Book, bool, error)
		mockSaveBooks  func(books []Book) error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Handle Unsupported HTTP Methods",
			method:         http.MethodGet,
			requestBody:    "",
			expectedStatus: 405,
			expectedBody:   "GET - Method not allowed",
		},
		{
			name:           "Handle Malformed JSON in Request Body",
			method:         http.MethodPost,
			requestBody:    "{invalid-json}",
			expectedStatus: 400,
			expectedBody:   "Bad Request",
		},
		{
			name:        "Handle Non-Existent Book ID",
			method:      http.MethodPost,
			requestBody: `{"id": "999", "title": "New Title", "author": "New Author", "price": "15.99", "image_url": "new_image.jpg"}`,
			mockGetBook: func(id string) (Book, bool, error) {
				return Book{}, false, nil
			},
			expectedStatus: 200,
			expectedBody:   "Book Not found",
		},
		{
			name:        "Successfully Update an Existing Book",
			method:      http.MethodPost,
			requestBody: `{"id": "123", "title": "Updated Title", "author": "Updated Author", "price": "20.99", "image_url": "updated_image.jpg"}`,
			mockGetBook: func(id string) (Book, bool, error) {
				return Book{Id: "123", Title: "Old Title", Author: "Old Author", Price: "10.99", Imageurl: "old_image.jpg"}, true, nil
			},
			mockSaveBooks: func(books []Book) error {
				return nil
			},
			expectedStatus: 200,
			expectedBody:   "Book updated successfully",
		},
		{
			name:        "Handle Internal Server Error During Book Save Operation",
			method:      http.MethodPost,
			requestBody: `{"id": "123", "title": "Updated Title", "author": "Updated Author", "price": "20.99", "image_url": "updated_image.jpg"}`,
			mockGetBook: func(id string) (Book, bool, error) {
				return Book{Id: "123", Title: "Old Title", Author: "Old Author", Price: "10.99", Imageurl: "old_image.jpg"}, true, nil
			},
			mockSaveBooks: func(books []Book) error {
				return assert.AnError
			},
			expectedStatus: 500,
			expectedBody:   "Internal server error",
		},
		{
			name:           "Handle Empty Request Body",
			method:         http.MethodPost,
			requestBody:    "",
			expectedStatus: 400,
			expectedBody:   "Bad Request",
		},
		{
			name:        "Handle Unexpected JSON Fields in Request",
			method:      http.MethodPost,
			requestBody: `{"id": "123", "title": "Updated Title", "author": "Updated Author", "price": "20.99", "image_url": "updated_image.jpg", "extra_field": "extra_value"}`,
			mockGetBook: func(id string) (Book, bool, error) {
				return Book{Id: "123", Title: "Old Title", Author: "Old Author", Price: "10.99", Imageurl: "old_image.jpg"}, true, nil
			},
			mockSaveBooks: func(books []Book) error {
				return nil
			},
			expectedStatus: 200,
			expectedBody:   "Book updated successfully",
		},
		{
			name:           "Handle Missing Required Fields in JSON Payload",
			method:         http.MethodPost,
			requestBody:    `{"title": "Updated Title", "author": "Updated Author", "price": "20.99", "image_url": "updated_image.jpg"}`,
			expectedStatus: 200,
			expectedBody:   "Book Not found",
		},
		{
			name:        "Handle Large Request Body",
			method:      http.MethodPost,
			requestBody: `{"id": "123", "title": "` + strings.Repeat("A", 10000) + `", "author": "Updated Author", "price": "20.99", "image_url": "updated_image.jpg"}`,
			mockGetBook: func(id string) (Book, bool, error) {
				return Book{Id: "123", Title: "Old Title", Author: "Old Author", Price: "10.99", Imageurl: "old_image.jpg"}, true, nil
			},
			mockSaveBooks: func(books []Book) error {
				return nil
			},
			expectedStatus: 200,
			expectedBody:   "Book updated successfully",
		},
		{
			name:        "Handle Special Characters in JSON Payload",
			method:      http.MethodPost,
			requestBody: `{"id": "123", "title": "Title!@#$%^&*()", "author": "Author!@#$%^&*()", "price": "20.99", "image_url": "image!@#$%^&*().jpg"}`,
			mockGetBook: func(id string) (Book, bool, error) {
				return Book{Id: "123", Title: "Old Title", Author: "Old Author", Price: "10.99", Imageurl: "old_image.jpg"}, true, nil
			},
			mockSaveBooks: func(books []Book) error {
				return nil
			},
			expectedStatus: 200,
			expectedBody:   "Book updated successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			req := httptest.NewRequest(tt.method, "/update-book", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			if tt.mockGetBook != nil {
				getBookById = tt.mockGetBook
			}
			if tt.mockSaveBooks != nil {
				saveBooks = tt.mockSaveBooks
			}

			handleUpdateBook(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			body, _ := ioutil.ReadAll(res.Body)

			assert.Equal(t, tt.expectedStatus, res.StatusCode, "Unexpected status code")
			assert.Contains(t, string(body), tt.expectedBody, "Unexpected response body")
		})
	}
}


/*
ROOST_METHOD_HASH=jsonMessageByte_2894d43084
ROOST_METHOD_SIG_HASH=jsonMessageByte_e3e47a5059

FUNCTION_DEF=func jsonMessageByte(msg string) [ // response as json format
]byte 

*/
func TestJsonMessageByte(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedJSON  string
		expectFailure bool
	}{
		{
			name:         "Convert a Simple Message to JSON Byte Array",
			input:        "Hello, World!",
			expectedJSON: `{"Msg":"Hello, World!"}`,
		},
		{
			name:         "Convert an Empty String to JSON Byte Array",
			input:        "",
			expectedJSON: `{"Msg":""}`,
		},
		{
			name:         "Convert a Message with Special Characters to JSON Byte Array",
			input:        `Hello, "GoLang"\nNew Line`,
			expectedJSON: `{"Msg":"Hello, \"GoLang\"\\nNew Line"}`,
		},
		{
			name:         "Convert a Long Message to JSON Byte Array",
			input:        "This is a very long string message that should be correctly serialized into JSON format without truncation or errors.",
			expectedJSON: `{"Msg":"This is a very long string message that should be correctly serialized into JSON format without truncation or errors."}`,
		},
		{
			name:         "Convert a Message with Unicode Characters to JSON Byte Array",
			input:        "こんにちは 🌍",
			expectedJSON: `{"Msg":"こんにちは 🌍"}`,
		},
		{
			name:         "Ensure JSON Output is Valid and Well-Formed",
			input:        "Valid JSON Message",
			expectedJSON: `{"Msg":"Valid JSON Message"}`,
		},
		{
			name:         "Verify JSON Output Does Not Include Unintended Fields",
			input:        "Check for unintended fields",
			expectedJSON: `{"Msg":"Check for unintended fields"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			result := jsonMessageByte(tt.input)

			var actual map[string]interface{}
			if err := json.Unmarshal(result, &actual); err != nil {
				t.Fatalf("Failed to unmarshal JSON output: %v", err)
			}

			expected := map[string]interface{}{}
			if err := json.Unmarshal([]byte(tt.expectedJSON), &expected); err != nil {
				t.Fatalf("Failed to unmarshal expected JSON: %v", err)
			}

			if !bytes.Equal(result, []byte(tt.expectedJSON)) {
				t.Errorf("Test %s failed: expected %s but got %s", tt.name, tt.expectedJSON, string(result))
			} else {
				t.Logf("Test %s passed.", tt.name)
			}

			if len(actual) != len(expected) {
				t.Errorf("Test %s failed: Unexpected fields in the JSON output", tt.name)
			}
		})
	}
}

