package book_dao

import (
	"maps"
	"slices"

	"github.com/go-spring/spring-core/gs"
)

func init() {
	gs.Object(&BookDao{store: map[string]Book{}})
}

type Book struct {
	SN     string `json:"sn"`
	Name   string `json:"name"`
	Author string `json:"author"`
}

type BookDao struct {
	store map[string]Book
}

func (dao *BookDao) ListBooks() ([]Book, error) {
	r := slices.Collect(maps.Values(dao.store))
	return r, nil
}

func (dao *BookDao) GetBook(sn string) (Book, error) {
	r, _ := dao.store[sn]
	return r, nil
}

func (dao *BookDao) SaveBook(book Book) error {
	dao.store[book.SN] = book
	return nil
}

func (dao *BookDao) DeleteBook(sn string) error {
	delete(dao.store, sn)
	return nil
}
