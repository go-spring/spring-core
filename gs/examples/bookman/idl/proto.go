package idl

import (
	"net/http"

	"github.com/go-spring/spring-core/gs"
)

func init() {
	gs.Provide(func(c Controller) *http.ServeMux {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /books", c.ListBooks)
		mux.HandleFunc("GET /books/{id}", c.GetBook)
		mux.HandleFunc("POST /books", c.SaveBook)
		mux.HandleFunc("DELETE /books/{id}", c.DeleteBook)
		return mux
	})
}

type Controller interface {
	ListBooks(w http.ResponseWriter, r *http.Request)
	GetBook(w http.ResponseWriter, r *http.Request)
	SaveBook(w http.ResponseWriter, r *http.Request)
	DeleteBook(w http.ResponseWriter, r *http.Request)
}
