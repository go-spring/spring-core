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
