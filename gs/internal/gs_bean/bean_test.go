/*
 * Copyright 2024 The Go-Spring Authors.
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

package gs_bean

import (
	"bytes"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/util"
	"github.com/go-spring/spring-core/util/assert"
	"go.uber.org/mock/gomock"
)

func TestBeanStatus(t *testing.T) {
	assert.Equal(t, BeanStatus(-2).String(), "unknown")
	assert.Equal(t, StatusDeleted.String(), "deleted")
	assert.Equal(t, StatusDefault.String(), "default")
	assert.Equal(t, StatusResolving.String(), "resolving")
	assert.Equal(t, StatusResolved.String(), "resolved")
	assert.Equal(t, StatusCreating.String(), "creating")
	assert.Equal(t, StatusCreated.String(), "created")
	assert.Equal(t, StatusWired.String(), "wired")
}

type TestBeanInterface interface {
	Dummy() int
}

type TestBean struct{ dummy int }

func (*TestBean) Init()            {}
func (*TestBean) InitV2() error    { return nil }
func (*TestBean) Destroy()         {}
func (*TestBean) DestroyV2() error { return nil }

func InitTestBean(*TestBean)            {}
func InitTestBeanV2(*TestBean) error    { return nil }
func DestroyTestBean(*TestBean)         {}
func DestroyTestBeanV2(*TestBean) error { return nil }

func (t *TestBean) Dummy() int { return t.dummy }

func TestBeanDefinition(t *testing.T) {

	t.Run("normal", func(t *testing.T) {
		a := &TestBean{}
		v := reflect.ValueOf(a)

		bean := NewBean(v.Type(), v, nil, "test")
		assert.Equal(t, bean.Name(), "test")
		assert.Equal(t, bean.Type(), reflect.TypeFor[*TestBean]())
		assert.Equal(t, bean.Value().Interface(), a)
		assert.Equal(t, bean.Interface(), a)
		assert.Nil(t, bean.Callable())

		bean.SetStatus(StatusCreated)
		assert.Equal(t, StatusCreated, bean.Status())

		bean.SetCaller(1)
		assert.String(t, bean.FileLine()).HasSuffix("gs/internal/gs_bean/bean_test.go:77")

		bean.SetName("test-1")
		assert.Equal(t, bean.Name(), "test-1")

		beanType, beanName := bean.TypeAndName()
		assert.Equal(t, beanType, reflect.TypeFor[*TestBean]())
		assert.Equal(t, beanName, "test-1")
		assert.Matches(t, bean.String(), `name=test-1 .*/gs/internal/gs_bean/bean_test.go:77`)

		assert.Nil(t, bean.BeanRuntime.Callable())
		assert.Equal(t, bean.BeanRuntime.Status(), StatusWired)
		assert.Equal(t, bean.BeanRuntime.String(), "test-1")
	})

	t.Run("depends on", func(t *testing.T) {
		v := reflect.ValueOf(&TestBean{})
		bean := NewBean(v.Type(), v, nil, "test")
		selector := gs.BeanSelectorFor[*http.ServeMux]()
		bean.SetDependsOn(selector)
		assert.Equal(t, bean.DependsOn(), []gs.BeanSelector{selector})
	})

	t.Run("init function", func(t *testing.T) {
		v := reflect.ValueOf(&TestBean{})
		bean := NewBean(v.Type(), v, nil, "test")
		assert.Panic(t, func() {
			bean.SetInit(3)
		}, "init should be func\\(bean\\) or func\\(bean\\)error")
		assert.Panic(t, func() {
			bean.SetInit(func() {})
		}, "init should be func\\(bean\\) or func\\(bean\\)error")
		assert.Panic(t, func() {
			bean.SetInit(func(int, string) {})
		}, "init should be func\\(bean\\) or func\\(bean\\)error")
		assert.Panic(t, func() {
			bean.SetInit(func(io.Reader) {})
		}, "init should be func\\(bean\\) or func\\(bean\\)error")
		assert.Panic(t, func() {
			bean.SetInit(func(*bytes.Buffer) {})
		}, "init should be func\\(bean\\) or func\\(bean\\)error")
		bean.SetInit(func(TestBeanInterface) {})
		assert.Panic(t, func() {
			bean.SetInit(func(TestBeanInterface) int { return 0 })
		}, "init should be func\\(bean\\) or func\\(bean\\)error")
		bean.SetInit(func(TestBeanInterface) error { return nil })
		assert.Panic(t, func() {
			bean.SetInit(func(TestBeanInterface) (int, error) { return 0, nil })
		}, "init should be func\\(bean\\) or func\\(bean\\)error")
		bean.SetInit(InitTestBean)
		assert.Equal(t, util.FuncName(bean.Init()), "gs_bean.InitTestBean")
		bean.SetInit(InitTestBeanV2)
		assert.Equal(t, util.FuncName(bean.Init()), "gs_bean.InitTestBeanV2")
		bean.SetInitMethod("Init")
		assert.Equal(t, util.FuncName(bean.Init()), "gs_bean.(*TestBean).Init")
		bean.SetInitMethod("InitV2")
		assert.Equal(t, util.FuncName(bean.Init()), "gs_bean.(*TestBean).InitV2")
		assert.Panic(t, func() {
			bean.SetInitMethod("InitV3")
		}, "method InitV3 not found on type \\*gs_bean.TestBean")
	})

	t.Run("destroy function", func(t *testing.T) {
		v := reflect.ValueOf(&TestBean{})
		bean := NewBean(v.Type(), v, nil, "test")
		assert.Panic(t, func() {
			bean.SetDestroy(3)
		}, "destroy should be func\\(bean\\) or func\\(bean\\)error")
		assert.Panic(t, func() {
			bean.SetDestroy(func() {})
		}, "destroy should be func\\(bean\\) or func\\(bean\\)error")
		assert.Panic(t, func() {
			bean.SetDestroy(func(int, string) {})
		}, "destroy should be func\\(bean\\) or func\\(bean\\)error")
		assert.Panic(t, func() {
			bean.SetDestroy(func(io.Reader) {})
		}, "destroy should be func\\(bean\\) or func\\(bean\\)error")
		assert.Panic(t, func() {
			bean.SetDestroy(func(*bytes.Buffer) {})
		}, "destroy should be func\\(bean\\) or func\\(bean\\)error")
		bean.SetDestroy(func(TestBeanInterface) {})
		assert.Panic(t, func() {
			bean.SetDestroy(func(TestBeanInterface) int { return 0 })
		}, "destroy should be func\\(bean\\) or func\\(bean\\)error")
		bean.SetDestroy(func(TestBeanInterface) error { return nil })
		assert.Panic(t, func() {
			bean.SetDestroy(func(TestBeanInterface) (int, error) { return 0, nil })
		}, "destroy should be func\\(bean\\) or func\\(bean\\)error")
		bean.SetDestroy(DestroyTestBean)
		assert.Equal(t, util.FuncName(bean.Destroy()), "gs_bean.DestroyTestBean")
		bean.SetDestroy(DestroyTestBeanV2)
		assert.Equal(t, util.FuncName(bean.Destroy()), "gs_bean.DestroyTestBeanV2")
		bean.SetDestroyMethod("Destroy")
		assert.Equal(t, util.FuncName(bean.Destroy()), "gs_bean.(*TestBean).Destroy")
		bean.SetDestroyMethod("DestroyV2")
		assert.Equal(t, util.FuncName(bean.Destroy()), "gs_bean.(*TestBean).DestroyV2")
		assert.Panic(t, func() {
			bean.SetDestroyMethod("DestroyV3")
		}, "method DestroyV3 not found on type \\*gs_bean.TestBean")
	})

	t.Run("export", func(t *testing.T) {
		v := reflect.ValueOf(&TestBean{})
		bean := NewBean(v.Type(), v, nil, "test")
		bean.SetExport(gs.As[TestBeanInterface]())
		assert.Equal(t, len(bean.Exports()), 1)
		bean.SetExport(gs.As[TestBeanInterface]())
		assert.Equal(t, len(bean.Exports()), 1)
		assert.Panic(t, func() {
			bean.SetExport(reflect.TypeFor[int]())
		}, "only interface type can be exported")
		assert.Panic(t, func() {
			bean.SetExport(reflect.TypeFor[io.Reader]())
		}, "doesn't implement interface io.Reader")
	})

	t.Run("on profiles", func(t *testing.T) {
		v := reflect.ValueOf(&TestBean{})
		bean := NewBean(v.Type(), v, nil, "test")
		bean.OnProfiles("dev,test")
		assert.Equal(t, len(bean.Conditions()), 1)

		t.Run("no profile property", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := gs.NewMockCondContext(ctrl)
			ctx.EXPECT().Prop(gomock.Any()).Return("")
			for _, c := range bean.Conditions() {
				ok, err := c.Matches(ctx)
				assert.Nil(t, err)
				assert.False(t, ok)
			}
		})

		t.Run("profile property not match", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := gs.NewMockCondContext(ctrl)
			ctx.EXPECT().Prop(gomock.Any()).Return("prod")
			for _, c := range bean.Conditions() {
				ok, err := c.Matches(ctx)
				assert.Nil(t, err)
				assert.False(t, ok)
			}
		})

		t.Run("profile property is dev", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := gs.NewMockCondContext(ctrl)
			ctx.EXPECT().Prop(gomock.Any()).Return("dev")
			for _, c := range bean.Conditions() {
				ok, err := c.Matches(ctx)
				assert.Nil(t, err)
				assert.True(t, ok)
			}
		})

		t.Run("profile property is test", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := gs.NewMockCondContext(ctrl)
			ctx.EXPECT().Prop(gomock.Any()).Return("test")
			for _, c := range bean.Conditions() {
				ok, err := c.Matches(ctx)
				assert.Nil(t, err)
				assert.True(t, ok)
			}
		})

		t.Run("profile property is dev&test", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := gs.NewMockCondContext(ctrl)
			ctx.EXPECT().Prop(gomock.Any()).Return("dev,test")
			for _, c := range bean.Conditions() {
				ok, err := c.Matches(ctx)
				assert.Nil(t, err)
				assert.True(t, ok)
			}
		})
	})

	t.Run("configuration", func(t *testing.T) {
		v := reflect.ValueOf(&TestBean{})
		bean := NewBean(v.Type(), v, nil, "test")
		assert.Nil(t, bean.Configuration())

		bean.SetConfiguration()
		assert.NotNil(t, bean.Configuration())
		assert.Nil(t, bean.Configuration().Includes)
		assert.Nil(t, bean.Configuration().Excludes)

		bean.SetConfiguration(gs.Configuration{
			Includes: []string{"New.*"},
		})
		assert.NotNil(t, bean.Configuration())
		assert.Equal(t, bean.Configuration().Includes, []string{"New.*"})
	})

	t.Run("mock success", func(t *testing.T) {
		v := reflect.ValueOf(&bytes.Buffer{})
		bean := NewBean(v.Type(), v, nil, "test")
		bean.SetExport(gs.As[io.Writer]())
		bean.SetMock(bytes.NewBufferString(""))
	})
}
