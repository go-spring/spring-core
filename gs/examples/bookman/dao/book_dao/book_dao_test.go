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
