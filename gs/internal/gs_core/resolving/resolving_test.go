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
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-spring/gs-assert/assert"
	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
)

type Logger interface {
	Print(msg string)
}

type SimpleLogger struct{}

func (l *SimpleLogger) Print(msg string) {}

type CtxLogger interface {
	CtxPrint(ctx context.Context, msg string)
}

type ZeroLogger struct {
	Name string
}

func NewLogger(name string) Logger {
	return NewZeroLogger(name)
}

func NewZeroLogger(name string) *ZeroLogger {
	return &ZeroLogger{Name: name}
}

func (l *ZeroLogger) Print(msg string) {}

func (l *ZeroLogger) CtxPrint(ctx context.Context, msg string) {}

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

func (b *TestBean) NewChildV2() (*ChildBean, error) {
	return &ChildBean{b.Value}, nil
}

func (b *TestBean) Echo() {}

func TestResolving(t *testing.T) {

	t.Run("register error", func(t *testing.T) {
		r := New()
		err := r.Refresh(conf.New())
		assert.That(t, err).Nil()
		assert.Panic(t, func() {
			r.Register(&gs.BeanDefinition{})
		}, "container is refreshing or already refreshed")
	})

	t.Run("group error", func(t *testing.T) {
		r := New()
		r.GroupRegister(func(p conf.Properties) ([]*gs.BeanDefinition, error) {
			return nil, fmt.Errorf("group error")
		})
		err := r.Refresh(conf.New())
		assert.ThatError(t, err).Matches("group error")
	})

	t.Run("configuration error - 1", func(t *testing.T) {
		r := New()
		r.Object(&TestBean{Value: 1}).Configuration()
		r.AddMock(gs.BeanMock{
			Object: &TestBean{Value: 9},
			Target: gs.BeanSelectorFor[*TestBean](),
		})
		r.AddMock(gs.BeanMock{
			Object: &TestBean{Value: 9},
			Target: gs.BeanSelectorFor[*TestBean](),
		})
		err := r.Refresh(conf.New())
		assert.ThatError(t, err).Matches("found duplicate mock bean for 'TestBean'")
	})

	t.Run("configuration error - 2", func(t *testing.T) {
		r := New()
		r.Object(&TestBean{Value: 1}).Configuration(
			gs.Configuration{
				Includes: []string{"*"},
			},
		)
		err := r.Refresh(conf.New())
		assert.ThatError(t, err).Matches("error parsing regexp: missing argument to repetition operator: `*`")
	})

	t.Run("configuration error - 3", func(t *testing.T) {
		r := New()
		r.Object(&TestBean{Value: 1}).Configuration(
			gs.Configuration{
				Excludes: []string{"*"},
			},
		)
		err := r.Refresh(conf.New())
		assert.ThatError(t, err).Matches("error parsing regexp: missing argument to repetition operator: `*`")
	})

	t.Run("mock error - 1", func(t *testing.T) {
		r := New()
		r.Provide(NewZeroLogger, gs_arg.Value("a")).
			Export(gs.As[Logger](), gs.As[CtxLogger]())
		r.AddMock(gs.BeanMock{
			Object: &SimpleLogger{},
			Target: gs.BeanSelectorFor[Logger](),
		})
		err := r.Refresh(conf.New())
		assert.ThatError(t, err).Matches("found unimplemented interface")
	})

	t.Run("mock error - 2", func(t *testing.T) {
		r := New()
		r.Object(&TestBean{Value: 1}).Name("TestBean-1")
		r.Object(&TestBean{Value: 2}).Name("TestBean-2")
		r.AddMock(gs.BeanMock{
			Object: &TestBean{},
			Target: gs.BeanSelectorFor[*TestBean](),
		})
		err := r.Refresh(conf.New())
		assert.ThatError(t, err).Matches("found duplicate mocked beans")
	})

	t.Run("resolve error - 1", func(t *testing.T) {
		r := New()
		r.Object(&TestBean{Value: 1}).Condition(
			gs_cond.OnFunc(func(ctx gs.CondContext) (bool, error) {
				return false, errors.New("condition error")
			}),
		)
		err := r.Refresh(conf.New())
		assert.ThatError(t, err).Matches("condition matches error: .* << condition error")
	})

	t.Run("resolve error - 2", func(t *testing.T) {
		r := New()
		r.Object(&TestBean{Value: 1}).Condition(
			gs_cond.OnBean[*TestBean](),
		)
		r.Object(&TestBean{Value: 1}).Condition(
			gs_cond.OnFunc(func(ctx gs.CondContext) (bool, error) {
				return false, errors.New("condition error")
			}),
		)
		err := r.Refresh(conf.New())
		assert.ThatError(t, err).Matches("condition matches error: .* << condition error")
	})

	t.Run("duplicate bean - 1", func(t *testing.T) {
		r := New()
		r.Object(&TestBean{Value: 1})
		r.Object(&TestBean{Value: 2})
		err := r.Refresh(conf.New())
		assert.ThatError(t, err).Matches("found duplicate beans")
	})

	t.Run("duplicate bean - 1", func(t *testing.T) {
		r := New()
		r.Object(&ZeroLogger{}).Name("a").Export(gs.As[Logger]())
		r.Object(&SimpleLogger{}).Name("a").Export(gs.As[Logger]())
		err := r.Refresh(conf.New())
		assert.ThatError(t, err).Matches("found duplicate beans")
	})

	t.Run("repeat refresh", func(t *testing.T) {
		r := New()
		err := r.Refresh(conf.New())
		assert.That(t, err).Nil()
		err = r.Refresh(conf.New())
		assert.ThatError(t, err).Matches("container is already refreshing or refreshed")
	})

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
					l := gs_bean.NewBean(NewZeroLogger, arg).
						Export(gs.As[Logger](), gs.As[CtxLogger]()).
						Name(name)
					beans = append(beans, l)
				}
				return
			})
			r.Provide(NewLogger, gs_arg.Value("c")).Name("c")
			r.AddMock(gs.BeanMock{
				Object: NewZeroLogger("a@mocked"),
				Target: gs.BeanSelectorFor[Logger]("a"),
			})
			r.AddMock(gs.BeanMock{
				Object: NewZeroLogger("b@mocked"),
				Target: gs.BeanSelectorFor[CtxLogger]("b"),
			})
		}
		{
			b := r.Object(&http.Server{}).
				Condition(gs_cond.OnBean[*http.ServeMux]())
			assert.That(t, b.BeanRegistration().Name()).Equal("Server")
		}
		{
			b := r.Provide(http.NewServeMux).Name("ServeMux-1").
				Condition(gs_cond.OnProperty("Enable.ServeMux-1").HavingValue("true"))
			assert.That(t, b.BeanRegistration().Name()).Equal("ServeMux-1")
		}
		{
			b := r.Provide(http.NewServeMux).Name("ServeMux-2").
				Condition(gs_cond.OnProperty("Enable.ServeMux-2").HavingValue("true"))
			assert.That(t, b.BeanRegistration().Name()).Equal("ServeMux-2")
		}
		{
			b := r.Object(&TestBean{Value: 1}).Configuration().Name("TestBean")
			assert.That(t, b.BeanRegistration().Name()).Equal("TestBean")
			r.AddMock(gs.BeanMock{
				Object: &TestBean{Value: 2},
				Target: b,
			})
		}
		{
			b := r.Object(&TestBean{Value: 1}).Name("TestBean-2").
				Configuration(gs.Configuration{
					Excludes: []string{"^NewChild$"},
				})
			assert.That(t, b.BeanRegistration().Name()).Equal("TestBean-2")
			r.AddMock(gs.BeanMock{
				Object: &ChildBean{Value: 2},
				Target: gs.BeanSelectorFor[*ChildBean]("TestBean-2_NewChildV2"),
			})
		}
		{
			r.AddMock(gs.BeanMock{
				Object: &bytes.Buffer{},
				Target: gs.BeanSelectorFor[*bytes.Buffer](),
			})
		}
		{
			b := r.Object(&TestBean{Value: 3}).Name("TestBean-3")
			r.Provide((*TestBean).NewChild, b)
		}

		p := conf.Map(map[string]any{
			"logger": map[string]string{
				"a": "",
				"b": "",
			},
			"Enable": map[string]any{
				"ServeMux-2": true,
			},
		})
		err := r.Refresh(p)
		assert.That(t, err).Nil()

		var names []string
		for _, b := range r.Beans() {
			names = append(names, b.Name())
		}
		assert.That(t, names).Equal([]string{
			"c",
			"Server",
			"ServeMux-2",
			"TestBean",
			"TestBean-2",
			"TestBean-3",
			"NewChild",
			"a",
			"b",
			"TestBean-2_NewChildV2",
		})
	})
}
