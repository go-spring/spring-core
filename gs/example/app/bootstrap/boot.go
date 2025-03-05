package bootstrap

import (
	"os"

	"github.com/go-spring/spring-core/gs"
)

func init() {
	gs.Bootstrap().Object(&Runner{}).AsRunner()
}

type Runner struct{}

func (r *Runner) Run() error {
	err := os.MkdirAll("./conf", os.ModePerm)
	if err != nil {
		return err
	}

	const data = `
server.addr=0.0.0.0:9090

log.biz.name=biz.log
log.biz.dir=./log

log.dao.name=dao.log
log.dao.dir=./log`

	const file = "conf/application.properties"
	return os.WriteFile(file, []byte(data), os.ModePerm)
}
