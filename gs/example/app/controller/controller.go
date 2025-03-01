package controller

import (
	"github.com/go-spring/spring-core/gs"
)

func init() {
	gs.Object(&Controller{})
}

type Controller struct {
	BookController
}
