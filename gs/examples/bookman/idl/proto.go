package idl

import (
	"net/http"
)

type Controller interface {
	ListBooks(w http.ResponseWriter, r *http.Request)
	GetBook(w http.ResponseWriter, r *http.Request)
	SaveBook(w http.ResponseWriter, r *http.Request)
	DeleteBook(w http.ResponseWriter, r *http.Request)
}

func RegisterRouter(mux *http.ServeMux, c Controller, wrap func(next http.Handler) http.Handler) {
	mux.Handle("GET /books", wrap(http.HandlerFunc(c.ListBooks)))
	mux.Handle("GET /books/{id}", wrap(http.HandlerFunc(c.GetBook)))
	mux.Handle("POST /books", wrap(http.HandlerFunc(c.SaveBook)))
	mux.Handle("DELETE /books/{id}", wrap(http.HandlerFunc(c.DeleteBook)))
}
