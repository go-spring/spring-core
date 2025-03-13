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

	x := gstest.Wire(t, &struct {
		SvrAddr string `value:"${server.addr}"`
	}{})
	assert.Equal(t, x.SvrAddr, "0.0.0.0:9090")

	o := gstest.Get[*BookDao](t)
	assert.NotNil(t, o)

	books, err := o.ListBooks()
	assert.Nil(t, err)
	assert.Equal(t, len(books), 0)

	err = o.SaveBook(Book{SN: "1", Name: "Go Spring"})
	assert.Equal(t, err, nil)

	books, err = o.ListBooks()
	assert.Nil(t, err)
	assert.Equal(t, len(books), 1)
	assert.Equal(t, books[0].SN, "1")
	assert.Equal(t, books[0].Name, "Go Spring")

	book, err := o.GetBook("1")
	assert.Nil(t, err)
	assert.Equal(t, book.SN, "1")
	assert.Equal(t, book.Name, "Go Spring")

	err = o.DeleteBook("1")
	assert.Nil(t, err)

	books, err = o.ListBooks()
	assert.Nil(t, err)
	assert.Equal(t, len(books), 0)
}
