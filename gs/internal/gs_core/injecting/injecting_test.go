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
		r := New(conf.Map(map[string]interface{}{
			"spring": map[string]interface{}{
				"allow-circular-references": true,
			},
			"logger": map[string]interface{}{
				"biz": map[string]interface{}{
					"file": "biz.log",
				},
			},
		}))

		beans := []*gs.BeanDefinition{
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

		err := r.Refresh(extractBeans(beans))
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

type A struct {
	B *B `autowire:""`
}

type B struct {
	C *C `autowire:""`
}

type C struct {
	A *A `autowire:"?"`
	D *D `autowire:"?"`
}

type D struct {
	E *E `autowire:""`
}

type E struct {
	c *C
	g *G
}

func NewE(c *C, g *G) *E {
	return &E{c: c, g: g}
}

type F struct {
	G *G `autowire:""`
}

type G struct {
	e *E
}

func NewG(e *E) *G {
	return &G{e: e}
}

type H struct {
	i *I `autowire:""`
}

func NewH(i *I) *H {
	return &H{i: i}
}

type I struct {
	J *J `autowire:""`
}

type J struct {
	H *H `autowire:",lazy"`
}

func NewJ() *J {
	return &J{}
}

func TestCircularBean(t *testing.T) {

	t.Run("not truly circular - 1", func(t *testing.T) {
		r := New(conf.New())
		beans := []*gs.BeanDefinition{
			objectBean(&A{}),
			objectBean(&B{}),
			objectBean(&C{}),
		}
		err := r.Refresh(extractBeans(beans))
		assert.Nil(t, err)
		var s struct {
			A *A `autowire:""`
			B *B `autowire:""`
			C *C `autowire:""`
		}
		err = r.Wire(&s)
		assert.Nil(t, err)
		assert.Equal(t, s.A.B, s.B)
		assert.Equal(t, s.B.C, s.C)
		assert.Equal(t, s.C.A, s.A)
	})

	t.Run("not truly circular - 2", func(t *testing.T) {
		r := New(conf.New())
		beans := []*gs.BeanDefinition{
			objectBean(&C{}),
			objectBean(&D{}),
			provideBean(NewE, gs_arg.Index(1, gs_arg.Tag("?"))),
		}
		err := r.Refresh(extractBeans(beans))
		assert.Nil(t, err)
		var s struct {
			C *C `autowire:""`
			D *D `autowire:""`
			E *E `autowire:""`
		}
		err = r.Wire(&s)
		assert.Nil(t, err)
		assert.Equal(t, s.C.D, s.D)
		assert.Equal(t, s.D.E, s.E)
		assert.Equal(t, s.E.c, s.C)
	})

	t.Run("found circular - 1", func(t *testing.T) {
		r := New(conf.New())
		beans := []*gs.BeanDefinition{
			provideBean(NewE, gs_arg.Tag("?")),
			objectBean(&F{}),
			provideBean(NewG),
		}
		err := r.Refresh(extractBeans(beans))
		assert.Error(t, err, "found circular autowire")
	})

	t.Run("found circular - 2", func(t *testing.T) {
		r := New(conf.New())
		beans := []*gs.BeanDefinition{
			provideBean(NewH),
			objectBean(&I{}),
			provideBean(NewJ),
		}
		err := r.Refresh(extractBeans(beans))
		assert.Error(t, err, "found circular autowire")
	})

	t.Run("found circular - 3", func(t *testing.T) {
		r := New(conf.Map(map[string]interface{}{
			"spring": map[string]interface{}{
				"allow-circular-references": true,
			},
		}))
		beans := []*gs.BeanDefinition{
			provideBean(NewH),
			objectBean(&I{}),
			provideBean(NewJ),
		}
		err := r.Refresh(extractBeans(beans))
		assert.Nil(t, err)
		var s struct {
			H *H `autowire:""`
			I *I `autowire:""`
			J *J `autowire:""`
		}
		err = r.Wire(&s)
		assert.Nil(t, err)
		assert.Equal(t, s.H.i, s.I)
		assert.Equal(t, s.I.J, s.J)
		assert.Equal(t, s.J.H, s.H)
	})
}

type DestroyA struct {
	called bool
}

type DestroyB struct {
	called bool
}

func (d *DestroyB) Destroy() {
	d.called = true
}

type DestroyC struct {
	called   bool
	DestroyD *DestroyD `autowire:""`
}

type DestroyD struct {
	DestroyE *DestroyE `autowire:""`
}

type DestroyE struct {
	called bool
}

func (d *DestroyE) Destroy() {
	d.called = true
}

func TestDestroy(t *testing.T) {

	t.Run("normal", func(t *testing.T) {
		r := New(conf.New())
		beans := []*gs.BeanDefinition{
			objectBean(&DestroyA{}).Destroy(func(d *DestroyA) {
				d.called = true
			}),
			objectBean(&DestroyB{}).DestroyMethod("Destroy"),
		}
		err := r.Refresh(extractBeans(beans))
		assert.Nil(t, err)
		var s struct {
			DestroyA *DestroyA `autowire:""`
			DestroyB *DestroyB `autowire:""`
		}
		err = r.Wire(&s)
		assert.Nil(t, err)
		r.Close()
		assert.True(t, s.DestroyA.called)
		assert.True(t, s.DestroyB.called)
	})

	t.Run("dependency", func(t *testing.T) {
		r := New(conf.New())
		beans := []*gs.BeanDefinition{
			objectBean(&DestroyC{}).Destroy(func(d *DestroyC) {
				d.called = true
			}),
			objectBean(&DestroyD{}),
			objectBean(&DestroyE{}).DestroyMethod("Destroy"),
		}
		err := r.Refresh(extractBeans(beans))
		assert.Nil(t, err)
		var s struct {
			DestroyC *DestroyC `autowire:""`
			DestroyE *DestroyE `autowire:""`
		}
		err = r.Wire(&s)
		assert.Nil(t, err)
		r.Close()
		assert.True(t, s.DestroyC.called)
		assert.True(t, s.DestroyE.called)
	})
}
