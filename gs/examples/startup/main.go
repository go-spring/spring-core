package main

import (
	"github.com/go-spring/spring-core/gs"
	"github.com/go-spring/spring-core/util/syslog"
)

func main() {
	if err := gs.Run(); err != nil {
		syslog.Errorf("app run failed: %s", err.Error())
	}
}
