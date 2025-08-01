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

	"github.com/go-spring/gs-assert/assert"
	"github.com/go-spring/gs-mock/gsmock"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/util"
)

func TestBeanStatus(t *testing.T) {
	assert.That(t, BeanStatus(-2).String()).Equal("unknown")
	assert.That(t, StatusDeleted.String()).Equal("deleted")
	assert.That(t, StatusDefault.String()).Equal("default")
	assert.That(t, StatusResolving.String()).Equal("resolving")
	assert.That(t, StatusResolved.String()).Equal("resolved")
	assert.That(t, StatusCreating.String()).Equal("creating")
	assert.That(t, StatusCreated.String()).Equal("created")
	assert.That(t, StatusWired.String()).Equal("wired")
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

func (t *TestBean) Dummy() int       { return t.dummy }
func (t *TestBean) Clone() *TestBean { return &TestBean{} }

func TestBeanDefinition(t *testing.T) {

	t.Run("normal", func(t *testing.T) {
		a := &TestBean{}
		v := reflect.ValueOf(a)

		bean := makeBean(v.Type(), v, nil, "test")
		assert.That(t, bean.Name()).Equal("test")
		assert.That(t, bean.Type()).Equal(reflect.TypeFor[*TestBean]())
		assert.That(t, bean.Value().Interface()).Equal(a)
		assert.That(t, bean.Interface()).Equal(a)
		assert.That(t, bean.Callable()).Nil()

		bean.SetStatus(StatusCreated)
		assert.That(t, StatusCreated).Equal(bean.Status())

		bean.SetCaller(1)
		assert.ThatString(t, bean.FileLine()).HasSuffix("gs/internal/gs_bean/bean_test.go:79")

		bean.SetName("test-1")
		assert.That(t, bean.Name()).Equal("test-1")

		beanType, beanName := bean.TypeAndName()
		assert.That(t, beanType).Equal(reflect.TypeFor[*TestBean]())
		assert.That(t, beanName).Equal("test-1")
		assert.ThatString(t, bean.String()).Matches(`name=test-1 .*/gs/internal/gs_bean/bean_test.go:79`)

		assert.That(t, bean.BeanRuntime.Callable()).Nil()
		assert.That(t, bean.BeanRuntime.Status()).Equal(StatusWired)
		assert.That(t, bean.BeanRuntime.String()).Equal("test-1")
	})

	t.Run("depends on", func(t *testing.T) {
		v := reflect.ValueOf(&TestBean{})
		bean := makeBean(v.Type(), v, nil, "test")
		selector := gs.BeanSelectorFor[*http.ServeMux]()
		bean.SetDependsOn(selector)
		assert.That(t, bean.DependsOn()).Equal([]gs.BeanSelector{selector})
	})

	t.Run("init function", func(t *testing.T) {
		v := reflect.ValueOf(&TestBean{})
		bean := makeBean(v.Type(), v, nil, "test")
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
		assert.That(t, util.FuncName(bean.Init())).Equal("gs_bean.InitTestBean")
		bean.SetInit(InitTestBeanV2)
		assert.That(t, util.FuncName(bean.Init())).Equal("gs_bean.InitTestBeanV2")
		bean.SetInitMethod("Init")
		assert.That(t, util.FuncName(bean.Init())).Equal("gs_bean.(*TestBean).Init")
		bean.SetInitMethod("InitV2")
		assert.That(t, util.FuncName(bean.Init())).Equal("gs_bean.(*TestBean).InitV2")
		assert.Panic(t, func() {
			bean.SetInitMethod("InitV3")
		}, "method InitV3 not found on type \\*gs_bean.TestBean")
	})

	t.Run("destroy function", func(t *testing.T) {
		v := reflect.ValueOf(&TestBean{})
		bean := makeBean(v.Type(), v, nil, "test")
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
		assert.That(t, util.FuncName(bean.Destroy())).Equal("gs_bean.DestroyTestBean")
		bean.SetDestroy(DestroyTestBeanV2)
		assert.That(t, util.FuncName(bean.Destroy())).Equal("gs_bean.DestroyTestBeanV2")
		bean.SetDestroyMethod("Destroy")
		assert.That(t, util.FuncName(bean.Destroy())).Equal("gs_bean.(*TestBean).Destroy")
		bean.SetDestroyMethod("DestroyV2")
		assert.That(t, util.FuncName(bean.Destroy())).Equal("gs_bean.(*TestBean).DestroyV2")
		assert.Panic(t, func() {
			bean.SetDestroyMethod("DestroyV3")
		}, "method DestroyV3 not found on type \\*gs_bean.TestBean")
	})

	t.Run("export", func(t *testing.T) {
		v := reflect.ValueOf(&TestBean{})
		bean := makeBean(v.Type(), v, nil, "test")
		bean.SetExport(gs.As[TestBeanInterface]())
		assert.That(t, len(bean.Exports())).Equal(1)
		bean.SetExport(gs.As[TestBeanInterface]())
		assert.That(t, len(bean.Exports())).Equal(1)
		assert.Panic(t, func() {
			bean.SetExport(reflect.TypeFor[int]())
		}, "only interface type can be exported")
		assert.Panic(t, func() {
			bean.SetExport(reflect.TypeFor[io.Reader]())
		}, "doesn't implement interface io.Reader")
	})

	t.Run("on profiles", func(t *testing.T) {
		v := reflect.ValueOf(&TestBean{})
		bean := makeBean(v.Type(), v, nil, "test")
		bean.OnProfiles("dev,test")
		assert.That(t, len(bean.Conditions())).Equal(1)

		t.Run("no profile property", func(t *testing.T) {
			m := gsmock.NewManager()
			ctx := gs.NewCondContextMockImpl(m)
			ctx.MockProp().ReturnValue("")

			for _, c := range bean.Conditions() {
				ok, err := c.Matches(ctx)
				assert.That(t, err).Nil()
				assert.That(t, ok).False()
			}
		})

		t.Run("profile property not match", func(t *testing.T) {
			m := gsmock.NewManager()
			ctx := gs.NewCondContextMockImpl(m)
			ctx.MockProp().ReturnValue("prod")

			for _, c := range bean.Conditions() {
				ok, err := c.Matches(ctx)
				assert.That(t, err).Nil()
				assert.That(t, ok).False()
			}
		})

		t.Run("profile property is dev", func(t *testing.T) {
			m := gsmock.NewManager()
			ctx := gs.NewCondContextMockImpl(m)
			ctx.MockProp().ReturnValue("dev")

			for _, c := range bean.Conditions() {
				ok, err := c.Matches(ctx)
				assert.That(t, err).Nil()
				assert.That(t, ok).True()
			}
		})

		t.Run("profile property is test", func(t *testing.T) {
			m := gsmock.NewManager()
			ctx := gs.NewCondContextMockImpl(m)
			ctx.MockProp().ReturnValue("test")

			for _, c := range bean.Conditions() {
				ok, err := c.Matches(ctx)
				assert.That(t, err).Nil()
				assert.That(t, ok).True()
			}
		})

		t.Run("profile property is dev&test", func(t *testing.T) {
			m := gsmock.NewManager()
			ctx := gs.NewCondContextMockImpl(m)
			ctx.MockProp().ReturnValue("dev,test")

			for _, c := range bean.Conditions() {
				ok, err := c.Matches(ctx)
				assert.That(t, err).Nil()
				assert.That(t, ok).True()
			}
		})
	})

	t.Run("configuration", func(t *testing.T) {
		v := reflect.ValueOf(&TestBean{})
		bean := makeBean(v.Type(), v, nil, "test")
		assert.That(t, bean.Configuration()).Nil()

		bean.SetConfiguration()
		assert.That(t, bean.Configuration()).NotNil()
		assert.That(t, bean.Configuration().Includes).Nil()
		assert.That(t, bean.Configuration().Excludes).Nil()

		bean.SetConfiguration(gs.Configuration{
			Includes: []string{"New.*"},
		})
		assert.That(t, bean.Configuration()).NotNil()
		assert.That(t, bean.Configuration().Includes).Equal([]string{"New.*"})
	})

	t.Run("mock success", func(t *testing.T) {
		v := reflect.ValueOf(&bytes.Buffer{})
		bean := makeBean(v.Type(), v, nil, "test")
		bean.SetExport(gs.As[io.Writer]())
		bean.SetMock(bytes.NewBufferString(""))
		assert.That(t, bean.Mocked()).True()
	})
}

func TestNewBean(t *testing.T) {

	t.Run("type error", func(t *testing.T) {

		assert.Panic(t, func() {
			NewBean(new(int))
		}, "bean must be ref type")

		assert.Panic(t, func() {
			var r **TestBean
			NewBean(r)
		}, "bean must be ref type")
	})

	t.Run("value error", func(t *testing.T) {
		assert.Panic(t, func() {
			NewBean((*TestBean)(nil))
		}, "bean can't be nil")
	})

	t.Run("object", func(t *testing.T) {
		bean := NewBean(&TestBean{})
		beanX := bean.BeanRegistration().(*BeanDefinition)
		assert.That(t, beanX.Name()).Equal("TestBean")
		assert.That(t, beanX.Type()).Equal(reflect.TypeFor[*TestBean]())
	})

	t.Run("object - reflect.Value", func(t *testing.T) {
		bean := NewBean(reflect.ValueOf(&TestBean{}))
		beanX := bean.BeanRegistration().(*BeanDefinition)
		assert.That(t, beanX.Name()).Equal("TestBean")
		assert.That(t, beanX.Type()).Equal(reflect.TypeFor[*TestBean]())
	})

	t.Run("function - reflect.Value", func(t *testing.T) {
		fn := func(int, int) string { return "" }
		bean := NewBean(reflect.ValueOf(fn)).Name("TestFunc")
		beanX := bean.BeanRegistration().(*BeanDefinition)
		assert.That(t, beanX.Name()).Equal("TestFunc")
		assert.That(t, beanX.Type()).Equal(reflect.TypeOf(fn))
	})

	t.Run("constructor error", func(t *testing.T) {

		assert.Panic(t, func() {
			NewBean(func() {})
		}, "constructor should be func\\(...\\)bean or func\\(...\\)\\(bean, error\\)")

		assert.Panic(t, func() {
			NewBean(func() (int, string) { return 0, "" })
		}, "constructor should be func\\(...\\)bean or func\\(...\\)\\(bean, error\\)")

		assert.Panic(t, func() {
			NewBean(
				func(a int, b int) int { return a + b },
				gs_arg.Tag("${v:=3}"),
				gs_arg.Index(1, gs_arg.Tag("${v:=3}")),
			)
		}, "NewArgList error << arguments must be all indexed or non-indexed")

		assert.Panic(t, func() {
			NewBean(func() int { return 0 })
		}, "bean must be ref type")
	})

	t.Run("constructor", func(t *testing.T) {
		fn := func(int, int) *TestBean { return nil }
		bean := NewBean(fn).Name("NewTestBean")
		beanX := bean.BeanRegistration().(*BeanDefinition)
		assert.That(t, beanX.Name()).Equal("NewTestBean")
		assert.That(t, beanX.Type()).Equal(reflect.TypeFor[*TestBean]())
	})

	t.Run("method - 1", func(t *testing.T) {
		bean := NewBean((*TestBean).Clone)
		beanX := bean.BeanRegistration().(*BeanDefinition)
		assert.That(t, beanX.Name()).Equal("Clone")
		assert.That(t, len(beanX.Conditions())).Equal(1)
	})

	t.Run("method - 2", func(t *testing.T) {
		parent := NewBean(&TestBean{})
		bean := NewBean((*TestBean).Clone, parent)
		beanX := bean.BeanRegistration().(*BeanDefinition)
		assert.That(t, beanX.Name()).Equal("Clone")
		assert.That(t, len(beanX.Conditions())).Equal(1)
	})

	t.Run("method - 3", func(t *testing.T) {
		parent := NewBean(&TestBean{})
		bean := NewBean((*TestBean).Clone, gs_arg.Index(0, parent))
		beanX := bean.BeanRegistration().(*BeanDefinition)
		assert.That(t, beanX.Name()).Equal("Clone")
		assert.That(t, len(beanX.Conditions())).Equal(1)
	})

	t.Run("method - 4", func(t *testing.T) {
		parent := gs.NewRegisteredBean(
			NewBean(&TestBean{}).BeanRegistration(),
		)
		bean := NewBean((*TestBean).Clone, parent)
		beanX := bean.BeanRegistration().(*BeanDefinition)
		assert.That(t, beanX.Name()).Equal("Clone")
		assert.That(t, len(beanX.Conditions())).Equal(1)
	})

	t.Run("method - 5", func(t *testing.T) {
		parent := gs.NewRegisteredBean(
			NewBean(&TestBean{}).BeanRegistration(),
		)
		bean := NewBean((*TestBean).Clone, gs_arg.Index(0, parent))
		beanX := bean.BeanRegistration().(*BeanDefinition)
		assert.That(t, beanX.Name()).Equal("Clone")
		assert.That(t, len(beanX.Conditions())).Equal(1)
	})

	t.Run("method error - 1", func(t *testing.T) {
		assert.Panic(t, func() {
			NewBean((*TestBean).Clone, gs_arg.Tag(""))
		}, "ctorArgs\\[0] should be \\*RegisteredBean or \\*BeanDefinition or IndexArg\\[0]")
	})

	t.Run("method error - 2", func(t *testing.T) {
		assert.Panic(t, func() {
			NewBean((*TestBean).Clone, gs_arg.Index(0, gs_arg.Tag("")))
		}, "the arg of IndexArg\\[0] should be \\*RegisteredBean or \\*BeanDefinition")
	})
}
