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

package controller

import (
	"encoding/json"
	"net/http"

	"github.com/go-spring/spring-core/gs/examples/bookman/biz/service/book_service"
	"github.com/go-spring/spring-core/gs/examples/bookman/dao/book_dao"
)

type BookController struct {
	BookService *book_service.BookService `autowire:""`
}

func (c *BookController) ListBooks(w http.ResponseWriter, r *http.Request) {
	books, err := c.BookService.ListBooks()
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_ = json.NewEncoder(w).Encode(books)
}

func (c *BookController) GetBook(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	book, err := c.BookService.GetBook(id)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_ = json.NewEncoder(w).Encode(book)
}

func (c *BookController) SaveBook(w http.ResponseWriter, r *http.Request) {
	var book book_dao.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	if err := c.BookService.SaveBook(book); err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_ = json.NewEncoder(w).Encode("OK!")
}

func (c *BookController) DeleteBook(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := c.BookService.DeleteBook(id)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_ = json.NewEncoder(w).Encode("OK!")
}
