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

func RegisterRouter(mux *http.ServeMux, c Controller) {
	mux.HandleFunc("GET /books", c.ListBooks)
	mux.HandleFunc("GET /books/{id}", c.GetBook)
	mux.HandleFunc("POST /books", c.SaveBook)
	mux.HandleFunc("DELETE /books/{id}", c.DeleteBook)
}
