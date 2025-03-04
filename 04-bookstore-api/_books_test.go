package bookstore

import (
	errors "errors"
	debug "runtime/debug"
	testing "testing"
	assert "github.com/stretchr/testify/assert"
	json "encoding/json"
	ioutil "io/ioutil"
	os "os"
	fmt "fmt"
	sync "sync"
)



var mockGetBooks func() ([]Book, error)




/*
ROOST_METHOD_HASH=getBookById_f77709c63b
ROOST_METHOD_SIG_HASH=getBookById_bbc495e91c

FUNCTION_DEF=func getBookById(id string) (Book, int, error) 

*/
func TestGetBookById(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  []Book
		mockError     error
		searchID      string
		expectedBook  Book
		expectedIndex int
		expectError   bool
	}{
		{
			name: "Successfully Retrieve a Book by Valid ID",
			mockResponse: []Book{
				{Id: "1", Title: "Book One", Author: "Author A", Price: "10", Imageurl: "url1"},
				{Id: "2", Title: "Book Two", Author: "Author B", Price: "20", Imageurl: "url2"},
			},
			searchID:      "1",
			expectedBook:  Book{Id: "1", Title: "Book One", Author: "Author A", Price: "10", Imageurl: "url1"},
			expectedIndex: 0,
			expectError:   false,
		},
		{
			name:          "Book Not Found for an Invalid ID",
			mockResponse:  []Book{},
			searchID:      "99",
			expectedBook:  Book{},
			expectedIndex: 0,
			expectError:   false,
		},
		{
			name:          "getBooks Function Returns an Error",
			mockError:     errors.New("database failure"),
			searchID:      "1",
			expectedBook:  Book{},
			expectedIndex: 0,
			expectError:   true,
		},
		{
			name: "Multiple Books with the Same ID",
			mockResponse: []Book{
				{Id: "1", Title: "Book One", Author: "Author A", Price: "10", Imageurl: "url1"},
				{Id: "1", Title: "Duplicate Book", Author: "Author C", Price: "15", Imageurl: "url3"},
			},
			searchID:      "1",
			expectedBook:  Book{Id: "1", Title: "Duplicate Book", Author: "Author C", Price: "15", Imageurl: "url3"},
			expectedIndex: 1,
			expectError:   false,
		},
		{
			name:          "Empty Book List",
			mockResponse:  []Book{},
			searchID:      "1",
			expectedBook:  Book{},
			expectedIndex: 0,
			expectError:   false,
		},
		{
			name: "Large Dataset Handling",
			mockResponse: func() []Book {
				books := make([]Book, 10000)
				for i := 0; i < 10000; i++ {
					books[i] = Book{Id: string(i), Title: "Book", Author: "Author", Price: "10", Imageurl: "url"}
				}
				return books
			}(),
			searchID:      "9999",
			expectedBook:  Book{Id: "9999", Title: "Book", Author: "Author", Price: "10", Imageurl: "url"},
			expectedIndex: 9999,
			expectError:   false,
		},
		{
			name: "Book ID Case Sensitivity",
			mockResponse: []Book{
				{Id: "abc", Title: "Case Test", Author: "Author X", Price: "9", Imageurl: "urlX"},
			},
			searchID:      "ABC",
			expectedBook:  Book{},
			expectedIndex: 0,
			expectError:   false,
		},
		{
			name:          "Book ID as Empty String",
			mockResponse:  []Book{{Id: "1", Title: "Book One", Author: "Author A", Price: "10", Imageurl: "url1"}},
			searchID:      "",
			expectedBook:  Book{},
			expectedIndex: 0,
			expectError:   false,
		},
		{
			name: "Book ID with Special Characters",
			mockResponse: []Book{
				{Id: "!@#$%", Title: "Special Char Book", Author: "Author Y", Price: "12", Imageurl: "urlY"},
			},
			searchID:      "!@#$%",
			expectedBook:  Book{Id: "!@#$%", Title: "Special Char Book", Author: "Author Y", Price: "12", Imageurl: "urlY"},
			expectedIndex: 0,
			expectError:   false,
		},
		{
			name: "Last Book in List Retrieval",
			mockResponse: []Book{
				{Id: "1", Title: "Book One", Author: "Author A", Price: "10", Imageurl: "url1"},
				{Id: "2", Title: "Book Two", Author: "Author B", Price: "20", Imageurl: "url2"},
			},
			searchID:      "2",
			expectedBook:  Book{Id: "2", Title: "Book Two", Author: "Author B", Price: "20", Imageurl: "url2"},
			expectedIndex: 1,
			expectError:   false,
		},
		{
			name: "First Book in List Retrieval",
			mockResponse: []Book{
				{Id: "1", Title: "Book One", Author: "Author A", Price: "10", Imageurl: "url1"},
				{Id: "2", Title: "Book Two", Author: "Author B", Price: "20", Imageurl: "url2"},
			},
			searchID:      "1",
			expectedBook:  Book{Id: "1", Title: "Book One", Author: "Author A", Price: "10", Imageurl: "url1"},
			expectedIndex: 0,
			expectError:   false,
		},
		{
			name: "Book ID as Numeric String",
			mockResponse: []Book{
				{Id: "123", Title: "Numeric ID Book", Author: "Author Z", Price: "14", Imageurl: "urlZ"},
			},
			searchID:      "123",
			expectedBook:  Book{Id: "123", Title: "Numeric ID Book", Author: "Author Z", Price: "14", Imageurl: "urlZ"},
			expectedIndex: 0,
			expectError:   false,
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

			mockGetBooks = func() ([]Book, error) {
				if tt.mockError != nil {
					return nil, tt.mockError
				}
				return tt.mockResponse, nil
			}

			book, index, err := getBookById(tt.searchID)

			if tt.expectError {
				assert.Error(t, err, "Expected an error but got none")
			} else {
				assert.NoError(t, err, "Did not expect an error but got one")
				assert.Equal(t, tt.expectedBook, book, "Returned book does not match expected")
				assert.Equal(t, tt.expectedIndex, index, "Returned index does not match expected")
			}

			t.Logf("Test '%s' passed successfully", tt.name)
		})
	}
}

