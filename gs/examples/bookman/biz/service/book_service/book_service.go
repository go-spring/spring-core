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
	"log/slog"

	"github.com/go-spring/spring-core/gs"
	"github.com/go-spring/spring-core/gs/examples/bookman/dao/book_dao"
)

func init() {
	gs.Object(&BookService{})
}

type BookService struct {
	BookDao *book_dao.BookDao `autowire:""`
	Logger  *slog.Logger      `autowire:"biz"`
}

func (s *BookService) ListBooks() ([]book_dao.Book, error) {
	return s.BookDao.ListBooks()
}

func (s *BookService) GetBook(isbn string) (book_dao.Book, error) {
	return s.BookDao.GetBook(isbn)
}

func (s *BookService) SaveBook(book book_dao.Book) error {
	return s.BookDao.SaveBook(book)
}

func (s *BookService) DeleteBook(isbn string) error {
	return s.BookDao.DeleteBook(isbn)
}
