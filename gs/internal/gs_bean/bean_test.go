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

	"github.com/go-spring/gs-mock/gsmock"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/stdlib/funcutil"
	"github.com/go-spring/stdlib/testing/assert"
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
		assert.That(t, bean.GetName()).Equal("test")
		assert.That(t, bean.GetType()).Equal(reflect.TypeFor[*TestBean]())
		assert.That(t, bean.GetValue().Interface()).Equal(a)
		assert.That(t, bean.Interface()).Equal(a)
		assert.That(t, bean.Callable()).Nil()

		bean.SetStatus(StatusCreated)
		assert.That(t, StatusCreated).Equal(bean.Status())

		bean.Caller(1)
		assert.String(t, bean.FileLine()).HasSuffix("gs/internal/gs_bean/bean_test.go:79")

		bean.Name("test-1")
		assert.That(t, bean.GetName()).Equal("test-1")
	})

	t.Run("depends on", func(t *testing.T) {
		v := reflect.ValueOf(&TestBean{})
		bean := makeBean(v.Type(), v, nil, "test")
		selector := gs.BeanIDFor[*http.ServeMux]()
		bean.DependsOn(selector)
		assert.That(t, bean.GetDependsOn()).Equal([]gs.BeanID{selector})
	})

	t.Run("init function", func(t *testing.T) {
		v := reflect.ValueOf(&TestBean{})
		bean := makeBean(v.Type(), v, nil, "test")
		assert.Panic(t, func() {
			bean.Init(3)
		}, "init should be func\\(bean\\) or func\\(bean\\)error")
		assert.Panic(t, func() {
			bean.Init(func() {})
		}, "init should be func\\(bean\\) or func\\(bean\\)error")
		assert.Panic(t, func() {
			bean.Init(func(int, string) {})
		}, "init should be func\\(bean\\) or func\\(bean\\)error")
		assert.Panic(t, func() {
			bean.Init(func(io.Reader) {})
		}, "init should be func\\(bean\\) or func\\(bean\\)error")
		assert.Panic(t, func() {
			bean.Init(func(*bytes.Buffer) {})
		}, "init should be func\\(bean\\) or func\\(bean\\)error")
		bean.Init(func(TestBeanInterface) {})
		assert.Panic(t, func() {
			bean.Init(func(TestBeanInterface) int { return 0 })
		}, "init should be func\\(bean\\) or func\\(bean\\)error")
		bean.Init(func(TestBeanInterface) error { return nil })
		assert.Panic(t, func() {
			bean.Init(func(TestBeanInterface) (int, error) { return 0, nil })
		}, "init should be func\\(bean\\) or func\\(bean\\)error")
		bean.Init(InitTestBean)
		assert.That(t, funcutil.FuncName(bean.GetInit())).Equal("gs_bean.InitTestBean")
		bean.Init(InitTestBeanV2)
		assert.That(t, funcutil.FuncName(bean.GetInit())).Equal("gs_bean.InitTestBeanV2")
		bean.InitMethod("Init")
		assert.That(t, funcutil.FuncName(bean.GetInit())).Equal("gs_bean.(*TestBean).Init")
		bean.InitMethod("InitV2")
		assert.That(t, funcutil.FuncName(bean.GetInit())).Equal("gs_bean.(*TestBean).InitV2")
		assert.Panic(t, func() {
			bean.InitMethod("InitV3")
		}, "method InitV3 not found on type \\*gs_bean.TestBean")
	})

	t.Run("destroy function", func(t *testing.T) {
		v := reflect.ValueOf(&TestBean{})
		bean := makeBean(v.Type(), v, nil, "test")
		assert.Panic(t, func() {
			bean.Destroy(3)
		}, "destroy should be func\\(bean\\) or func\\(bean\\)error")
		assert.Panic(t, func() {
			bean.Destroy(func() {})
		}, "destroy should be func\\(bean\\) or func\\(bean\\)error")
		assert.Panic(t, func() {
			bean.Destroy(func(int, string) {})
		}, "destroy should be func\\(bean\\) or func\\(bean\\)error")
		assert.Panic(t, func() {
			bean.Destroy(func(io.Reader) {})
		}, "destroy should be func\\(bean\\) or func\\(bean\\)error")
		assert.Panic(t, func() {
			bean.Destroy(func(*bytes.Buffer) {})
		}, "destroy should be func\\(bean\\) or func\\(bean\\)error")
		bean.Destroy(func(TestBeanInterface) {})
		assert.Panic(t, func() {
			bean.Destroy(func(TestBeanInterface) int { return 0 })
		}, "destroy should be func\\(bean\\) or func\\(bean\\)error")
		bean.Destroy(func(TestBeanInterface) error { return nil })
		assert.Panic(t, func() {
			bean.Destroy(func(TestBeanInterface) (int, error) { return 0, nil })
		}, "destroy should be func\\(bean\\) or func\\(bean\\)error")
		bean.Destroy(DestroyTestBean)
		assert.That(t, funcutil.FuncName(bean.GetDestroy())).Equal("gs_bean.DestroyTestBean")
		bean.Destroy(DestroyTestBeanV2)
		assert.That(t, funcutil.FuncName(bean.GetDestroy())).Equal("gs_bean.DestroyTestBeanV2")
		bean.DestroyMethod("Destroy")
		assert.That(t, funcutil.FuncName(bean.GetDestroy())).Equal("gs_bean.(*TestBean).Destroy")
		bean.DestroyMethod("DestroyV2")
		assert.That(t, funcutil.FuncName(bean.GetDestroy())).Equal("gs_bean.(*TestBean).DestroyV2")
		assert.Panic(t, func() {
			bean.DestroyMethod("DestroyV3")
		}, "method DestroyV3 not found on type \\*gs_bean.TestBean")
	})

	t.Run("export", func(t *testing.T) {
		v := reflect.ValueOf(&TestBean{})
		bean := makeBean(v.Type(), v, nil, "test")
		bean.Export(gs.As[TestBeanInterface]())
		assert.That(t, len(bean.Exports())).Equal(1)
		bean.Export(gs.As[TestBeanInterface]())
		assert.That(t, len(bean.Exports())).Equal(1)
		assert.Panic(t, func() {
			bean.Export(reflect.TypeFor[int]())
		}, "only interface type can be exported")
		assert.Panic(t, func() {
			bean.Export(reflect.TypeFor[io.Reader]())
		}, "doesn't implement interface io.Reader")
	})

	t.Run("on profiles", func(t *testing.T) {
		v := reflect.ValueOf(&TestBean{})
		bean := makeBean(v.Type(), v, nil, "test")
		bean.OnProfiles("dev,test")
		assert.That(t, len(bean.Conditions())).Equal(1)

		t.Run("no profile property", func(t *testing.T) {
			m := gsmock.NewManager()
			ctx := gs.NewConditionContextMockImpl(m)
			ctx.MockProp().ReturnValue("")

			for _, c := range bean.Conditions() {
				ok, err := c.Matches(ctx)
				assert.That(t, err).Nil()
				assert.That(t, ok).False()
			}
		})

		t.Run("profile property not match", func(t *testing.T) {
			m := gsmock.NewManager()
			ctx := gs.NewConditionContextMockImpl(m)
			ctx.MockProp().ReturnValue("prod")

			for _, c := range bean.Conditions() {
				ok, err := c.Matches(ctx)
				assert.That(t, err).Nil()
				assert.That(t, ok).False()
			}
		})

		t.Run("profile property is dev", func(t *testing.T) {
			m := gsmock.NewManager()
			ctx := gs.NewConditionContextMockImpl(m)
			ctx.MockProp().ReturnValue("dev")

			for _, c := range bean.Conditions() {
				ok, err := c.Matches(ctx)
				assert.That(t, err).Nil()
				assert.That(t, ok).True()
			}
		})

		t.Run("profile property is test", func(t *testing.T) {
			m := gsmock.NewManager()
			ctx := gs.NewConditionContextMockImpl(m)
			ctx.MockProp().ReturnValue("test")

			for _, c := range bean.Conditions() {
				ok, err := c.Matches(ctx)
				assert.That(t, err).Nil()
				assert.That(t, ok).True()
			}
		})

		t.Run("profile property is dev&test", func(t *testing.T) {
			m := gsmock.NewManager()
			ctx := gs.NewConditionContextMockImpl(m)
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
		assert.That(t, bean.GetConfiguration()).Nil()

		bean.Configuration()
		assert.That(t, bean.GetConfiguration()).NotNil()
		assert.That(t, bean.GetConfiguration().Includes).Nil()
		assert.That(t, bean.GetConfiguration().Excludes).Nil()

		bean.Configuration(Configuration{
			Includes: []string{"New.*"},
		})
		assert.That(t, bean.GetConfiguration()).NotNil()
		assert.That(t, bean.GetConfiguration().Includes).Equal([]string{"New.*"})
	})
}

func TestNewBean(t *testing.T) {

	t.Run("invalid bean type", func(t *testing.T) {

		assert.Panic(t, func() {
			NewBean(new(int))
		}, "bean must be ref type")

		assert.Panic(t, func() {
			var r **TestBean
			NewBean(r)
		}, "bean must be ref type")
	})

	t.Run("nil bean value", func(t *testing.T) {
		assert.Panic(t, func() {
			NewBean((*TestBean)(nil))
		}, "bean can't be nil")
	})

	t.Run("object", func(t *testing.T) {
		bean := NewBean(&TestBean{})
		assert.That(t, bean.GetName()).Equal("TestBean")
		assert.That(t, bean.GetType()).Equal(reflect.TypeFor[*TestBean]())
	})

	t.Run("object by reflect.Value", func(t *testing.T) {
		bean := NewBean(reflect.ValueOf(&TestBean{}))
		assert.That(t, bean.GetName()).Equal("TestBean")
		assert.That(t, bean.GetType()).Equal(reflect.TypeFor[*TestBean]())
	})

	t.Run("function by reflect.Value", func(t *testing.T) {
		fn := func(int, int) string { return "" }
		bean := NewBean(reflect.ValueOf(fn)).Name("TestFunc")
		assert.That(t, bean.GetName()).Equal("TestFunc")
		assert.That(t, bean.GetType()).Equal(reflect.TypeFor[func(int, int) string]())
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
		}, "NewArgList error: arguments must be all indexed or non-indexed")

		assert.Panic(t, func() {
			NewBean(func() int { return 0 })
		}, "bean must be ref type")
	})

	t.Run("constructor success", func(t *testing.T) {
		fn := func(int, int) *TestBean { return nil }
		bean := NewBean(fn).Name("NewTestBean")
		assert.That(t, bean.GetName()).Equal("NewTestBean")
		assert.That(t, bean.GetType()).Equal(reflect.TypeFor[*TestBean]())
	})

	t.Run("method - 1", func(t *testing.T) {
		bean := NewBean((*TestBean).Clone)
		assert.That(t, bean.GetName()).Equal("Clone")
		assert.That(t, len(bean.Conditions())).Equal(1)
	})

	t.Run("method - 2", func(t *testing.T) {
		parent := NewBean(&TestBean{})
		bean := NewBean((*TestBean).Clone, parent)
		assert.That(t, bean.GetName()).Equal("Clone")
		assert.That(t, len(bean.Conditions())).Equal(1)
	})

	t.Run("method - 3", func(t *testing.T) {
		parent := NewBean(&TestBean{})
		bean := NewBean((*TestBean).Clone, gs_arg.Index(0, parent))
		assert.That(t, bean.GetName()).Equal("Clone")
		assert.That(t, len(bean.Conditions())).Equal(1)
	})

	t.Run("method - 4", func(t *testing.T) {
		parent := NewBean(&TestBean{})
		bean := NewBean((*TestBean).Clone, parent)
		assert.That(t, bean.GetName()).Equal("Clone")
		assert.That(t, len(bean.Conditions())).Equal(1)
	})

	t.Run("method - 5", func(t *testing.T) {
		parent := NewBean(&TestBean{})
		bean := NewBean((*TestBean).Clone, gs_arg.Index(0, parent))
		assert.That(t, bean.GetName()).Equal("Clone")
		assert.That(t, len(bean.Conditions())).Equal(1)
	})

	t.Run("method error - 1", func(t *testing.T) {
		assert.Panic(t, func() {
			NewBean((*TestBean).Clone, gs_arg.Tag(""))
		}, "ctorArgs\\[0] should be \\*BeanDefinition or IndexArg\\[0]")
	})

	t.Run("method error - 2", func(t *testing.T) {
		assert.Panic(t, func() {
			NewBean((*TestBean).Clone, gs_arg.Index(0, gs_arg.Tag("")))
		}, "the arg of IndexArg\\[0] should be \\*BeanDefinition")
	})
}