func getBooks() ([]Book, error) {
	return mockGetBooks()
}


/*
ROOST_METHOD_HASH=getBooks_95b80b4f2d
ROOST_METHOD_SIG_HASH=getBooks_f21354163b

FUNCTION_DEF=func getBooks() ([ // Get books - returns books and error
]Book, error) 

*/
func TestGetBooks(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() error
		expectedError bool
		expectedBooks []Book
	}{
		{
			name: "Successfully Retrieve Books from JSON File",
			setup: func() error {
				data := `[{"id": "1", "title": "Book One", "author": "Author One", "price": "10.99", "image_url": "url1"},
				         {"id": "2", "title": "Book Two", "author": "Author Two", "price": "15.99", "image_url": "url2"}]`
				return ioutil.WriteFile("./books.json", []byte(data), 0644)
			},
			expectedError: false,
			expectedBooks: []Book{
				{"1", "Book One", "Author One", "10.99", "url1"},
				{"2", "Book Two", "Author Two", "15.99", "url2"},
			},
		},
		{
			name: "Handle Missing JSON File Gracefully",
			setup: func() error {
				return os.Remove("./books.json")
			},
			expectedError: true,
			expectedBooks: nil,
		},
		{
			name: "Handle Corrupted JSON File",
			setup: func() error {
				data := `{"id": "1", "title": "Book One", "author": "Author One", "price": "10.99", "image_url": "url1"`
				return ioutil.WriteFile("./books.json", []byte(data), 0644)
			},
			expectedError: true,
			expectedBooks: nil,
		},
		{
			name: "Handle Empty JSON File",
			setup: func() error {
				return ioutil.WriteFile("./books.json", []byte(""), 0644)
			},
			expectedError: false,
			expectedBooks: []Book{},
		},
		{
			name: "Handle JSON File with Empty Array",
			setup: func() error {
				return ioutil.WriteFile("./books.json", []byte("[]"), 0644)
			},
			expectedError: false,
			expectedBooks: []Book{},
		},
		{
			name: "Handle JSON File with Partially Valid Data",
			setup: func() error {
				data := `[{"id": "1", "title": "Book One", "author": "Author One", "price": "10.99", "image_url": "url1"},
				         {"id": "2", "title": "Invalid JSON Entry" "author": "Author Two", "price": "15.99", "image_url": "url2"}]`
				return ioutil.WriteFile("./books.json", []byte(data), 0644)
			},
			expectedError: true,
			expectedBooks: nil,
		},
		{
			name: "Handle Large JSON File",
			setup: func() error {
				var books []Book
				for i := 0; i < 10000; i++ {
					books = append(books, Book{
						Id:       fmt.Sprintf("%d", i),
						Title:    fmt.Sprintf("Book %d", i),
						Author:   fmt.Sprintf("Author %d", i),
						Price:    "19.99",
						Imageurl: fmt.Sprintf("url%d", i),
					})
				}
				data, err := json.Marshal(books)
				if err != nil {
					return err
				}
				return ioutil.WriteFile("./books.json", data, 0644)
			},
			expectedError: false,
			expectedBooks: nil,
		},
		{
			name: "Handle Special Characters in JSON File",
			setup: func() error {
				data := `[{"id": "1", "title": "Bøøk \"One\"", "author": "Äuthor One", "price": "10.99", "image_url": "url1"}]`
				return ioutil.WriteFile("./books.json", []byte(data), 0644)
			},
			expectedError: false,
			expectedBooks: []Book{
				{"1", "Bøøk \"One\"", "Äuthor One", "10.99", "url1"},
			},
		},
		{
			name: "Handle Unexpected Data Types in JSON File",
			setup: func() error {
				data := `[{"id": "1", "title": "Book One", "author": "Author One", "price": 10.99, "image_url": "url1"}]`
				return ioutil.WriteFile("./books.json", []byte(data), 0644)
			},
			expectedError: true,
			expectedBooks: nil,
		},
		{
			name: "Handle File Permission Issues",
			setup: func() error {
				data := `[{"id": "1", "title": "Book One", "author": "Author One", "price": "10.99", "image_url": "url1"}]`
				err := ioutil.WriteFile("./books.json", []byte(data), 0000)
				if err != nil {
					return err
				}
				return nil
			},
			expectedError: true,
			expectedBooks: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			if err := tc.setup(); err != nil {
				t.Fatalf("Failed to set up test case: %v", err)
			}

			books, err := getBooks()

			if (err != nil) != tc.expectedError {
				t.Errorf("Expected error: %v, got: %v", tc.expectedError, err)
			}

			if err == nil && tc.expectedBooks != nil {
				if len(books) != len(tc.expectedBooks) {
					t.Errorf("Expected %d books, got %d", len(tc.expectedBooks), len(books))
				}
				for i, book := range books {
					if book != tc.expectedBooks[i] {
						t.Errorf("Book mismatch at index %d: expected %+v, got %+v", i, tc.expectedBooks[i], book)
					}
				}
			}

			_ = os.Remove("./books.json")
		})
	}
}


