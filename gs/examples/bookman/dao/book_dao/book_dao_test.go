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

package book_dao

import (
	"os"
	"testing"

	"github.com/go-spring/spring-core/gs/gstest"
	"github.com/go-spring/spring-core/util/assert"
)

func init() {
	_ = os.Setenv("GS_SPRING_APP_CONFIG_DIR", "../../conf")
}

func TestMain(m *testing.M) {
	gstest.TestMain(m)
}

func TestBookDao(t *testing.T) {

	// Wire dependencies and retrieve the server address
	x := gstest.Wire(t, &struct {
		SvrAddr string `value:"${server.addr}"`
	}{})
	assert.Equal(t, x.SvrAddr, "0.0.0.0:9090")

	// Retrieve BookDao instance
	o := gstest.Get[*BookDao](t)
	assert.NotNil(t, o)

	// Test listing books
	books, err := o.ListBooks()
	assert.Nil(t, err)
	assert.Equal(t, len(books), 1)
	assert.Equal(t, books[0].ISBN, "978-0134190440")
	assert.Equal(t, books[0].Title, "The Go Programming Language")

	// Test saving a new book
	err = o.SaveBook(Book{
		Title:     "Clean Code",
		Author:    "Robert C. Martin",
		ISBN:      "978-0132350884",
		Publisher: "Prentice Hall",
	})
	assert.Equal(t, err, nil)

	// Verify book was added
	books, err = o.ListBooks()
	assert.Nil(t, err)
	assert.Equal(t, len(books), 2)
	assert.Equal(t, books[0].ISBN, "978-0132350884")
	assert.Equal(t, books[0].Title, "Clean Code")

	// Test retrieving a book by ISBN
	book, err := o.GetBook("978-0132350884")
	assert.Nil(t, err)
	assert.Equal(t, book.Title, "Clean Code")
	assert.Equal(t, book.Publisher, "Prentice Hall")

	// Test deleting a book
	err = o.DeleteBook("978-0132350884")
	assert.Nil(t, err)

	// Verify book was deleted
	books, err = o.ListBooks()
	assert.Nil(t, err)
	assert.Equal(t, len(books), 1)
	assert.Equal(t, books[0].ISBN, "978-0134190440")
}
