package book_service

import (
	"testing"

	"github.com/go-spring/spring-core/gs"
	"github.com/go-spring/spring-core/gs/example/dao/book_dao"
	"github.com/go-spring/spring-core/gs/gstest"
	"github.com/go-spring/spring-core/util/assert"
)

func init() {
	gstest.MockFor[*book_dao.BookDao]().With(&book_dao.BookDao{
		Store: map[string]book_dao.Book{
			"0": {SN: "0", Name: "Go Programing"},
		},
	})
	gs.Config().LocalFile.AddDir("../../../conf")
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

	assert.Equal(t, x.SvrAddr, "0.0.0.0:9090")

	s, o := x.Service, x.BookDao
	assert.NotNil(t, o)

	books, err := s.ListBooks()
	assert.Nil(t, err)
	assert.Equal(t, len(books), 1)

	err = s.SaveBook(book_dao.Book{SN: "1", Name: "Go Spring"})
	assert.Nil(t, err)

	books, err = s.ListBooks()
	assert.Nil(t, err)
	assert.Equal(t, len(books), 2)
	assert.Equal(t, books[1].SN, "1")
	assert.Equal(t, books[1].Name, "Go Spring")

	book, err := s.GetBook("1")
	assert.Nil(t, err)
	assert.Equal(t, book.SN, "1")
	assert.Equal(t, book.Name, "Go Spring")

	err = s.DeleteBook("1")
	assert.Nil(t, err)

	books, err = s.ListBooks()
	assert.Nil(t, err)
	assert.Equal(t, len(books), 1)
}