/*
ROOST_METHOD_HASH=saveBooks_f944094c1b
ROOST_METHOD_SIG_HASH=saveBooks_1fdb6f7496

FUNCTION_DEF=func saveBooks(books [ // save books to books.json file
]Book) error 

*/
func TestSaveBooks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		books          []Book
		mockWriteError bool
		expectError    bool
	}{
		{
			name: "Successfully Save a List of Books",
			books: []Book{
				{Id: "1", Title: "Go Programming", Author: "John Doe", Price: "29.99", Imageurl: "https://example.com/go.jpg"},
				{Id: "2", Title: "Advanced Go", Author: "Jane Doe", Price: "39.99", Imageurl: "https://example.com/advgo.jpg"},
			},
			expectError: false,
		},
		{
			name:        "Save an Empty Book List",
			books:       []Book{},
			expectError: false,
		},
		{
			name: "Handle JSON Serialization Failure",
			books: []Book{
				{Id: "1", Title: "Invalid", Author: "John Doe", Price: "NaN", Imageurl: "https://example.com/go.jpg"},
			},
			expectError: true,
		},
		{
			name:           "Handle File Write Failure",
			books:          []Book{{Id: "1", Title: "Go Programming", Author: "John Doe", Price: "29.99", Imageurl: "https://example.com/go.jpg"}},
			mockWriteError: true,
			expectError:    true,
		},
		{
			name: "Handle Large List of Books",
			books: func() []Book {
				var books []Book
				for i := 0; i < 10000; i++ {
					books = append(books, Book{Id: string(i), Title: "Book " + string(i), Author: "Author " + string(i), Price: "9.99", Imageurl: "https://example.com/book.jpg"})
				}
				return books
			}(),
			expectError: false,
		},
		{
			name: "Handle Special Characters in Book Data",
			books: []Book{
				{Id: "1", Title: "Go & Rust", Author: "John \"Doe\"", Price: "29.99", Imageurl: "https://example.com/go.jpg"},
			},
			expectError: false,
		},
		{
			name: "Handle Non-UTF-8 Characters in Book Data",
			books: []Book{
				{Id: "1", Title: "Golang 🚀", Author: "张伟", Price: "29.99", Imageurl: "https://example.com/go.jpg"},
			},
			expectError: false,
		},
		{
			name: "Validate JSON Structure in Output File",
			books: []Book{
				{Id: "1", Title: "Go Programming", Author: "John Doe", Price: "29.99", Imageurl: "https://example.com/go.jpg"},
			},
			expectError: false,
		},
		{
			name: "Handle Concurrent Access to books.json",
			books: []Book{
				{Id: "1", Title: "Go Programming", Author: "John Doe", Price: "29.99", Imageurl: "https://example.com/go.jpg"},
			},
			expectError: false,
		},
		{
			name: "Handle Missing or Deleted books.json File",
			books: []Book{
				{Id: "1", Title: "Go Programming", Author: "John Doe", Price: "29.99", Imageurl: "https://example.com/go.jpg"},
			},
			expectError: false,
		},
	}

	originalWriteFile := ioutil.WriteFile
	defer func() { ioutil.WriteFile = originalWriteFile }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			if tt.mockWriteError {
				ioutil.WriteFile = func(filename string, data []byte, perm os.FileMode) error {
					return errors.New("mock write error")
				}
			} else {
				ioutil.WriteFile = originalWriteFile
			}

			if tt.name == "Handle Concurrent Access to books.json" {
				var wg sync.WaitGroup
				for i := 0; i < 5; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						err := saveBooks(tt.books)
						if (err != nil) != tt.expectError {
							t.Errorf("Unexpected error: %v", err)
						}
					}()
				}
				wg.Wait()
			} else {
				err := saveBooks(tt.books)
				if (err != nil) != tt.expectError {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if tt.name == "Validate JSON Structure in Output File" {
				data, err := ioutil.ReadFile("./books.json")
				if err != nil {
					t.Errorf("Failed to read books.json: %v", err)
				}

				var storedBooks []Book
				err = json.Unmarshal(data, &storedBooks)
				if err != nil {
					t.Errorf("Invalid JSON structure: %v", err)
				}

				if len(storedBooks) != len(tt.books) {
					t.Errorf("Mismatch in stored books count")
				}
			}

			if tt.name == "Handle Missing or Deleted books.json File" {
				os.Remove("./books.json")
				err := saveBooks(tt.books)
				if err != nil {
					t.Errorf("Failed to create books.json: %v", err)
				}
				if _, err := os.Stat("./books.json"); os.IsNotExist(err) {
					t.Errorf("books.json file was not created")
				}
			}
		})
	}
}

