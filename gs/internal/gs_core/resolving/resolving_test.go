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
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/util/assert"
)

type Logger interface {
	Print(msg string)
}

type CtxLogger interface {
	CtxPrint(ctx context.Context, msg string)
}

type MyLogger struct {
	Name string
}

func NewLogger(name string) *MyLogger {
	return &MyLogger{Name: name}
}

func (l *MyLogger) Print(msg string) {}

func (l *MyLogger) CtxPrint(ctx context.Context, msg string) {}

type ChildBean struct {
	Value int
}

func (b *ChildBean) Echo() {}

type TestBean struct {
	Value int
}

func (b *TestBean) NewChild() *ChildBean {
	return &ChildBean{b.Value}
}

func (b *TestBean) NewChildV2() *ChildBean {
	return &ChildBean{b.Value}
}

func TestResolving(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		r := New()
		{
			r.GroupRegister(func(p conf.Properties) (beans []*gs.BeanDefinition, err error) {
				keys, err := p.SubKeys("logger")
				if err != nil {
					return nil, err
				}
				for _, name := range keys {
					arg := gs_arg.Tag(fmt.Sprintf("logger.%s", name))
					l := gs_bean.NewBean(NewLogger, arg).
						Export(gs.As[Logger](), gs.As[CtxLogger]()).
						Name(name)
					beans = append(beans, l)
				}
				return
			})
			r.Provide(NewLogger, gs_arg.Value("c")).Name("c")
			r.Mock(NewLogger("a@mocked"), gs.BeanSelectorFor[Logger]("a"))
			r.Mock(NewLogger("b@mocked"), gs.BeanSelectorFor[CtxLogger]("b"))
		}
		{
			b := r.Provide(http.NewServeMux).
				Condition(gs_cond.OnProperty("Enable.ServeMux"))
			assert.Equal(t, b.BeanRegistration().Name(), "NewServeMux")
		}
		{
			b := r.Provide(http.NewServeMux).Name("ServeMux")
			assert.Equal(t, b.BeanRegistration().Name(), "ServeMux")
		}
		{
			b := r.Object(&TestBean{Value: 1}).Configuration().Name("TestBean")
			assert.Equal(t, b.BeanRegistration().Name(), "TestBean")
			r.Mock(&TestBean{Value: 2}, b)
		}
		{
			b := r.Object(&TestBean{Value: 1}).Name("TestBean-2").
				Configuration(gs.Configuration{
					Excludes: []string{"NewChildV2"},
				})
			assert.Equal(t, b.BeanRegistration().Name(), "TestBean-2")
			r.Mock(&ChildBean{Value: 2}, gs.BeanSelectorFor[*ChildBean]())
		}

		p := conf.Map(map[string]interface{}{
			"logger": map[string]string{
				"a": "",
				"b": "",
			},
		})
		err := r.Refresh(p)
		assert.Nil(t, err)
		assert.Equal(t, len(r.Beans()), 7)
	})
}
