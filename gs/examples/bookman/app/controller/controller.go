package controller

import (
	"github.com/go-spring/spring-core/gs"
	"github.com/go-spring/spring-core/gs/examples/bookman/idl"
)

func init() {
	gs.Object(&Controller{}).Export(gs.As[idl.Controller]())
}

var _ idl.Controller = (*Controller)(nil)

type Controller struct {
	BookController
}
