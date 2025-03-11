package bookstore

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
)








/*
ROOST_METHOD_HASH=getBookById_f77709c63b
ROOST_METHOD_SIG_HASH=getBookById_bbc495e91c

FUNCTION_DEF=func getBookById(id string) (Book, int, error) 

*/
func TestGetBookById(t *testing.T) {
	tests := []struct {
		name          string
		mockBooks     []Book
		mockError     error
		searchID      string
		expectedBook  Book
		expectedIndex int
		expectedError error
	}{
		{
			name: "Retrieve a Book by Valid ID",
			mockBooks: []Book{
				{Id: "1", Title: "Book One", Author: "Author A", Price: "10", Imageurl: "url1"},
				{Id: "2", Title: "Book Two", Author: "Author B", Price: "15", Imageurl: "url2"},
			},
			searchID:      "2",
			expectedBook:  Book{Id: "2", Title: "Book Two", Author: "Author B", Price: "15", Imageurl: "url2"},
			expectedIndex: 1,
			expectedError: nil,
		},
		{
			name: "Retrieve a Book with a Non-Existent ID",
			mockBooks: []Book{
				{Id: "1", Title: "Book One", Author: "Author A", Price: "10", Imageurl: "url1"},
			},
			searchID:      "99",
			expectedBook:  Book{},
			expectedIndex: 0,
			expectedError: nil,
		},
		{
			name:          "Handle an Error from getBooks()",
			mockBooks:     nil,
			mockError:     errors.New("database error"),
			searchID:      "1",
			expectedBook:  Book{},
			expectedIndex: 0,
			expectedError: errors.New("database error"),
		},
		{
			name: "Retrieve the First Book in the List",
			mockBooks: []Book{
				{Id: "100", Title: "First Book", Author: "Author X", Price: "20", Imageurl: "urlX"},
				{Id: "200", Title: "Second Book", Author: "Author Y", Price: "25", Imageurl: "urlY"},
			},
			searchID:      "100",
			expectedBook:  Book{Id: "100", Title: "First Book", Author: "Author X", Price: "20", Imageurl: "urlX"},
			expectedIndex: 0,
			expectedError: nil,
		},
		{
			name: "Retrieve the Last Book in the List",
			mockBooks: []Book{
				{Id: "300", Title: "First Book", Author: "Author X", Price: "20", Imageurl: "urlX"},
				{Id: "400", Title: "Last Book", Author: "Author Y", Price: "30", Imageurl: "urlY"},
			},
			searchID:      "400",
			expectedBook:  Book{Id: "400", Title: "Last Book", Author: "Author Y", Price: "30", Imageurl: "urlY"},
			expectedIndex: 1,
			expectedError: nil,
		},
		{
			name: "Retrieve a Book When Multiple Books Have the Same ID",
			mockBooks: []Book{
				{Id: "500", Title: "First Occurrence", Author: "Author A", Price: "50", Imageurl: "urlA"},
				{Id: "500", Title: "Second Occurrence", Author: "Author B", Price: "55", Imageurl: "urlB"},
			},
			searchID:      "500",
			expectedBook:  Book{Id: "500", Title: "Second Occurrence", Author: "Author B", Price: "55", Imageurl: "urlB"},
			expectedIndex: 1,
			expectedError: nil,
		},
		{
			name:          "Retrieve a Book When List is Empty",
			mockBooks:     []Book{},
			searchID:      "1",
			expectedBook:  Book{},
			expectedIndex: 0,
			expectedError: nil,
		},
		{
			name: "Retrieve a Book with an ID Containing Special Characters",
			mockBooks: []Book{
				{Id: "!@#123$", Title: "Special ID Book", Author: "Author S", Price: "60", Imageurl: "urlS"},
			},
			searchID:      "!@#123$",
			expectedBook:  Book{Id: "!@#123$", Title: "Special ID Book", Author: "Author S", Price: "60", Imageurl: "urlS"},
			expectedIndex: 0,
			expectedError: nil,
		},
		{
			name: "Retrieve a Book with a Numeric ID",
			mockBooks: []Book{
				{Id: "123456", Title: "Numeric ID Book", Author: "Author N", Price: "70", Imageurl: "urlN"},
			},
			searchID:      "123456",
			expectedBook:  Book{Id: "123456", Title: "Numeric ID Book", Author: "Author N", Price: "70", Imageurl: "urlN"},
			expectedIndex: 0,
			expectedError: nil,
		},
		{
			name: "Retrieve a Book with a Long ID String",
			mockBooks: []Book{
				{Id: strings.Repeat("A", 100), Title: "Long ID Book", Author: "Author L", Price: "80", Imageurl: "urlL"},
			},
			searchID:      strings.Repeat("A", 100),
			expectedBook:  Book{Id: strings.Repeat("A", 100), Title: "Long ID Book", Author: "Author L", Price: "80", Imageurl: "urlL"},
			expectedIndex: 0,
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGetBooks = func() ([]Book, error) {
				return tt.mockBooks, tt.mockError
			}

			book, index, err := getBookById(tt.searchID)

			if err != nil && tt.expectedError == nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if err == nil && tt.expectedError != nil {
				t.Fatalf("Expected error but got none")
			}
			if err != nil && tt.expectedError != nil && err.Error() != tt.expectedError.Error() {
				t.Fatalf("Expected error %v but got %v", tt.expectedError, err)
			}
			if book != tt.expectedBook {
				t.Errorf("Expected book %+v but got %+v", tt.expectedBook, book)
			}
			if index != tt.expectedIndex {
				t.Errorf("Expected index %d but got %d", tt.expectedIndex, index)
			}
		})
	}
}

