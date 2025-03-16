/*
 * Copyright 2025 The Go-Spring Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package book_service

import (
	"testing"

	"github.com/go-spring/spring-core/gs"
	"github.com/go-spring/spring-core/gs/examples/bookman/src/dao/book_dao"
	"github.com/go-spring/spring-core/gs/gstest"
	"github.com/go-spring/spring-core/util/assert"
)

func init() {
	// Mock the BookDao with initial test data
	gstest.MockFor[*book_dao.BookDao]().With(&book_dao.BookDao{
		Store: map[string]book_dao.Book{
			"978-0132350884": {
				Title:     "Clean Code",
				Author:    "Robert C. Martin",
				ISBN:      "978-0132350884",
				Publisher: "Prentice Hall",
			},
		},
	})
	// Load local configuration files
	gs.Config().LocalFile.AddDir("../../../../conf")
}

func TestMain(m *testing.M) {
	gstest.TestMain(m)
}

func TestBookService(t *testing.T) {

	x := gstest.Wire(t, new(struct {
		SvrAddr string            `value:"${server.addr}"`
		Service *BookService      `autowire:""`
		BookDao *book_dao.BookDao `autowire:""`
	}))

	// Verify server address configuration
	assert.Equal(t, x.SvrAddr, "0.0.0.0:9090")

	s, o := x.Service, x.BookDao
	assert.NotNil(t, o)

	// Test listing books
	books, err := s.ListBooks()
	assert.Nil(t, err)
	assert.Equal(t, len(books), 1)
	assert.Equal(t, books[0].ISBN, "978-0132350884")

	// Test saving a new book
	err = s.SaveBook(book_dao.Book{
		Title:     "Introduction to Algorithms",
		Author:    "Thomas H. Cormen, Charles E. Leiserson, ...",
		ISBN:      "978-0262033848",
		Publisher: "MIT Press",
	})
	assert.Nil(t, err)

	// Verify book was added successfully
	books, err = s.ListBooks()
	assert.Nil(t, err)
	assert.Equal(t, len(books), 2)
	assert.Equal(t, books[1].ISBN, "978-0262033848")
	assert.Equal(t, books[1].Title, "Introduction to Algorithms")

	// Test retrieving a book by ISBN
	book, err := s.GetBook("978-0132350884")
	assert.Nil(t, err)
	assert.Equal(t, book.ISBN, "978-0132350884")
	assert.Equal(t, book.Title, "Clean Code")

	// Test deleting a book
	err = s.DeleteBook("978-0132350884")
	assert.Nil(t, err)

	// Verify book deletion
	books, err = s.ListBooks()
	assert.Nil(t, err)
	assert.Equal(t, len(books), 1)
}
