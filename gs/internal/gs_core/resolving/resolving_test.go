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

package resolving

import (
	"net/http"
	"testing"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/util/assert"
)

type Printer interface {
	Print()
}

type Logger struct {
	Name string
}

func (l *Logger) Print() {}

type ChildBean struct {
	Value int
}

func (b *ChildBean) Dump() {}

type TestBean struct {
	Value int
}

func (b *TestBean) NewChild() *ChildBean {
	return &ChildBean{b.Value}
}

func (b *TestBean) NewChildV2() *ChildBean {
	return &ChildBean{b.Value}
}

func NewBean(objOrCtor interface{}, ctorArgs ...gs.Arg) *gs.BeanDefinition {
	return gs_bean.NewBean(objOrCtor, ctorArgs...).Caller(1)
}

func TestResolving(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		r := New()
		{
			r.GroupRegister(func(p conf.Properties) (beans []*gs.BeanDefinition, err error) {
				loggers, err := p.SubKeys("logger")
				if err != nil {
					return nil, err
				}
				for _, name := range loggers {
					l := gs_bean.NewBean(&Logger{Name: name}).
						Export(gs.As[Printer]()).
						Name(name)
					beans = append(beans, l)
				}
				return
			})
			r.Mock(&Logger{Name: "a"}, gs.BeanSelectorFor[Printer]("a"))
		}
		{
			b := r.Register(NewBean(http.NewServeMux())).
				Condition(gs_cond.OnProperty("Enable.ServeMux"))
			assert.Equal(t, b.BeanRegistration().Name(), "ServeMux")
		}
		{
			b := r.Register(NewBean(http.NewServeMux())).Name("ServeMux-2")
			assert.Equal(t, b.BeanRegistration().Name(), "ServeMux-2")
		}
		{
			b := r.Register(NewBean(&TestBean{Value: 1})).Configuration(gs.Configuration{
				Excludes: []string{"NewChildV2"},
			}).Condition(gs_cond.OnBean[Printer]())
			assert.Equal(t, b.BeanRegistration().Name(), "TestBean")
			r.Mock(&ChildBean{Value: 2}, gs.BeanSelectorFor[*ChildBean]())
		}
		{
			b := r.Register(NewBean(&TestBean{Value: 1})).Configuration().Name("TestBean-2")
			r.Mock(&TestBean{Value: 2}, b)
		}
		p := conf.Map(map[string]interface{}{
			"logger": map[string]string{
				"a": "",
				"b": "",
			},
		})
		err := r.Refresh(p)
		assert.Nil(t, err)
		assert.Equal(t, len(r.Beans()), 6)
	})
}
