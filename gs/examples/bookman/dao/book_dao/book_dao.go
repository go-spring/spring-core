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
	"log/slog"
	"maps"
	"slices"

	"github.com/go-spring/spring-core/gs"
)

func init() {
	gs.Object(&BookDao{Store: map[string]Book{}})
}

type Book struct {
	SN     string `json:"sn"`
	Name   string `json:"name"`
	Author string `json:"author"`
}

type BookDao struct {
	Store  map[string]Book
	Logger *slog.Logger `autowire:"dao"`
}

func (dao *BookDao) ListBooks() ([]Book, error) {
	r := slices.Collect(maps.Values(dao.Store))
	return r, nil
}

func (dao *BookDao) GetBook(sn string) (Book, error) {
	r, _ := dao.Store[sn]
	return r, nil
}

func (dao *BookDao) SaveBook(book Book) error {
	dao.Store[book.SN] = book
	return nil
}

func (dao *BookDao) DeleteBook(sn string) error {
	delete(dao.Store, sn)
	return nil
}
