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

package injecting

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_dync"
	"github.com/go-spring/spring-core/util/assert"
)

type Logger interface {
	Print(msg string)
}

type CtxLogger interface {
	CtxPrint(ctx context.Context, msg string)
}

type SimpleLogger struct{}

func (l *SimpleLogger) Print(msg string) {}

type ZeroLogger struct {
	File string
}

func NewZeroLogger(file string) *ZeroLogger {
	return &ZeroLogger{File: file}
}

func (l *ZeroLogger) Print(msg string) {}

func (l *ZeroLogger) CtxPrint(ctx context.Context, msg string) {}

type CycleBean struct {
	Bean *LazyBean `autowire:""`
}

type LazyBean struct {
	Bean *CycleBean `autowire:""`
}

type Filter interface {
	Do(ctx context.Context)
}

type Controller struct {
	Loggers []Logger `inject:"biz,sys"`
	Service *Service `autowire:""`
	Filters []Filter `inject:"?"`
}

type Service struct {
	Loggers    map[string]CtxLogger `inject:"*,sys?"`
	Repository *Repository          `inject:""`
	Status     int
}

func (s *Service) Destroy() {
	s.Status = 0
}

type Repository struct {
	Addr gs_dync.Value[string] `value:"${addr:=127.0.0.1:5050}"`
	stop chan struct{}
}

func (r *Repository) Init() {
	r.stop = make(chan struct{})
	go func() {
		l := r.Addr.NewListener()
		for {
			select {
			case <-l.C:
				fmt.Println(r.Addr.Value())
			case <-r.stop:
				return
			}
		}
	}()
}

func (r *Repository) GetAddr() string {
	return r.Addr.Value()
}

func objectBean(i interface{}) *gs.BeanDefinition {
	return gs_bean.NewBean(reflect.ValueOf(i))
}

func provideBean(ctor interface{}, args ...gs.Arg) *gs.BeanDefinition {
	return gs_bean.NewBean(ctor, args...)
}

func extractBeans(beans []*gs.BeanDefinition) []*gs_bean.BeanDefinition {
	var ret []*gs_bean.BeanDefinition
	for _, b := range beans {
		ret = append(ret, b.BeanRegistration().(*gs_bean.BeanDefinition))
	}
	return ret
}

func TestInjecting(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		r := New()

		beans := []*gs.BeanDefinition{
			objectBean(&CycleBean{}),
			objectBean(&LazyBean{}).Destroy(func(*LazyBean) {}),
			objectBean(&Repository{}).InitMethod("Init").Destroy(func(r *Repository) {
				r.stop <- struct{}{}
			}),
			objectBean(&Controller{}),
			objectBean(&Service{}).DestroyMethod("Destroy").Init(func(s *Service) {
				s.Status = 1
			}),
			objectBean(&SimpleLogger{}).Name("sys").Export(gs.As[Logger]()),
			provideBean(NewZeroLogger, gs_arg.Tag("${logger.biz.file}")).Name("biz").
				Export(gs.As[Logger](), gs.As[CtxLogger]()),
		}

		err := r.RefreshProperties(conf.Map(map[string]interface{}{
			"spring": map[string]interface{}{
				"allow-circular-references": true,
			},
			"logger": map[string]interface{}{
				"biz": map[string]interface{}{
					"file": "biz.log",
				},
			},
		}))
		assert.Nil(t, err)

		err = r.Refresh(extractBeans(beans))
		assert.Nil(t, err)

		time.Sleep(time.Millisecond * 50)

		c := &Controller{}
		err = r.Wire(c)
		assert.Nil(t, err)
		assert.Equal(t, len(c.Loggers), 2)

		s := &struct {
			Service *Service `autowire:""`
		}{}
		err = r.Wire(s)
		assert.Nil(t, err)
		assert.Equal(t, s.Service.Status, 1)

		r.Close()

		assert.Equal(t, s.Service.Status, 0)
	})
}

func TestWireTag(t *testing.T) {

	t.Run("empty str", func(t *testing.T) {
		tag, err := parseWireTag(conf.New(), "")
		assert.Nil(t, err)
		assert.Equal(t, tag, WireTag{})
		assert.Equal(t, tag.String(), "")
	})

	t.Run("only name", func(t *testing.T) {
		tag, err := parseWireTag(conf.New(), "a")
		assert.Nil(t, err)
		assert.Equal(t, tag, WireTag{beanName: "a"})
		assert.Equal(t, tag.String(), "a")
	})

	t.Run("only nullable", func(t *testing.T) {
		tag, err := parseWireTag(conf.New(), "?")
		assert.Nil(t, err)
		assert.Equal(t, tag, WireTag{nullable: true})
		assert.Equal(t, tag.String(), "?")
	})

	t.Run("name and nullable", func(t *testing.T) {
		tag, err := parseWireTag(conf.New(), "a?")
		assert.Nil(t, err)
		assert.Equal(t, tag, WireTag{beanName: "a", nullable: true})
		assert.Equal(t, tag.String(), "a?")
	})

	t.Run("resolve error", func(t *testing.T) {
		_, err := parseWireTag(conf.New(), "${?")
		assert.Error(t, err, "resolve string .* error: invalid syntax")
	})

	t.Run("resolve success", func(t *testing.T) {
		tag, err := parseWireTag(conf.New(), "${k:=a}?")
		assert.Nil(t, err)
		assert.Equal(t, tag, WireTag{beanName: "a", nullable: true})
		assert.Equal(t, tag.String(), "a?")
	})

	t.Run("tags - 1", func(t *testing.T) {
		tags := []WireTag{
			{"a", true},
		}
		assert.Equal(t, toWireString(tags), "a?")
	})

	t.Run("tags - 2", func(t *testing.T) {
		tags := []WireTag{
			{"a", true},
			{"b", false},
		}
		assert.Equal(t, toWireString(tags), "a?,b")
	})

	t.Run("tags - 3", func(t *testing.T) {
		tags := []WireTag{
			{"a", true},
			{"b", false},
			{"c", true},
		}
		assert.Equal(t, toWireString(tags), "a?,b,c?")
	})
}
