package log

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-spring/spring-core/gs"
)

func init() {
	gs.GroupRegister(func(p gs.Properties) ([]*gs.BeanDefinition, error) {
		var loggers map[string]struct {
			Name string `value:"${name}"`
			Dir  string `value:"${dir}"`
		}
		err := p.Bind(&loggers, "${log}")
		if err != nil {
			return nil, err
		}
		var ret []*gs.BeanDefinition
		for k, l := range loggers {
			var (
				f    *os.File
				flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
			)
			f, err = os.OpenFile(filepath.Join(l.Dir, l.Name), flag, os.ModePerm)
			if err != nil {
				return nil, err
			}
			o := slog.New(slog.NewTextHandler(f, nil))
			b := gs.NewBean(o).Name(k).Destroy(func(_ *slog.Logger) {
				_ = f.Close()
			})
			ret = append(ret, b)
		}
		return ret, nil
	})
}
