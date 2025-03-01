package book_service

import (
	"testing"

	"github.com/go-spring/spring-core/gs/example/dao/book_dao"
	"github.com/go-spring/spring-core/gs/gstest"
	"github.com/go-spring/spring-core/util/assert"
)

func init() {
	gstest.MockFor[*book_dao.BookDao]().With(nil)
}

func TestMain(m *testing.M) {
	gstest.Run(m)
}

func TestBookService(t *testing.T) {
	gstest.Case(t, func(s *BookService, o *book_dao.BookDao) {
		assert.NotNil(t, o)

		books, err := s.ListBooks()
		assert.Nil(t, err)
		assert.Equal(t, len(books), 0)

		err = s.SaveBook(book_dao.Book{SN: "1", Name: "Go Spring"})
		assert.Nil(t, err)

		books, err = s.ListBooks()
		assert.Nil(t, err)
		assert.Equal(t, len(books), 1)
		assert.Equal(t, books[0].SN, "1")
		assert.Equal(t, books[0].Name, "Go Spring")

		book, err := s.GetBook("1")
		assert.Nil(t, err)
		assert.Equal(t, book.SN, "1")
		assert.Equal(t, book.Name, "Go Spring")

		err = s.DeleteBook("1")
		assert.Nil(t, err)

		books, err = s.ListBooks()
		assert.Nil(t, err)
		assert.Equal(t, len(books), 0)
	})
}
