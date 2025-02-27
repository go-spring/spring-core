package book_service

import (
	"github.com/go-spring/spring-core/gs"
	"github.com/go-spring/spring-core/gs/testcase/dao/book_dao"
)

func init() {
	gs.Object(&BookService{})
}

type BookService struct {
	BookDao *book_dao.BookDao `autowire:""`
}

func (s *BookService) ListBooks() ([]book_dao.Book, error) {
	return s.BookDao.ListBooks()
}

func (s *BookService) GetBook(sn string) (book_dao.Book, error) {
	return s.BookDao.GetBook(sn)
}

func (s *BookService) SaveBook(book book_dao.Book) error {
	return s.BookDao.SaveBook(book)
}

func (s *BookService) DeleteBook(sn string) error {
	return s.BookDao.DeleteBook(sn)
}
