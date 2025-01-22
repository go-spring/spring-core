package syslog_test

import (
	"testing"

	"github.com/go-spring/spring-core/gs/syslog"
)

func TestLog(t *testing.T) {
	syslog.Infof("hello %s", "world")
}
