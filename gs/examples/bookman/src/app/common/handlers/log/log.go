/*
 * Copyright 2025 The Go-Spring Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package log

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs"
)

func init() {
	gs.GroupRegister(func(p conf.Properties) ([]*gs.BeanDefinition, error) {
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
