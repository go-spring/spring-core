package book_sdk

import (
	"github.com/go-spring/spring-core/gs"
)

func init() {
	gs.Object(&BookSDK{})
}

type BookSDK struct{}

func (s *BookSDK) GetPrice(isbn string) string {
	return "￥10"
}
