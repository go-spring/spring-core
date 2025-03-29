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

package gs_core_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/gs/internal/gs_core"
	pkg1 "github.com/go-spring/spring-core/gs/internal/gs_core/testdata/pkg/bar"
	pkg2 "github.com/go-spring/spring-core/gs/internal/gs_core/testdata/pkg/foo"
	"github.com/go-spring/spring-core/gs/internal/gs_dync"
	"github.com/go-spring/spring-core/util/assert"
	"github.com/spf13/cast"
)

func runTest(c *gs_core.Container, fn func(*gs_core.Container)) error {
	err := c.Refresh()
	if err != nil {
		return err
	}
	fn(c)
	return nil
}

func TestRegisterBeanFrozen(t *testing.T) {
	// assert.Panic(t, func() {
	// 	c := gs.New()
	// 	c.Object(func() {}).Init(func(f func()) {
	// 		c.Object(func() {}) // 不能在这里注册新的 Object
	// 	})
	// 	_ = c.Refresh()
	// }, "should call before Refresh")
}

func TestContext(t *testing.T) {

	// ///////////////////////////////////////
	// 自定义数据类型

	t.Run("pkg1.SamePkg", func(t *testing.T) {
		c := gs_core.New()
		e := pkg1.SamePkg{}

		// assert.Panic(t, func() {
		// 	c.Object(e)
		// }, "bean must be ref type")

		c.Object(&e)
		c.Object(&e).Name("i3")
		c.Object(&e).Name("i4")

		err := c.Refresh()
		assert.Nil(t, err)
	})

	t.Run("pkg2.SamePkg", func(t *testing.T) {
		c := gs_core.New()
		e := pkg2.SamePkg{}

		// assert.Panic(t, func() {
		// 	c.Object(e)
		// }, "bean must be ref type")

		c.Object(&e)
		c.Object(&e).Name("i3")
		c.Object(&e).Name("i4")
		c.Object(&e).Name("i5")

		err := c.Refresh()
		assert.Nil(t, err)
	})
}

type TestBincoreng struct {
	i int
}

func (b *TestBincoreng) String() string {
	if b == nil {
		return ""
	} else {
		return strconv.Itoa(b.i)
	}
}

type TestObject struct {

	// 自定义类型指针
	StructByType *TestBincoreng `inject:"?"`
	StructByName *TestBincoreng `autowire:"struct_ptr?"`

	// 自定义类型指针数组
	StructPtrSliceByType []*TestBincoreng `inject:"?"`
	StructPtrCollection  []*TestBincoreng `autowire:"?"`
	StructPtrSliceByName []*TestBincoreng `autowire:"struct_ptr_slice?"`

	// 接口
	InterfaceByType fmt.Stringer `inject:"?"`
	InterfaceByName fmt.Stringer `autowire:"struct_ptr?"`

	// 接口数组
	InterfaceSliceByType []fmt.Stringer `autowire:"?"`

	InterfaceCollection  []fmt.Stringer `inject:"?"`
	InterfaceCollection2 []fmt.Stringer `autowire:"?"`

	// 指定名称时使用精确匹配模式，不对数组元素进行转换，即便能做到似乎也无意义
	InterfaceSliceByName []fmt.Stringer `autowire:"struct_ptr_slice?"`

	MapTyType map[string]fmt.Stringer `inject:"?"`
	MapByName map[string]fmt.Stringer `autowire:"map?"`
	MapByNam2 map[string]fmt.Stringer `autowire:"struct_ptr?"`
}

func TestAutoWireBeans(t *testing.T) {

	c := gs_core.New()

	obj := &TestObject{}
	c.Object(obj)

	b := TestBincoreng{1}
	c.Object(&b).Name("struct_ptr").Export(
		gs.As[fmt.Stringer](),
	)

	err := runTest(c, func(p *gs_core.Container) {})
	assert.Nil(t, err)

	assert.Equal(t, len(obj.MapTyType), 1)
	assert.Equal(t, len(obj.MapByName), 0)
	assert.Equal(t, len(obj.MapByNam2), 1)
	fmt.Printf("%+v\n", obj)
}

type SubSubSetting struct {
	Int        int `value:"${int}"`
	DefaultInt int `value:"${default.int:=2}"`
}

type SubSetting struct {
	Int        int `value:"${int}"`
	DefaultInt int `value:"${default.int:=2}"`

	SubSubSetting SubSubSetting `value:"${sub}"`
}

type Setting struct {
	Int        int `value:"${int}"`
	DefaultInt int `value:"${default.int:=2}"`
	// IntPtr     *int `value:"${int}"` // 不支持指针

	Uint        uint `value:"${uint}"`
	DefaultUint uint `value:"${default.uint:=2}"`

	Float        float32 `value:"${float}"`
	DefaultFloat float32 `value:"${default.float:=2}"`

	// Complex complex64 `value:"${complex}"` // 不支持复数

	String        string `value:"${string}"`
	DefaultString string `value:"${default.string:=2}"`

	Bool        bool `value:"${bool}"`
	DefaultBool bool `value:"${default.bool:=false}"`

	SubSetting SubSetting `value:"${sub}"`
	// SubSettingPtr *SubSetting `value:"${sub}"` // 不支持指针

	SubSubSetting SubSubSetting `value:"${sub_sub}"`

	IntSlice    []int    `value:"${int_slice}"`
	StringSlice []string `value:"${string_slice}"`
	// FloatSlice  []float64 `value:"${float_slice}"`
}

func TestValueTag(t *testing.T) {
	c := gs_core.New()

	p := conf.New()
	p.Set("int", "3")
	p.Set("uint", "3")
	p.Set("float", "3")
	//p.Set("complex", complex(3, 0))
	p.Set("string", "3")
	p.Set("bool", "true")

	setting := &Setting{}
	c.Object(setting)

	p.Set("sub.int", "4")
	p.Set("sub.sub.int", "5")
	p.Set("sub_sub.int", "6")

	p.Set("int_slice[0]", "1")
	p.Set("int_slice[1]", "2")
	p.Set("string_slice[0]", "1")
	p.Set("string_slice[1]", "2")
	// c.Property("float_slice", []float64{1, 2})

	err := c.RefreshProperties(p)
	assert.Nil(t, err)

	err = c.Refresh()
	assert.Nil(t, err)

	fmt.Printf("%+v\n", setting)
}

type GreetingService struct {
}

func (gs *GreetingService) Greeting(name string) string {
	return "hello " + name
}

type PrototypeBean struct {
	Service *GreetingService `autowire:""`
	name    string
	t       time.Time
}

func (p *PrototypeBean) Greeting() string {
	return p.t.Format("15:04:05.000") + " " + p.Service.Greeting(p.name)
}

type PrototypeBeanFactory struct {
	PrototypeBean
}

func (f *PrototypeBeanFactory) New(name string) *PrototypeBean {
	b := f.PrototypeBean
	b.name = name
	b.t = time.Now()
	return &b
}

type PrototypeBeanService struct {
	Provide *PrototypeBeanFactory `autowire:""`
}

func (s *PrototypeBeanService) Service(name string) {
	// 通过 PrototypeBean 的工厂获取新的实例，并且每个实例都有自己的时间戳
	fmt.Println(s.Provide.New(name).Greeting())
}

func TestPrototypeBean(t *testing.T) {
	c := gs_core.New()

	greetingService := &GreetingService{}
	c.Object(greetingService)

	s := &PrototypeBeanService{}
	c.Object(s)

	f := &PrototypeBeanFactory{}
	c.Object(f)

	err := c.Refresh()
	assert.Nil(t, err)

	s.Service("Li Lei")
	time.Sleep(50 * time.Millisecond)

	s.Service("Jim Green")
	time.Sleep(50 * time.Millisecond)

	s.Service("Han MeiMei")
}

type EnvEnum string

const ENV_TEST EnvEnum = "test"

type EnvEnumBean struct {
	EnvType EnvEnum `value:"${env.type}"`
}

type PointBean struct {
	Point        image.Point   `value:"${point}"`
	DefaultPoint image.Point   `value:"${default_point:=(3,4)}"`
	PointList    []image.Point `value:"${point_list}"`
}

func PointConverter(val string) (image.Point, error) {
	if !(strings.HasPrefix(val, "(") && strings.HasSuffix(val, ")")) {
		return image.Point{}, errors.New("数据格式错误")
	}
	ss := strings.Split(val[1:len(val)-1], ",")
	x := cast.ToInt(ss[0])
	y := cast.ToInt(ss[1])
	return image.Point{X: x, Y: y}, nil
}

type DB struct {
	UserName string `value:"${username}"`
	Password string `value:"${password}"`
	Url      string `value:"${url}"`
	Port     string `value:"${port}"`
	DB       string `value:"${db}"`
}

type DbConfig struct {
	DB []DB `value:"${db}"`
}

func TestTypeConverter(t *testing.T) {
	prop, _ := conf.Load("testdata/config/app.yaml")

	c := gs_core.New()

	b := &EnvEnumBean{}
	c.Object(b)

	prop.Set("env.type", "test")

	p := &PointBean{}
	c.Object(p)

	conf.RegisterConverter(PointConverter)
	prop.Set("point", "(7,5)")

	dbConfig := &DbConfig{}
	c.Object(dbConfig)

	err := c.RefreshProperties(prop)
	assert.Nil(t, err)

	err = c.Refresh()
	assert.Nil(t, err)

	assert.Equal(t, b.EnvType, ENV_TEST)

	fmt.Printf("%+v\n", b)
	fmt.Printf("%+v\n", p)

	fmt.Printf("%+v\n", dbConfig)
}

type Grouper interface {
	Group()
}

type MyGrouper struct {
}

func (g *MyGrouper) Group() {

}

type ProxyGrouper struct {
	Grouper `autowire:""`
}

func TestNestedBean(t *testing.T) {
	c := gs_core.New()
	c.Object(new(MyGrouper)).Export(
		gs.As[Grouper](),
	)
	c.Object(new(ProxyGrouper))
	err := c.Refresh()
	assert.Nil(t, err)
}

type BeanZero struct {
	Int int
}

type BeanOne struct {
	Zero *BeanZero `autowire:""`
}

type BeanTwo struct {
	One *BeanOne `autowire:""`
}

func (t *BeanTwo) Group() {
}

type BeanThree struct {
	One *BeanTwo `autowire:""`
}

func (t *BeanThree) String() string {
	return ""
}

// func TestGet(t *testing.T) {
//
// 	t.Run("panic", func(t *testing.T) {
// 		c := gs_core.New()
// 		err := runTest(c, func(p *gs_core.Container) {
// 			{
// 				var s fmt.Stringer
// 				err := p.Wire(s)
// 				assert.Error(t, err, "i can't be nil")
// 			}
// 			{
// 				var s fmt.Stringer
// 				err := p.Wire(&s)
// 				assert.Error(t, err, "can't find bean, bean:\"\"")
// 			}
// 		})
// 		assert.Nil(t, err)
// 	})
//
// 	t.Run("success", func(t *testing.T) {
// 		c := gs_core.New()
// 		c.Object(&BeanZero{5})
// 		c.Object(new(BeanOne))
// 		c.Object(new(BeanTwo)).Export(
// 			gs.As[Grouper](),
// 		)
// 		err := runTest(c, func(p *gs_core.Container) {
//
// 			var two *BeanTwo
// 			err := p.Wire(&two)
// 			assert.Nil(t, err)
//
// 			var grouper Grouper
// 			err = p.Wire(&grouper)
// 			assert.Nil(t, err)
//
// 			err = p.Wire(&two)
// 			assert.Nil(t, err)
//
// 			err = p.Wire(&grouper)
// 			assert.Nil(t, err)
//
// 			err = p.Wire(&two, "BeanTwo")
// 			assert.Nil(t, err)
//
// 			err = p.Wire(&grouper, "BeanTwo")
// 			assert.Nil(t, err)
//
// 			var three *BeanThree
// 			err = p.Wire(&three)
// 			assert.Error(t, err, "can't find bean, bean:\"\"")
// 		})
// 		assert.Nil(t, err)
// 	})
// }

// func TestFindByName(t *testing.T) {
//
//	c := runTest(c,func(p gs.Context) {})
//	c.Object(&BeanZero{5})
//	c.Object(new(BeanOne))
//	c.Object(new(BeanTwo))
//	c.Refresh()
//
//	p := <-ch
//
//	b, _ := p.Find("")
//	assert.Equal(t, len(b), 4)
//
//	b, _ = p.Find("BeanTwo")
//	fmt.Println(util.ToJsonString(b))
//	assert.Equal(t, len(b), 0)
//
//	b, _ = p.Find("BeanTwo")
//	fmt.Println(util.ToJsonString(b))
//	assert.Equal(t, len(b), 1)
//
//	b, _ = p.Find(":BeanTwo")
//	fmt.Println(util.ToJsonString(b))
//	assert.Equal(t, len(b), 1)
//
//	b, _ = p.Find("github.com/go-spring/spring-core/gs/gs_test.BeanTwo:BeanTwo")
//	fmt.Println(util.ToJsonString(b))
//	assert.Equal(t, len(b), 1)
//
//	b, _ = p.Find("xxx:BeanTwo")
//	fmt.Println(util.ToJsonString(b))
//	assert.Equal(t, len(b), 0)
//
//	b, _ = p.Find((*BeanTwo)(nil))
//	fmt.Println(util.ToJsonString(b))
//	assert.Equal(t, len(b), 1)
//
//	b, _ = p.Find((*fmt.Stringer)(nil))
//	assert.Equal(t, len(b), 0)
//
//	b, _ = p.Find((*Grouper)(nil))
//	assert.Equal(t, len(b), 0)
// }

type Teacher interface {
	Course() string
}

type historyTeacher struct {
	name string
}

func newHistoryTeacher(name string) *historyTeacher {
	return &historyTeacher{name: name}
}

func newTeacher(course string, name string) Teacher {
	switch course {
	case "history":
		return &historyTeacher{name: name}
	default:
		return nil
	}
}

func (t *historyTeacher) Course() string {
	return "history"
}

type Student struct {
	Teacher Teacher
	Room    string
}

// 入参可以进行注入或者属性绑定，返回值可以是 struct、map、slice、func 等。
func NewStudent(teacher Teacher, room string) Student {
	return Student{
		Teacher: teacher,
		Room:    room,
	}
}

// 入参可以进行注入或者属性绑定，返回值可以是 struct、map、slice、func 等。
func NewPtrStudent(teacher Teacher, room string) *Student {
	return &Student{
		Teacher: teacher,
		Room:    room,
	}
}

func TestRegisterBeanFn(t *testing.T) {
	prop := conf.New()
	prop.Set("room", "Class 3 Grade 1")

	c := gs_core.New()

	// 用接口注册时实际使用的是原始类型
	c.Object(Teacher(newHistoryTeacher(""))).Export(
		gs.As[Teacher](),
	)

	c.Provide(NewStudent, gs_arg.Tag(""), gs_arg.Tag("${room}")).Name("st1")
	c.Provide(NewPtrStudent, gs_arg.Tag(""), gs_arg.Tag("${room}")).Name("st2")
	c.Provide(NewStudent, gs_arg.Tag("?"), gs_arg.Tag("${room:=https://}")).Name("st3")
	c.Provide(NewPtrStudent, gs_arg.Tag("?"), gs_arg.Tag("${room:=4567}")).Name("st4")

	c.Object(newTeacher("history", "")).Init(func(teacher Teacher) {
		fmt.Println(teacher.Course())
	}).Name("newTeacher")

	c.RefreshProperties(prop)
	err := runTest(c, func(p *gs_core.Container) {

		var st1 struct {
			Value *Student `autowire:"st1"`
		}
		err := p.Wire(&st1)
		_ = err

		// assert.Nil(t, err)
		// fmt.Println("::", cast.ToString(st1))
		// assert.Equal(t, st1.Room, p.Prop("room"))

		var st2 struct {
			Value *Student `autowire:"st2"`
		}
		err = p.Wire(&st2)

		// assert.Nil(t, err)
		// fmt.Println("::", cast.ToString(st2))
		// assert.Equal(t, st2.Room, p.Prop("room"))

		assert.NotSame(t, st1.Value, st2.Value)

		var st3 struct {
			Value *Student `autowire:"st3"`
		}
		err = p.Wire(&st3)

		// assert.Nil(t, err)
		// fmt.Println("::", cast.ToString(st3))
		// assert.Equal(t, st3.Room, p.Prop("room"))

		var st4 struct {
			Value *Student `autowire:"st4"`
		}
		err = p.Wire(&st4)

		// assert.Nil(t, err)
		// fmt.Println("::", cast.ToString(st4))
		// assert.Equal(t, st4.Room, p.Prop("room"))
	})
	assert.Nil(t, err)
}

func TestProfile(t *testing.T) {

	t.Run("bean:_c:", func(t *testing.T) {
		c := gs_core.New()
		c.Object(&BeanZero{5})
		err := runTest(c, func(p *gs_core.Container) {
			var b struct {
				Value *BeanZero `autowire:""`
			}
			err := p.Wire(&b)
			assert.Nil(t, err)
		})
		assert.Nil(t, err)
	})

	t.Run("bean:_c:test", func(t *testing.T) {
		prop := conf.New()
		prop.Set("spring.profiles.active", "test")

		c := gs_core.New()
		c.Object(&BeanZero{5})

		c.RefreshProperties(prop)
		err := runTest(c, func(p *gs_core.Container) {
			var b struct {
				Value *BeanZero `autowire:""`
			}
			err := p.Wire(&b)
			assert.Nil(t, err)
		})
		assert.Nil(t, err)
	})
}

type BeanFour struct{}

func TestDependsOn(t *testing.T) {

	t.Run("random", func(t *testing.T) {
		c := gs_core.New()
		c.Object(&BeanZero{5})
		c.Object(new(BeanOne))
		c.Object(new(BeanFour))
		err := c.Refresh()
		assert.Nil(t, err)
	})

	t.Run("dependsOn", func(t *testing.T) {
		c := gs_core.New()
		c.Object(&BeanZero{5})
		c.Object(new(BeanOne))
		c.Object(new(BeanFour)).DependsOn(
			gs.BeanSelectorFor[*BeanOne](),
			gs.BeanSelectorImpl{Name: "BeanZero"},
		)
		err := c.Refresh()
		assert.Nil(t, err)
	})
}

func TestDuplicate(t *testing.T) {
	t.Run("duplicate", func(t *testing.T) {
		c := gs_core.New()
		c.Object(&BeanZero{5})
		c.Object(&BeanZero{6})
		c.Object(new(BeanOne))
		c.Object(new(BeanTwo))
		err := c.Refresh()
		assert.Error(t, err, "duplicate beans ")
	})
}

type FuncObj struct {
	Fn func(int) int `autowire:""`
}

func TestDefaultProperties_WireFunc(t *testing.T) {
	c := gs_core.New()
	c.Object(func(int) int { return 6 })
	obj := new(FuncObj)
	c.Object(obj)
	err := c.Refresh()
	assert.Nil(t, err)
	i := obj.Fn(3)
	assert.Equal(t, i, 6)
}

type Manager interface {
	Cluster() string
}

func NewManager() Manager {
	return localManager{}
}

func NewManagerRetError() (Manager, error) {
	return localManager{}, errors.New("NewManagerRetError error")
}

func NewManagerRetErrorNil() (Manager, error) {
	return localManager{}, nil
}

func NewNullPtrManager() Manager {
	return nil
}

func NewPtrManager() Manager {
	return &localManager{}
}

type localManager struct {
	Version string `value:"${manager.version}"`
}

func (m localManager) Cluster() string {
	return "local"
}

func TestRegisterBeanFn2(t *testing.T) {

	t.Run("ptr manager", func(t *testing.T) {
		prop := conf.New()
		prop.Set("manager.version", "1.0.0")

		c := gs_core.New()
		c.Provide(NewPtrManager)

		c.RefreshProperties(prop)
		err := runTest(c, func(p *gs_core.Container) {

			var m struct {
				Value Manager `autowire:""`
			}
			err := p.Wire(&m)
			assert.Nil(t, err)

			// 因为用户是按照接口注册的，所以理论上在依赖
			// 系统中用户并不关心接口对应的真实类型是什么。
			var lm struct {
				Value *localManager `autowire:""`
			}
			err = p.Wire(&lm)
			assert.Error(t, err, "can't find bean, bean:\"\"")
		})
		assert.Nil(t, err)
	})

	t.Run("manager", func(t *testing.T) {
		prop := conf.New()
		prop.Set("manager.version", "1.0.0")

		c := gs_core.New()
		c.RefreshProperties(prop)

		c.Provide(NewManager)
		err := runTest(c, func(p *gs_core.Container) {

			var m struct {
				Value Manager `autowire:""`
			}
			err := p.Wire(&m)
			assert.Nil(t, err)

			var lm struct {
				Value *localManager `autowire:""`
			}
			err = p.Wire(&lm)
			assert.Error(t, err, "can't find bean, bean:\"\"")
		})
		assert.Nil(t, err)
	})

	t.Run("manager return error", func(t *testing.T) {
		prop := conf.New()
		prop.Set("manager.version", "1.0.0")
		c := gs_core.New()
		c.Provide(NewManagerRetError)
		c.RefreshProperties(prop)
		err := c.Refresh()
		assert.Error(t, err, "NewManagerRetError error")
	})

	t.Run("manager return error nil", func(t *testing.T) {
		prop := conf.New()
		prop.Set("manager.version", "1.0.0")

		c := gs_core.New()
		c.Provide(NewManagerRetErrorNil)

		c.RefreshProperties(prop)
		err := c.Refresh()
		assert.Nil(t, err)
	})

	t.Run("manager return nil", func(t *testing.T) {
		prop := conf.New()
		prop.Set("manager.version", "1.0.0")

		c := gs_core.New()
		c.Provide(NewNullPtrManager)

		c.RefreshProperties(prop)
		err := c.Refresh()
		assert.Error(t, err, "return nil")
	})
}

type destroyable interface {
	Init()
	Destroy()
	InitWithError() error
	DestroyWithError() error
}

type callDestroy struct {
	i         int
	inited    bool
	destroyed bool
}

func (d *callDestroy) Init() {
	d.inited = true
}

func (d *callDestroy) Destroy() {
	d.destroyed = true
}

func (d *callDestroy) InitWithError() error {
	if d.i == 0 {
		d.inited = true
		return nil
	}
	return errors.New("InitWithError error")
}

func (d *callDestroy) DestroyWithError() error {
	if d.i == 0 {
		d.destroyed = true
		return nil
	}
	return errors.New("DestroyWithError error")
}

type nestedCallDestroy struct {
	callDestroy
}

type nestedDestroyable struct {
	destroyable
}

func TestRegisterBean_InitFunc(t *testing.T) {

	t.Run("call init", func(t *testing.T) {

		c := gs_core.New()
		c.Object(new(callDestroy)).Init((*callDestroy).Init)
		err := runTest(c, func(p *gs_core.Container) {
			var d struct {
				Value *callDestroy `autowire:""`
			}
			err := p.Wire(&d)
			assert.Nil(t, err)
			assert.True(t, d.Value.inited)
		})
		assert.Nil(t, err)
	})

	t.Run("call init with error", func(t *testing.T) {

		{
			c := gs_core.New()
			c.Object(&callDestroy{i: 1}).Init((*callDestroy).InitWithError)
			err := c.Refresh()
			assert.Error(t, err, "error")
		}

		prop := conf.New()
		prop.Set("int", "0")

		c := gs_core.New()
		c.Object(&callDestroy{}).Init((*callDestroy).InitWithError)

		c.RefreshProperties(prop)
		err := runTest(c, func(p *gs_core.Container) {
			var d struct {
				Value *callDestroy `autowire:""`
			}
			err := p.Wire(&d)
			assert.Nil(t, err)
			assert.True(t, d.Value.inited)
		})
		assert.Nil(t, err)
	})

	t.Run("call interface init", func(t *testing.T) {
		c := gs_core.New()
		c.Provide(func() destroyable { return new(callDestroy) }).Init(destroyable.Init)
		err := runTest(c, func(p *gs_core.Container) {
			var d struct {
				Value destroyable `autowire:""`
			}
			err := p.Wire(&d)
			assert.Nil(t, err)
			assert.True(t, d.Value.(*callDestroy).inited)
		})
		assert.Nil(t, err)
	})

	t.Run("call interface init with error", func(t *testing.T) {

		{
			c := gs_core.New()
			c.Provide(func() destroyable { return &callDestroy{i: 1} }).Init(destroyable.InitWithError)
			err := c.Refresh()
			assert.Error(t, err, "error")
		}

		prop := conf.New()
		prop.Set("int", "0")

		c := gs_core.New()
		c.Provide(func() destroyable { return &callDestroy{} }).Init(destroyable.InitWithError)

		c.RefreshProperties(prop)
		err := runTest(c, func(p *gs_core.Container) {
			var d struct {
				Value destroyable `autowire:""`
			}
			err := p.Wire(&d)
			assert.Nil(t, err)
			assert.True(t, d.Value.(*callDestroy).inited)
		})
		assert.Nil(t, err)
	})

	t.Run("call nested init", func(t *testing.T) {
		c := gs_core.New()
		c.Object(new(nestedCallDestroy)).Init((*nestedCallDestroy).Init)
		err := runTest(c, func(p *gs_core.Container) {
			var d struct {
				Value *nestedCallDestroy `autowire:""`
			}
			err := p.Wire(&d)
			assert.Nil(t, err)
			assert.True(t, d.Value.inited)
		})
		assert.Nil(t, err)
	})

	t.Run("call nested interface init", func(t *testing.T) {
		c := gs_core.New()
		c.Object(&nestedDestroyable{
			destroyable: new(callDestroy),
		}).Init((*nestedDestroyable).Init)
		err := runTest(c, func(p *gs_core.Container) {
			var d struct {
				Value *nestedDestroyable `autowire:""`
			}
			err := p.Wire(&d)
			assert.Nil(t, err)
			assert.True(t, d.Value.destroyable.(*callDestroy).inited)
		})
		assert.Nil(t, err)
	})
}

type RecoresCluster struct {
	Endpoints string `value:"${redis.endpoints}"`
}

func TestValueBincoreng(t *testing.T) {
	prop := conf.New()
	prop.Set("redis.endpoints", "redis://localhost:6379")

	c := gs_core.New()
	c.Object(new(RecoresCluster))

	c.RefreshProperties(prop)
	err := runTest(c, func(p *gs_core.Container) {
		var cluster struct {
			Value *RecoresCluster `autowire:""`
		}
		err := p.Wire(&cluster)
		fmt.Println(cluster)
		assert.Nil(t, err)
	})
	assert.Nil(t, err)
}

func TestCollect(t *testing.T) {

	t.Run("", func(t *testing.T) {
		c := gs_core.New()
		c.Object(&struct {
			Events []ServerInterface `autowire:""`
		}{})
		err := runTest(c, func(ctx *gs_core.Container) {})
		assert.Error(t, err, "no beans collected for \"\"")
	})

	t.Run("", func(t *testing.T) {
		c := gs_core.New()
		err := runTest(c, func(ctx *gs_core.Container) {
			var Events struct {
				Events []ServerInterface `autowire:""`
			}
			err := ctx.Wire(&Events)
			assert.Error(t, err, "no beans collected for \"\"")
		})
		assert.Nil(t, err)
	})

	t.Run("", func(t *testing.T) {
		c := gs_core.New()
		c.Object(&struct {
			Events []ServerInterface `autowire:"?"`
		}{})
		err := runTest(c, func(ctx *gs_core.Container) {})
		assert.Nil(t, err)
	})

	t.Run("", func(t *testing.T) {
		c := gs_core.New()
		err := runTest(c, func(ctx *gs_core.Container) {
			var Events struct {
				Events []ServerInterface `autowire:"?"`
			}
			err := ctx.Wire(&Events)
			assert.Nil(t, err)
		})
		assert.Nil(t, err)
	})

	t.Run("", func(t *testing.T) {
		prop := conf.New()
		prop.Set("redis.endpoints", "redis://localhost:6379")

		c := gs_core.New()
		c.Object(new(RecoresCluster)).Name("one")
		c.Object(new(RecoresCluster))

		c.RefreshProperties(prop)
		err := runTest(c, func(p *gs_core.Container) {
			var rcs struct {
				Value []*RecoresCluster `autowire:""`
			}
			err := p.Wire(&rcs)
			assert.Nil(t, err)
			assert.Equal(t, len(rcs.Value), 2)
		})
		assert.Nil(t, err)
	})
}

var defaultClassOption = ClassOption{
	className: "default",
}

type ClassBuilder struct {
	param string
}

type ClassOption struct {
	className string
	students  []*Student
	floor     int
	builder   *ClassBuilder
}

type ClassOptionFunc func(opt *ClassOption)

func withClassName(className string, floor int) ClassOptionFunc {
	return func(opt *ClassOption) {
		opt.className = className
		opt.floor = floor
	}
}

func withStudents(students []*Student) ClassOptionFunc {
	return func(opt *ClassOption) {
		opt.students = students
	}
}

func withBuilder(builder *ClassBuilder) ClassOptionFunc {
	return func(opt *ClassOption) {
		opt.builder = builder
	}
}

type ClassRoom struct {
	President string `value:"${president}"`
	className string
	floor     int
	students  []*Student
	desktop   Desktop
	builder   *ClassBuilder
}

type Desktop interface {
}

type MetalDesktop struct {
}

func (cls *ClassRoom) Desktop() Desktop {
	return cls.desktop
}

func NewClassRoom(options ...ClassOptionFunc) ClassRoom {
	opt := defaultClassOption
	for _, fn := range options {
		fn(&opt)
	}
	return ClassRoom{
		className: opt.className,
		students:  opt.students,
		floor:     opt.floor,
		desktop:   &MetalDesktop{},
		builder:   opt.builder,
	}
}

func TestOptionPattern(t *testing.T) {

	students := []*Student{
		new(Student), new(Student),
	}

	cls := NewClassRoom()
	assert.Equal(t, cls.className, "default")

	cls = NewClassRoom(withClassName("二年级03班", 3))
	assert.Equal(t, cls.floor, 3)
	assert.Equal(t, len(cls.students), 0)
	assert.Equal(t, cls.className, "二年级03班")

	cls = NewClassRoom(withStudents(students))
	assert.Equal(t, cls.floor, 0)
	assert.Equal(t, cls.students, students)
	assert.Equal(t, cls.className, "default")

	cls = NewClassRoom(withClassName("二年级03班", 3), withStudents(students))
	assert.Equal(t, cls.className, "二年级03班")
	assert.Equal(t, cls.students, students)
	assert.Equal(t, cls.floor, 3)

	cls = NewClassRoom(withStudents(students), withClassName("二年级03班", 3))
	assert.Equal(t, cls.className, "二年级03班")
	assert.Equal(t, cls.students, students)
	assert.Equal(t, cls.floor, 3)
}

func TestOptionConstructorArg(t *testing.T) {

	t.Run("option default", func(t *testing.T) {
		prop := conf.New()
		prop.Set("president", "CaiYuanPei")

		c := gs_core.New()
		c.Provide(NewClassRoom)

		c.RefreshProperties(prop)
		err := runTest(c, func(p *gs_core.Container) {
			var cls struct {
				Value *ClassRoom `autowire:""`
			}
			err := p.Wire(&cls)
			assert.Nil(t, err)
			assert.Equal(t, len(cls.Value.students), 0)
			assert.Equal(t, cls.Value.className, "default")
			assert.Equal(t, cls.Value.President, "CaiYuanPei")
		})
		assert.Nil(t, err)
	})

	t.Run("option withClassName", func(t *testing.T) {
		prop := conf.New()
		prop.Set("president", "CaiYuanPei")

		c := gs_core.New()
		c.Provide(NewClassRoom, gs_arg.Bind(withClassName, gs_arg.Tag("${class_name:=二年级03班}"), gs_arg.Tag("${class_floor:=3}")))

		c.RefreshProperties(prop)
		err := runTest(c, func(p *gs_core.Container) {
			var cls struct {
				Value *ClassRoom `autowire:""`
			}
			err := p.Wire(&cls)
			assert.Nil(t, err)
			assert.Equal(t, cls.Value.floor, 3)
			assert.Equal(t, len(cls.Value.students), 0)
			assert.Equal(t, cls.Value.className, "二年级03班")
			assert.Equal(t, cls.Value.President, "CaiYuanPei")
		})
		assert.Nil(t, err)
	})

	t.Run("option withStudents", func(t *testing.T) {
		prop := conf.New()
		prop.Set("class_name", "二年级03班")
		prop.Set("president", "CaiYuanPei")

		c := gs_core.New()
		c.Provide(NewClassRoom, gs_arg.Bind(withStudents))
		c.Object(new(Student)).Name("Student1")
		c.Object(new(Student)).Name("Student2")

		c.RefreshProperties(prop)
		err := runTest(c, func(p *gs_core.Container) {
			var cls struct {
				Value *ClassRoom `autowire:""`
			}
			err := p.Wire(&cls)
			assert.Nil(t, err)
			assert.Equal(t, cls.Value.floor, 0)
			assert.Equal(t, len(cls.Value.students), 2)
			assert.Equal(t, cls.Value.className, "default")
			assert.Equal(t, cls.Value.President, "CaiYuanPei")
		})
		assert.Nil(t, err)
	})

	// t.Run("option withStudents withClassName", func(t *testing.T) {
	//	prop := conf.New()
	//	prop.Set("class_name", "二年级06班")
	//	prop.Set("president", "CaiYuanPei")
	//
	//	c := gs_core.New()
	//	c.Provide(NewClassRoom,
	//		gs_arg.Bind(withStudents),
	//		gs_arg.Bind(withClassName, gs_arg.Tag("${class_name:=二年级03班}"), gs_arg.Tag("${class_floor:=3}")),
	//		gs_arg.Bind(withBuilder, gs_arg.MustBind(func(param string) *ClassBuilder {
	//			return &ClassBuilder{param: param}
	//		}, gs_arg.Value("1"))),
	//	)
	//	c.Object(&Student{}).Name("Student1")
	//	c.Object(&Student{}).Name("Student2")
	//
	//	c.RefreshProperties(prop)
	//	err := runTest(c, func(p *gs_core.Container) {
	//		var cls struct {
	//			Value *ClassRoom `autowire:""`
	//		}
	//		err := p.Wire(&cls)
	//		assert.Nil(t, err)
	//		assert.Equal(t, cls.Value.floor, 3)
	//		assert.Equal(t, len(cls.Value.students), 2)
	//		assert.Equal(t, cls.Value.className, "二年级06班")
	//		assert.Equal(t, cls.Value.President, "CaiYuanPei")
	//		assert.Equal(t, cls.Value.builder.param, "1")
	//	})
	//	assert.Nil(t, err)
	// })
}

type ServerInterface interface {
	Consumer() *Consumer
	ConsumerT() *Consumer
	ConsumerArg(i int) *Consumer
}

type Server struct {
	Version string `value:"${server.version}"`
}

func NewServerInterface() ServerInterface {
	return new(Server)
}

type Consumer struct {
	s *Server
}

func (s *Server) Consumer() *Consumer {
	if nil == s {
		panic("server is nil")
	}
	return &Consumer{s}
}

func (s *Server) ConsumerT() *Consumer {
	return s.Consumer()
}

func (s *Server) ConsumerArg(_ int) *Consumer {
	if nil == s {
		panic("server is nil")
	}
	return &Consumer{s}
}

type Service struct {
	Consumer *Consumer `autowire:""`
}

func TestRegisterMethodBean(t *testing.T) {

	t.Run("method bean", func(t *testing.T) {
		prop := conf.New()
		prop.Set("server.version", "1.0.0")

		c := gs_core.New()
		parent := c.Object(new(Server))
		c.Provide((*Server).Consumer, parent)

		c.RefreshProperties(prop)
		err := runTest(c, func(p *gs_core.Container) {

			var s struct {
				Value *Server `autowire:""`
			}
			err := p.Wire(&s)
			assert.Nil(t, err)
			assert.Equal(t, s.Value.Version, "1.0.0")

			s.Value.Version = "2.0.0"

			var consumer struct {
				Value *Consumer `autowire:""`
			}
			err = p.Wire(&consumer)
			assert.Nil(t, err)
			assert.Equal(t, consumer.Value.s.Version, "2.0.0")
		})
		assert.Nil(t, err)
	})

	t.Run("method bean condition", func(t *testing.T) {
		prop := conf.New()
		prop.Set("server.version", "1.0.0")

		c := gs_core.New()
		parent := c.Object(new(Server)).Condition(gs_cond.Not(gs_cond.OnMissingProperty("a")))
		c.Provide((*Server).Consumer, parent)

		c.RefreshProperties(prop)
		err := runTest(c, func(p *gs_core.Container) {

			var s struct {
				Value *Server `autowire:""`
			}
			err := p.Wire(&s)
			assert.Error(t, err, "can't find bean, bean:\"\" type:\"\\*gs_core_test.Server\"")

			var consumer struct {
				Value *Consumer `autowire:""`
			}
			err = p.Wire(&consumer)
			assert.Error(t, err, "can't find bean, bean:\"\" type:\"\\*gs_core_test.Consumer\"")
		})
		assert.Nil(t, err)
	})

	t.Run("method bean arg", func(t *testing.T) {
		prop := conf.New()
		prop.Set("server.version", "1.0.0")

		c := gs_core.New()
		parent := c.Object(new(Server))
		c.Provide((*Server).ConsumerArg, parent, gs_arg.Tag("${i:=9}"))

		c.RefreshProperties(prop)
		err := runTest(c, func(p *gs_core.Container) {

			var s struct {
				Value *Server `autowire:""`
			}
			err := p.Wire(&s)
			assert.Nil(t, err)
			assert.Equal(t, s.Value.Version, "1.0.0")

			s.Value.Version = "2.0.0"

			var consumer struct {
				Value *Consumer `autowire:""`
			}
			err = p.Wire(&consumer)
			assert.Nil(t, err)
			assert.Equal(t, consumer.Value.s.Version, "2.0.0")
		})
		assert.Nil(t, err)
	})

	t.Run("method bean wire to other bean", func(t *testing.T) {
		prop := conf.New()
		prop.Set("server.version", "1.0.0")

		c := gs_core.New()
		parent := c.Provide(NewServerInterface)
		c.Provide(ServerInterface.Consumer, parent).DependsOn(gs.BeanSelectorImpl{Name: "ServerInterface"})
		c.Object(new(Service))

		c.RefreshProperties(prop)
		err := runTest(c, func(p *gs_core.Container) {

			var si struct {
				Value ServerInterface `autowire:""`
			}
			err := p.Wire(&si)
			assert.Nil(t, err)

			s := si.Value.(*Server)
			assert.Equal(t, s.Version, "1.0.0")

			s.Version = "2.0.0"

			var consumer struct {
				Value *Consumer `autowire:""`
			}
			err = p.Wire(&consumer)
			assert.Nil(t, err)
			assert.Equal(t, consumer.Value.s.Version, "2.0.0")
		})
		assert.Nil(t, err)
	})

	t.Run("circle autowire", func(t *testing.T) {
		okCount := 0
		errCount := 0
		for i := 0; i < 20; i++ { // 不要排序
			func() {

				defer func() {
					if err := recover(); err != nil {
						errCount++

						var v string
						switch e := err.(type) {
						case error:
							v = e.Error()
						case string:
							v = e
						}

						if !strings.Contains(v, "found circle autowire") {
							panic("test error")
						}
					} else {
						okCount++
					}
				}()

				prop := conf.New()
				prop.Set("server.version", "1.0.0")

				c := gs_core.New()
				parent := c.Object(new(Server)).DependsOn(gs.BeanSelectorImpl{Name: "Service"})
				c.Provide((*Server).Consumer, parent).DependsOn(gs.BeanSelectorImpl{Name: "Service"})
				c.Object(new(Service))
				c.RefreshProperties(prop)
				err := c.Refresh()
				if err != nil {
					panic(err)
				}
			}()
		}
		fmt.Printf("ok:%d err:%d\n", okCount, errCount)
	})

	t.Run("method bean autowire", func(t *testing.T) {
		prop := conf.New()
		prop.Set("server.version", "1.0.0")

		c := gs_core.New()
		c.Object(new(Server))

		c.RefreshProperties(prop)
		err := runTest(c, func(p *gs_core.Container) {
			var s struct {
				Value *Server `autowire:""`
			}
			err := p.Wire(&s)
			assert.Nil(t, err)
			assert.Equal(t, s.Value.Version, "1.0.0")
		})
		assert.Nil(t, err)
	})

	t.Run("method bean selector type", func(t *testing.T) {
		prop := conf.New()
		prop.Set("server.version", "1.0.0")

		c := gs_core.New()
		c.Object(new(Server))
		c.Provide(func(s *Server) *Consumer { return s.Consumer() })

		c.RefreshProperties(prop)
		err := runTest(c, func(p *gs_core.Container) {

			var s struct {
				Value *Server `autowire:""`
			}
			err := p.Wire(&s)
			assert.Nil(t, err)
			assert.Equal(t, s.Value.Version, "1.0.0")

			s.Value.Version = "2.0.0"

			var consumer struct {
				Value *Consumer `autowire:""`
			}
			err = p.Wire(&consumer)
			assert.Nil(t, err)
			assert.Equal(t, consumer.Value.s.Version, "2.0.0")
		})
		assert.Nil(t, err)
	})

	// t.Run("method bean selector type error", func(t *testing.T) {
	// 	prop := conf.New()
	// 	prop.Set("server.version", "1.0.0")
	//
	// 	c := gs_core.New()
	// 	c.Object(new(Server))
	// 	c.Provide(func(s *Server) *Consumer { return s.Consumer() }) // gs_arg.BeanTag[int](),
	//
	// 	c.RefreshProperties(prop)
	// 	err := c.Refresh()
	// 	assert.Error(t, err, "can't find bean, bean:\"int:\" type:\"\\*gs_core_test.Server\"")
	// })

	t.Run("method bean selector beanId", func(t *testing.T) {
		prop := conf.New()
		prop.Set("server.version", "1.0.0")

		c := gs_core.New()
		c.Object(new(Server))
		c.Provide(func(s *Server) *Consumer { return s.Consumer() }, gs_arg.Tag("Server"))

		c.RefreshProperties(prop)
		err := runTest(c, func(p *gs_core.Container) {

			var s struct {
				Value *Server `autowire:""`
			}
			err := p.Wire(&s)
			assert.Nil(t, err)
			assert.Equal(t, s.Value.Version, "1.0.0")

			s.Value.Version = "2.0.0"

			var consumer struct {
				Value *Consumer `autowire:""`
			}
			err = p.Wire(&consumer)
			assert.Nil(t, err)
			assert.Equal(t, consumer.Value.s.Version, "2.0.0")
		})
		assert.Nil(t, err)
	})

	t.Run("method bean selector beanId error", func(t *testing.T) {
		prop := conf.New()
		prop.Set("server.version", "1.0.0")

		c := gs_core.New()
		c.Object(new(Server))
		c.Provide(func(s *Server) *Consumer { return s.Consumer() }, gs_arg.Tag("NULL"))

		c.RefreshProperties(prop)
		err := c.Refresh()
		assert.Error(t, err, "can't find bean, bean:\"NULL\" type:\"\\*gs_core_test.Server\"")
	})
}

func TestUserDefinedTypeProperty(t *testing.T) {

	type level int

	var config struct {
		Duration time.Duration `value:"${duration}"`
		Level    level         `value:"${level}"`
		Time     time.Time     `value:"${time}"`
		Complex  complex64     // `value:"${complex}"`
	}

	c := gs_core.New()

	conf.RegisterConverter(func(v string) (level, error) {
		if v == "debug" {
			return 1, nil
		}
		return 0, errors.New("error level")
	})

	prop := conf.New()
	prop.Set("time", "2018-12-20")
	prop.Set("duration", "1h")
	prop.Set("level", "debug")
	prop.Set("complex", "1+i")

	c.Object(&config)

	c.RefreshProperties(prop)
	err := c.Refresh()
	assert.Nil(t, err)

	fmt.Printf("%+v\n", config)
}

type CircleA struct {
	B *CircleB `autowire:""`
}

type CircleB struct {
	C *CircleC `autowire:""`
}

type CircleC struct {
	A *CircleA `autowire:""`
}

func TestCircleAutowire(t *testing.T) {

	// 直接创建的 Object 直接发生循环依赖是没有关系的。
	t.Run("", func(t *testing.T) {
		c := gs_core.New()
		c.Object(new(CircleA))
		c.Object(new(CircleB))
		c.Object(new(CircleC))
		err := c.Refresh()
		assert.Nil(t, err)
	})

	t.Run("", func(t *testing.T) {
		c := gs_core.New()
		c.Object(new(CircleA))
		c.Object(new(CircleB))
		c.Provide(func() *CircleC {
			return new(CircleC)
		})
		err := c.Refresh()
		assert.Nil(t, err)
	})

	t.Run("", func(t *testing.T) {
		c := gs_core.New()
		c.Object(new(CircleA))
		c.Provide(func() *CircleB {
			return new(CircleB)
		})
		c.Provide(func() *CircleC {
			return new(CircleC)
		})
		err := c.Refresh()
		assert.Nil(t, err)
	})

	t.Run("", func(t *testing.T) {
		c := gs_core.New()
		c.Provide(func(b *CircleB) *CircleA {
			return new(CircleA)
		})
		c.Provide(func(c *CircleC) *CircleB {
			return new(CircleB)
		})
		c.Provide(func(a *CircleA) *CircleC {
			return new(CircleC)
		})
		err := c.Refresh()
		assert.Error(t, err, "found circular autowire")
	})
}

type VarInterfaceOptionFunc func(opt *VarInterfaceOption)

type VarInterfaceOption struct {
	v []interface{}
}

func withVarInterface(v ...interface{}) VarInterfaceOptionFunc {
	return func(opt *VarInterfaceOption) {
		opt.v = v
	}
}

type VarInterfaceObj struct {
	v []interface{}
}

func NewVarInterfaceObj(options ...VarInterfaceOptionFunc) *VarInterfaceObj {
	opt := new(VarInterfaceOption)
	for _, option := range options {
		option(opt)
	}
	return &VarInterfaceObj{opt.v}
}

type Var struct {
	name string
}

type VarOption struct {
	v []*Var
}

type VarOptionFunc func(opt *VarOption)

func withVar(v ...*Var) VarOptionFunc {
	return func(opt *VarOption) {
		opt.v = v
	}
}

type VarObj struct {
	v []*Var
	s string
}

func NewVarObj(s string, options ...VarOptionFunc) *VarObj {
	opt := new(VarOption)
	for _, option := range options {
		option(opt)
	}
	return &VarObj{opt.v, s}
}

func NewNilVarObj(i interface{}, options ...VarOptionFunc) *VarObj {
	opt := new(VarOption)
	for _, option := range options {
		option(opt)
	}
	return &VarObj{opt.v, fmt.Sprint(i)}
}

func TestRegisterOptionBean(t *testing.T) {

	t.Run("nil param 0", func(t *testing.T) {
		prop := conf.New()
		prop.Set("var.obj", "description")

		c := gs_core.New()
		c.Object(&Var{"v1"}).Name("v1")
		c.Object(&Var{"v2"}).Name("v2")
		c.Provide(NewNilVarObj, gs_arg.Value(nil))

		c.RefreshProperties(prop)
		err := runTest(c, func(p *gs_core.Container) {
			var obj struct {
				Value *VarObj `autowire:""`
			}
			err := p.Wire(&obj)
			assert.Nil(t, err)
			assert.Equal(t, len(obj.Value.v), 0)
			assert.Equal(t, obj.Value.s, "<nil>")
		})
		assert.Nil(t, err)
	})

	t.Run("variable option param 1", func(t *testing.T) {
		prop := conf.New()
		prop.Set("var.obj", "description")

		c := gs_core.New()
		c.Object(&Var{"v1"}).Name("v1")
		c.Object(&Var{"v2"}).Name("v2")
		c.Provide(NewVarObj, gs_arg.Tag("${var.obj}"), gs_arg.Bind(withVar, gs_arg.Tag("v1")))

		c.RefreshProperties(prop)
		err := runTest(c, func(p *gs_core.Container) {
			var obj struct {
				Value *VarObj `autowire:""`
			}
			err := p.Wire(&obj)
			assert.Nil(t, err)
			assert.Equal(t, len(obj.Value.v), 1)
			assert.Equal(t, obj.Value.v[0].name, "v1")
			assert.Equal(t, obj.Value.s, "description")
		})
		assert.Nil(t, err)
	})

	t.Run("variable option param 2", func(t *testing.T) {
		prop := conf.New()
		prop.Set("var.obj", "description")

		c := gs_core.New()
		c.Object(&Var{"v1"}).Name("v1")
		c.Object(&Var{"v2"}).Name("v2")
		c.Provide(NewVarObj, gs_arg.Value("description"), gs_arg.Bind(withVar, gs_arg.Tag("v1"), gs_arg.Tag("v2")))

		c.RefreshProperties(prop)
		err := runTest(c, func(p *gs_core.Container) {
			var obj struct {
				Value *VarObj `autowire:""`
			}
			err := p.Wire(&obj)
			assert.Nil(t, err)
			assert.Equal(t, len(obj.Value.v), 2)
			assert.Equal(t, obj.Value.v[0].name, "v1")
			assert.Equal(t, obj.Value.v[1].name, "v2")
			assert.Equal(t, obj.Value.s, "description")
		})
		assert.Nil(t, err)
	})

	t.Run("variable option interface param 1", func(t *testing.T) {
		c := gs_core.New()
		c.Object(&Var{"v1"}).Name("v1").Export(
			gs.As[interface{}](),
		)
		c.Object(&Var{"v2"}).Name("v2").Export(
			gs.As[interface{}](),
		)
		c.Provide(NewVarInterfaceObj, gs_arg.Bind(withVarInterface, gs_arg.Tag("v1")))
		err := runTest(c, func(p *gs_core.Container) {
			var obj struct {
				Value *VarInterfaceObj `autowire:""`
			}
			err := p.Wire(&obj)
			assert.Nil(t, err)
			assert.Equal(t, len(obj.Value.v), 1)
		})
		assert.Nil(t, err)
	})

	t.Run("variable option interface param 1", func(t *testing.T) {
		c := gs_core.New()
		c.Object(&Var{"v1"}).Name("v1").Export(
			gs.As[interface{}](),
		)
		c.Object(&Var{"v2"}).Name("v2").Export(
			gs.As[interface{}](),
		)
		c.Provide(NewVarInterfaceObj, gs_arg.Bind(withVarInterface, gs_arg.Tag("v1"), gs_arg.Tag("v2")))
		err := runTest(c, func(p *gs_core.Container) {
			var obj struct {
				Value *VarInterfaceObj `autowire:""`
			}
			err := p.Wire(&obj)
			assert.Nil(t, err)
			assert.Equal(t, len(obj.Value.v), 2)
		})
		assert.Nil(t, err)
	})
}

func TestClose(t *testing.T) {

	// t.Run("destroy type", func(t *testing.T) {
	//
	// 	assert.Panic(t, func() {
	// 		c := gs.New()
	// 		c.Object(func() {}).Destroy(func() {})
	// 	}, "destroy should be func\\(bean\\) or func\\(bean\\)error")
	//
	// 	assert.Panic(t, func() {
	// 		c := gs.New()
	// 		c.Object(func() {}).Destroy(func() int { return 0 })
	// 	}, "destroy should be func\\(bean\\) or func\\(bean\\)error")
	//
	// 	assert.Panic(t, func() {
	// 		c := gs.New()
	// 		c.Object(func() {}).Destroy(func(int) {})
	// 	}, "destroy should be func\\(bean\\) or func\\(bean\\)error")
	//
	// 	assert.Panic(t, func() {
	// 		c := gs.New()
	// 		c.Object(func() {}).Destroy(func(int, int) {})
	// 	}, "destroy should be func\\(bean\\) or func\\(bean\\)error")
	// })

	t.Run("call destroy fn", func(t *testing.T) {
		called := false

		c := gs_core.New()
		c.Object(func() {}).Destroy(func(f func()) { called = true })
		err := c.Refresh()
		assert.Nil(t, err)
		c.Close()

		assert.True(t, called)
	})

	t.Run("call destroy", func(t *testing.T) {
		c := gs_core.New()
		d := new(callDestroy)
		c.Object(d).Destroy((*callDestroy).Destroy)
		err := runTest(c, func(p *gs_core.Container) {
			var d struct {
				Value *callDestroy `autowire:""`
			}
			err := p.Wire(&d)
			assert.Nil(t, err)
		})
		assert.Nil(t, err)
		c.Close()
		assert.True(t, d.destroyed)
	})

	t.Run("call destroy with error", func(t *testing.T) {

		// error
		{
			c := gs_core.New()
			d := &callDestroy{i: 1}
			c.Object(d).Destroy((*callDestroy).DestroyWithError)
			err := runTest(c, func(p *gs_core.Container) {
				var d struct {
					Value *callDestroy `autowire:""`
				}
				err := p.Wire(&d)
				assert.Nil(t, err)
			})
			assert.Nil(t, err)
			c.Close()
			assert.False(t, d.destroyed)
		}

		// nil
		{
			c := gs_core.New()
			d := &callDestroy{}
			c.Object(d).Destroy((*callDestroy).DestroyWithError)
			err := runTest(c, func(p *gs_core.Container) {
				var d struct {
					Value *callDestroy `autowire:""`
				}
				err := p.Wire(&d)
				assert.Nil(t, err)
			})
			assert.Nil(t, err)
			c.Close()
			assert.True(t, d.destroyed)
		}
	})

	t.Run("call interface destroy", func(t *testing.T) {
		c := gs_core.New()
		bd := c.Provide(func() destroyable { return new(callDestroy) }).Destroy(destroyable.Destroy)
		err := runTest(c, func(p *gs_core.Container) {
			var d struct {
				Value destroyable `autowire:""`
			}
			err := p.Wire(&d)
			assert.Nil(t, err)
		})
		assert.Nil(t, err)
		c.Close()
		d := bd.BeanRegistration().(*gs_bean.BeanDefinition).Interface().(*callDestroy)
		assert.True(t, d.destroyed)
	})

	t.Run("call interface destroy with error", func(t *testing.T) {

		// error
		{
			c := gs_core.New()
			bd := c.Provide(func() destroyable { return &callDestroy{i: 1} }).Destroy(destroyable.DestroyWithError)
			err := runTest(c, func(p *gs_core.Container) {
				var d struct {
					Value destroyable `autowire:""`
				}
				err := p.Wire(&d)
				assert.Nil(t, err)
			})
			assert.Nil(t, err)
			c.Close()
			d := bd.BeanRegistration().(*gs_bean.BeanDefinition).Interface()
			assert.False(t, d.(*callDestroy).destroyed)
		}

		// nil
		{
			prop := conf.New()
			prop.Set("int", "0")

			c := gs_core.New()
			bd := c.Provide(func() destroyable { return &callDestroy{} }).Destroy(destroyable.DestroyWithError)

			c.RefreshProperties(prop)
			err := runTest(c, func(p *gs_core.Container) {
				var d struct {
					Value destroyable `autowire:""`
				}
				err := p.Wire(&d)
				assert.Nil(t, err)
			})
			assert.Nil(t, err)
			c.Close()
			d := bd.BeanRegistration().(*gs_bean.BeanDefinition).Interface()
			assert.True(t, d.(*callDestroy).destroyed)
		}
	})
}

type SubNestedAutowireBean struct {
}

type NestedAutowireBean struct {
	SubNestedAutowireBean
}

type PtrNestedAutowireBean struct {
	*SubNestedAutowireBean
}

func TestNestedAutowireBean(t *testing.T) {
	c := gs_core.New()
	c.Object(new(NestedAutowireBean))
	c.Object(&PtrNestedAutowireBean{
		SubNestedAutowireBean: new(SubNestedAutowireBean),
	})
	err := runTest(c, func(p *gs_core.Container) {

		var b struct {
			Value *NestedAutowireBean `autowire:""`
		}
		err := p.Wire(&b)
		assert.Nil(t, err)

		var b0 struct {
			Value *PtrNestedAutowireBean `autowire:""`
		}
		err = p.Wire(&b0)
		assert.Nil(t, err)
	})
	assert.Nil(t, err)
}

type BaseChannel struct {
	AutoCreate bool `value:"${auto-create}"`
	Enable     bool `value:"${enable:=false}"`
}

type WXChannel struct {
	BaseChannel `value:"${sdk.wx}"`
}

type baseChannel struct {
	AutoCreate bool `value:"${auto-create}"`

	// nolint 支持对私有字段注入，但是不推荐！代码扫描请忽略这行。
	enable bool `value:"${enable:=false}"`
}

type wxChannel struct {
	baseChannel `value:"${sdk.wx}"`
}

func TestNestValueField(t *testing.T) {

	t.Run("private", func(t *testing.T) {

		prop := conf.New()
		prop.Set("sdk.wx.auto-create", "true")
		prop.Set("sdk.wx.enable", "true")

		c := gs_core.New()
		c.Object(new(wxChannel))

		c.RefreshProperties(prop)
		err := runTest(c, func(p *gs_core.Container) {
			var channel struct {
				Value *wxChannel `autowire:""`
			}
			err := p.Wire(&channel)
			assert.Nil(t, err)
			assert.Equal(t, channel.Value.enable, true)
			assert.Equal(t, channel.Value.AutoCreate, true)
		})
		assert.Nil(t, err)
	})

	t.Run("public", func(t *testing.T) {
		prop := conf.New()
		prop.Set("sdk.wx.auto-create", "true")
		prop.Set("sdk.wx.enable", "true")

		c := gs_core.New()
		c.Object(new(WXChannel))

		c.RefreshProperties(prop)
		err := runTest(c, func(p *gs_core.Container) {
			var channel struct {
				Value *WXChannel `autowire:""`
			}
			err := p.Wire(&channel)
			assert.Nil(t, err)
			assert.True(t, channel.Value.Enable)
			assert.True(t, channel.Value.AutoCreate)
		})
		assert.Nil(t, err)
	})
}

func TestFnArgCollectBean(t *testing.T) {

	t.Run("interface type", func(t *testing.T) {
		c := gs_core.New()
		c.Provide(newHistoryTeacher("t1")).Name("t1").Export(
			gs.As[Teacher](),
		)
		c.Provide(newHistoryTeacher("t2")).Name("t2").Export(
			gs.As[Teacher](),
		)
		c.Provide(func(teachers []Teacher) func() {
			names := make([]string, 0)
			for _, teacher := range teachers {
				names = append(names, teacher.(*historyTeacher).name)
			}
			sort.Strings(names)
			assert.Equal(t, names, []string{"t1", "t2"})
			return func() {}
		})
		err := c.Refresh()
		assert.Nil(t, err)
	})
}

type filter interface {
	Filter(input string) string
}

type filterImpl struct {
}

func (_ *filterImpl) Filter(input string) string {
	return input
}

func TestBeanCache(t *testing.T) {

	t.Run("not implement interface", func(t *testing.T) {
		assert.Panic(t, func() {
			c := gs_core.New()
			c.Object(func() {}).Export(
				gs.As[filter](),
			)
		}, "doesn't implement interface gs_core_test.filter")
	})

	t.Run("implement interface", func(t *testing.T) {

		var server struct {
			F1 filter `autowire:"f1"`
			F2 filter `autowire:"f2"`
		}

		c := gs_core.New()
		c.Provide(func() filter { return new(filterImpl) }).Name("f1")
		c.Object(new(filterImpl)).Export(
			gs.As[filter](),
		).Name("f2")
		c.Object(&server)

		err := c.Refresh()
		assert.Nil(t, err)
	})
}

type IntInterface interface {
	Value() int
}

type Integer int

func (i Integer) Value() int {
	return int(i)
}

func TestIntInterface(t *testing.T) {
	c := gs_core.New()
	c.Provide(func() IntInterface { return Integer(5) })
	err := c.Refresh()
	assert.Nil(t, err)
}

type ArrayProperties struct {
	Int      []int           `value:"${int.array:=}"`
	Int8     []int8          `value:"${int8.array:=}"`
	Int16    []int16         `value:"${int16.array:=}"`
	Int32    []int32         `value:"${int32.array:=}"`
	Int64    []int64         `value:"${int64.array:=}"`
	UInt     []uint          `value:"${uint.array:=}"`
	UInt8    []uint8         `value:"${uint8.array:=}"`
	UInt16   []uint16        `value:"${uint16.array:=}"`
	UInt32   []uint32        `value:"${uint32.array:=}"`
	UInt64   []uint64        `value:"${uint64.array:=}"`
	String   []string        `value:"${string.array:=}"`
	Bool     []bool          `value:"${bool.array:=}"`
	Duration []time.Duration `value:"${duration.array:=}"`
	Time     []time.Time     `value:"${time.array:=}"`
}

func TestProperties(t *testing.T) {

	t.Run("array properties", func(t *testing.T) {
		b := new(ArrayProperties)
		c := gs_core.New()
		c.Object(b)
		err := c.Refresh()
		assert.Nil(t, err)
	})

	t.Run("map default value ", func(t *testing.T) {

		obj := struct {
			Int  int               `value:"${int:=5}"`
			IntA int               `value:"${int_a:=5}"`
			Map  map[string]string `value:"${map:=}"`
			MapA map[string]string `value:"${map_a:=}"`
		}{}

		prop := conf.New()
		prop.Set("map_a.nba", "nba")
		prop.Set("map_a.cba", "cba")
		prop.Set("int_a", "3")
		prop.Set("int_b", "4")

		c := gs_core.New()
		c.Object(&obj)

		c.RefreshProperties(prop)
		err := c.Refresh()
		assert.Nil(t, err)

		assert.Equal(t, obj.Int, 5)
		assert.Equal(t, obj.IntA, 3)
	})
}

type FirstDestroy struct {
	T1 *Second1Destroy `autowire:""`
	T2 *Second2Destroy `autowire:""`
}

type Second1Destroy struct {
	T *ThirdDestroy `autowire:""`
}

type Second2Destroy struct {
	T *ThirdDestroy `autowire:""`
}

type ThirdDestroy struct {
}

func TestDestroy(t *testing.T) {

	destroyIndex := 0
	destroyArray := []int{0, 0, 0, 0}

	c := gs_core.New()
	c.Object(new(FirstDestroy)).Destroy(
		func(_ *FirstDestroy) {
			fmt.Println("::FirstDestroy")
			destroyArray[destroyIndex] = 1
			destroyIndex++
		})
	c.Object(new(Second2Destroy)).Destroy(
		func(_ *Second2Destroy) {
			fmt.Println("::Second2Destroy")
			destroyArray[destroyIndex] = 2
			destroyIndex++
		})
	c.Object(new(Second1Destroy)).Destroy(
		func(_ *Second1Destroy) {
			fmt.Println("::Second1Destroy")
			destroyArray[destroyIndex] = 2
			destroyIndex++
		})
	c.Object(new(ThirdDestroy)).Destroy(
		func(_ *ThirdDestroy) {
			fmt.Println("::ThirdDestroy")
			destroyArray[destroyIndex] = 4
			destroyIndex++
		})
	err := c.Refresh()
	assert.Nil(t, err)
	c.Close()

	assert.Equal(t, destroyArray, []int{1, 2, 2, 4})
}

// type Obj struct {
// 	i int
// }
//
// type ObjFactory struct{}
//
// func (factory *ObjFactory) NewObj(i int) *Obj { return &Obj{i: i} }
//
// func TestCreateBean(t *testing.T) {
// 	c := gs_core.New()
// 	c.Object(&ObjFactory{})
// 	err := runTest(c, func(p *gs_core.Container) {
// 		b, err := p.Wire((*ObjFactory).NewObj, gs_arg.Index(1, gs_arg.Tag("${i:=5}")))
// 		fmt.Println(b, err)
// 	})
// 	assert.Nil(t, err)
// }

func TestDefaultSpringContext(t *testing.T) {

	// t.Run("bean:test_ctx:", func(t *testing.T) {
	//
	// 	c := gs_core.New()
	//
	// 	c.Object(&BeanZero{5}).
	// 		Condition(gs_cond.OnMissingBean("null")).
	// 		Condition(gs_cond.OnMissingProperty("a"))
	//
	// 	err := runTest(c, func(p *gs_core.Container) {
	// 		var b *BeanZero
	// 		err := p.Wire(&b)
	// 		assert.Error(t, err, "can't find bean, bean:\"\"")
	// 	})
	// 	assert.Nil(t, err)
	// })

	t.Run("option withClassName Condition", func(t *testing.T) {

		prop := conf.New()
		prop.Set("president", "CaiYuanPei")
		prop.Set("class_floor", "2")

		c := gs_core.New()
		c.Provide(NewClassRoom, gs_arg.Bind(withClassName,
			gs_arg.Tag("${class_name:=二年级03班}"),
			gs_arg.Tag("${class_floor:=3}"),
		).Condition(gs_cond.OnProperty("class_name_enable")))

		c.RefreshProperties(prop)
		err := runTest(c, func(p *gs_core.Container) {
			var cls struct {
				Value *ClassRoom `autowire:""`
			}
			err := p.Wire(&cls)
			assert.Nil(t, err)
			assert.Equal(t, cls.Value.floor, 0)
			assert.Equal(t, len(cls.Value.students), 0)
			assert.Equal(t, cls.Value.className, "default")
			assert.Equal(t, cls.Value.President, "CaiYuanPei")
		})
		assert.Nil(t, err)
	})

	t.Run("option withClassName Apply", func(t *testing.T) {
		prop := conf.New()
		prop.Set("president", "CaiYuanPei")

		onProperty := gs_cond.OnProperty("class_name_enable")
		c := gs_core.New()
		c.Provide(NewClassRoom,
			gs_arg.Bind(withClassName,
				gs_arg.Tag("${class_name:=二年级03班}"),
				gs_arg.Tag("${class_floor:=3}"),
			).Condition(onProperty),
		)

		c.RefreshProperties(prop)
		err := runTest(c, func(p *gs_core.Container) {
			// var cls *ClassRoom
			// err := p.Wire(&cls)
			// assert.Nil(t, err)
			// assert.Equal(t, cls.floor, 0)
			// assert.Equal(t, len(cls.students), 0)
			// assert.Equal(t, cls.className, "default")
			// assert.Equal(t, cls.President, "CaiYuanPei")
		})
		assert.Nil(t, err)
	})

	t.Run("method bean cond", func(t *testing.T) {
		prop := conf.New()
		prop.Set("server.version", "1.0.0")

		c := gs_core.New()
		parent := c.Object(new(Server))
		c.Provide((*Server).Consumer, parent).Condition(gs_cond.OnProperty("consumer.enable"))

		c.RefreshProperties(prop)
		err := runTest(c, func(p *gs_core.Container) {

			// var s *Server
			// err := p.Wire(&s)
			// assert.Nil(t, err)
			// assert.Equal(t, s.Version, "1.0.0")
			//
			// var consumer *Consumer
			// err = p.Wire(&consumer)
			// assert.Error(t, err, "can't find bean, bean:\"\"")
		})
		assert.Nil(t, err)
	})
}

// TODO 现在的方式父 Bean 不存在子 Bean 创建的时候会报错
// func TestDefaultSpringContext_ParentNotRegister(t *testing.T) {
//
//	c := gs.New()
//	parent := c.Provide(NewServerInterface).Condition(cond.OnProperty("server.is.nil"))
//	c.Provide(ServerInterface.Consumer, parent.Name())
//
//	c.Refresh()
//
//	var s *Server
//	ok := p.Wire(&s)
//	util.Equal(t, ok, false)
//
//	var c *Consumer
//	ok = p.Wire(&c)
//	util.Equal(t, ok, false)
// }

func TestDefaultSpringContext_ConditionOnBean(t *testing.T) {
	c := gs_core.New()

	c1 := gs_cond.Or(
		gs_cond.OnProperty("null").MatchIfMissing(),
	)

	c.Object(&BeanZero{5}).Condition(
		gs_cond.And(
			c1,
			gs_cond.OnMissingBeanSelector(gs.BeanSelectorImpl{Name: "null"}),
		),
	)
	c.Object(new(BeanOne)).Condition(
		gs_cond.And(
			c1,
			gs_cond.OnMissingBeanSelector(gs.BeanSelectorImpl{Name: "null"}),
		),
	)

	c.Object(new(BeanTwo)).Condition(gs_cond.OnBeanSelector(gs.BeanSelectorImpl{Name: "BeanOne"}))
	c.Object(new(BeanTwo)).Name("another_two").Condition(gs_cond.OnBeanSelector(gs.BeanSelectorImpl{Name: "Null"}))

	err := runTest(c, func(p *gs_core.Container) {

		// var two *BeanTwo
		// err := p.Wire(&two)
		// assert.Nil(t, err)
		//
		// err = p.Wire(&two, "another_two")
		// assert.Error(t, err, "can't find bean, bean:\"another_two\"")
	})
	assert.Nil(t, err)
}

func TestDefaultSpringContext_ConditionOnMissingBean(t *testing.T) {
	for i := 0; i < 20; i++ { // 测试 Find 无需绑定，不要排序
		c := gs_core.New()
		c.Object(&BeanZero{5})
		c.Object(new(BeanOne))
		c.Object(new(BeanTwo)).Condition(gs_cond.OnMissingBeanSelector(gs.BeanSelectorImpl{Name: "BeanOne"}))
		c.Object(new(BeanTwo)).Name("another_two").Condition(gs_cond.OnMissingBeanSelector(gs.BeanSelectorImpl{Name: "Null"}))
		err := runTest(c, func(p *gs_core.Container) {

			// var two *BeanTwo
			// err := p.Wire(&two)
			// assert.Nil(t, err)
			//
			// err = p.Wire(&two, "another_two")
			// assert.Nil(t, err)
		})
		assert.Nil(t, err)
	}
}

// func TestFunctionCondition(t *testing.T) {
//	c := gs.New()
//
//	fn := func(c cond.Context) bool { return true }
//	c1 := cond.OnMatches(fn)
//	assert.True(t, c1.Matches(c))
//
//	fn = func(c cond.Context) bool { return false }
//	c2 := cond.OnMatches(fn)
//	assert.False(t, c2.Matches(c))
// }
//
// func TestPropertyCondition(t *testing.T) {
//
//	c := gs.New()
//	c.Property("int", 3)
//	c.Property("parent.child", 0)
//
//	c1 := cond.OnProperty("int")
//	assert.True(t, c1.Matches(c))
//
//	c2 := cond.OnProperty("bool")
//	assert.False(t, c2.Matches(c))
//
//	c3 := cond.OnProperty("parent")
//	assert.True(t, c3.Matches(c))
//
//	c4 := cond.OnProperty("parent123")
//	assert.False(t, c4.Matches(c))
// }
//
// func TestMissingPropertyCondition(t *testing.T) {
//
//	c := gs.New()
//	c.Property("int", 3)
//	c.Property("parent.child", 0)
//
//	c1 := cond.OnMissingProperty("int")
//	assert.False(t, c1.Matches(c))
//
//	c2 := cond.OnMissingProperty("bool")
//	assert.True(t, c2.Matches(c))
//
//	c3 := cond.OnMissingProperty("parent")
//	assert.False(t, c3.Matches(c))
//
//	c4 := cond.OnMissingProperty("parent123")
//	assert.True(t, c4.Matches(c))
// }
//
// func TestPropertyValueCondition(t *testing.T) {
//
//	c := gs.New()
//	c.Property("str", "this is a str")
//	c.Property("int", 3)
//
//	c1 := cond.OnPropertyValue("int", 3)
//	assert.True(t, c1.Matches(c))
//
//	c2 := cond.OnPropertyValue("int", "3")
//	assert.False(t, c2.Matches(c))
//
//	c3 := cond.OnPropertyValue("int", "go:$>2&&$<4")
//	assert.True(t, c3.Matches(c))
//
//	c4 := cond.OnPropertyValue("bool", true)
//	assert.False(t, c4.Matches(c))
//
//	c5 := cond.OnPropertyValue("str", "\"$\"==\"this is a str\"")
//	assert.True(t, c5.Matches(c))
// }
//
// func TestBeanCondition(t *testing.T) {
//
//	c := gs.New()
//	c.Object(&BeanZero{5})
//	c.Object(new(BeanOne))
//	c.Refresh()
//
//	c1 := cond.OnBean("BeanOne")
//	assert.True(t, c1.Matches(c))
//
//	c2 := cond.OnBean("Null")
//	assert.False(t, c2.Matches(c))
// }
//
// func TestMissingBeanCondition(t *testing.T) {
//
//	c := gs.New()
//	c.Object(&BeanZero{5})
//	c.Object(new(BeanOne))
//	c.Refresh()
//
//	c1 := cond.OnMissingBean("BeanOne")
//	assert.False(t, c1.Matches(c))
//
//	c2 := cond.OnMissingBean("Null")
//	assert.True(t, c2.Matches(c))
// }
//
// func TestExpressionCondition(t *testing.T) {
//
// }
//
// func TestConditional(t *testing.T) {
//
//	c := gs.New()
//	c.Property("bool", false)
//	c.Property("int", 3)
//	c.Refresh()
//
//	c1 := cond.OnProperty("int")
//	assert.True(t, c1.Matches(c))
//
//	c2 := cond.OnProperty("int").OnBean("null")
//	assert.False(t, c2.Matches(c))
//
//	assert.Panic(t, func() {
//		c3 := cond.OnProperty("int").And()
//		assert.Equal(t, c3.Matches(c), true)
//	}, "no condition in last node")
//
//	c4 := cond.OnPropertyValue("int", 3).
//		And().
//		OnPropertyValue("bool", false)
//	assert.True(t, c4.Matches(c))
//
//	c5 := cond.OnPropertyValue("int", 3).
//		And().
//		OnPropertyValue("bool", true)
//	assert.False(t, c5.Matches(c))
//
//	c6 := cond.OnPropertyValue("int", 2).
//		Or().
//		OnPropertyValue("bool", true)
//	assert.False(t, c6.Matches(c))
//
//	c7 := cond.OnPropertyValue("int", 2).
//		Or().
//		OnPropertyValue("bool", false)
//	assert.True(t, c7.Matches(c))
//
//	assert.Panic(t, func() {
//		c8 := cond.OnPropertyValue("int", 2).
//			Or().
//			OnPropertyValue("bool", false).
//			Or()
//		assert.Equal(t, c8.Matches(c), true)
//	}, "no condition in last node")
//
//	c9 := cond.OnPropertyValue("int", 2).
//		Or().
//		OnPropertyValue("bool", false).
//		OnPropertyValue("bool", false)
//	assert.True(t, c9.Matches(c))
// }
//
// func TestNotCondition(t *testing.T) {
//
//	c := gs.New()
//	c.Property(environ.SpringProfilesActive, "test")
//	c.Refresh()
//
//	profileCond := cond.OnProfile("test")
//	assert.True(t, profileCond.Matches(c))
//
//	notCond := cond.Not(profileCond)
//	assert.False(t, notCond.Matches(c))
//
//	c1 := cond.OnPropertyValue("int", 2).
//		And().
//		On(cond.Not(profileCond))
//	assert.False(t, c1.Matches(c))
//
//	c2 := cond.OnProfile("test").
//		And().
//		On(cond.Not(profileCond))
//	assert.False(t, c2.Matches(c))
// }

// func TestInvoke(t *testing.T) {
//
// 	// t.Run("not run", func(t *testing.T) {
// 	// 	prop := conf.New()
// 	// 	prop.Set("version", "v0.0.1")
// 	//
// 	// 	c := gs_core.New()
// 	// 	c.Object(func() {})
// 	//
// 	// 	c.RefreshProperties(prop)
// 	// 	err := runTest(c, func(p *gs_core.Container) {
// 	// 		_, _ = p.Invoke(func(f func(), version string) {
// 	// 			fmt.Println("version:", version)
// 	// 		}, gs_arg.Tag(""), gs_arg.Tag("${version}"))
// 	// 	})
// 	// 	assert.Nil(t, err)
// 	// })
//
// 	// t.Run("run", func(t *testing.T) {
// 	// 	prop := conf.New()
// 	// 	prop.Set("version", "v0.0.1")
// 	// 	prop.Set("spring.profiles.active", "dev")
// 	//
// 	// 	c := gs_core.New()
// 	// 	c.Object(func() {})
// 	//
// 	// 	c.RefreshProperties(prop)
// 	// 	err := runTest(c, func(p *gs_core.Container) {
// 	// 		fn := func(f func(), version string) {
// 	// 			fmt.Println("version:", version)
// 	// 		}
// 	// 		_, _ = p.Invoke(fn, gs_arg.Tag(""), gs_arg.Tag("${version}"))
// 	// 	})
// 	// 	assert.Nil(t, err)
// 	// })
// }

type emptyStructA struct{}

type emptyStructB struct{}

func TestEmptyStruct(t *testing.T) {

	c := gs_core.New()
	objA := &emptyStructA{}
	c.Object(objA)
	objB := &emptyStructB{}
	c.Object(objB)
	err := c.Refresh()
	assert.Nil(t, err)

	// objA 和 objB 的地址相同但是类型确实不一样。
	fmt.Printf("objA:%p objB:%p\n", objA, objB)
	fmt.Printf("objA:%#v objB:%#v\n", objA, objB)
}

func TestMapCollection(t *testing.T) {

	type mapValue struct {
		v string
	}

	t.Run("", func(t *testing.T) {
		c := gs_core.New()
		c.Object(&mapValue{"a"}).Name("a")
		c.Object(&mapValue{"b"}).Name("b")
		c.Object(&mapValue{"c"}).Name("c").Condition(gs_cond.Not(gs_cond.OnMissingProperty("a")))
		err := runTest(c, func(p *gs_core.Container) {

			// var vSlice []*mapValue
			// err := p.Wire(&vSlice)
			// assert.Nil(t, err)
			// fmt.Println(vSlice)
			//
			// var vMap map[string]*mapValue
			// err = p.Wire(&vMap)
			// assert.Nil(t, err)
			// fmt.Println(vMap)
		})
		assert.Nil(t, err)
	})
}

type circularA struct {
	b *circularB
}

func newCircularA(b *circularB) *circularA {
	return &circularA{b: b}
}

type circularB struct {
	A *circularA `autowire:",lazy"`
}

func newCircularB() *circularB {
	return new(circularB)
}

func TestLazy(t *testing.T) {
	for i := 0; i < 1; i++ {
		prop := conf.New()
		prop.Set("spring.allow-circular-references", "true")

		c := gs_core.New()
		c.Provide(newCircularA)
		c.Provide(newCircularB)
		d := struct {
			b *circularB `autowire:""`
		}{}
		c.Object(&d)

		c.RefreshProperties(prop)
		err := c.Refresh()
		assert.Nil(t, err)
		assert.NotNil(t, d.b)
		assert.NotNil(t, d.b.A)
	}
}

type memory struct {
}

func (m *memory) OnInit(ctx *gs_core.Container) error {
	fmt.Println("memory.OnInit")
	return nil
}

func (m *memory) OnDestroy() {
	fmt.Println("memory.OnDestroy")
}

type table struct {
	_ *memory `autowire:""`
}

func (t *table) OnInit(ctx *gs_core.Container) error {
	fmt.Println("table.OnInit")
	return nil
}

func (t *table) OnDestroy() {
	fmt.Println("table.OnDestroy")
}

func TestDestroyDependence(t *testing.T) {
	c := gs_core.New()
	c.Object(new(memory))
	c.Object(new(table)).Name("aaa")
	c.Object(new(table)).Name("bbb")
	err := c.Refresh()
	assert.Nil(t, err)
}

type DynamicConfig struct {
	Int   gs_dync.Value[int64]             `value:"${int:=3}" expr:"$<6"`
	Float gs_dync.Value[float64]           `value:"${float:=1.2}"`
	Map   gs_dync.Value[map[string]string] `value:"${map:=}"`
	Slice gs_dync.Value[[]string]          `value:"${slice:=}"`
}

var _ gs.Refreshable = (*DynamicConfig)(nil)

func (d *DynamicConfig) OnRefresh(prop conf.Properties, param conf.BindParam) error {
	fmt.Println("DynamicConfig.OnRefresh")
	return nil
}

type DynamicConfigWrapper struct {
	Wrapper DynamicConfig `value:"${wrapper}"` // struct 自身的 refresh 会抑制 fields 的 refresh
}

var _ gs.Refreshable = (*DynamicConfigWrapper)(nil)

func (d *DynamicConfigWrapper) OnRefresh(prop conf.Properties, param conf.BindParam) error {
	fmt.Println("DynamicConfig.OnRefresh")
	return nil
}

func TestDynamic(t *testing.T) {

	cfg := new(DynamicConfig)
	wrapper := new(DynamicConfigWrapper)

	c := gs_core.New()
	c.Object(cfg)
	c.Object(wrapper)
	err := c.Refresh()
	assert.Nil(t, err)

	{
		b, _ := json.Marshal(cfg)
		assert.Equal(t, string(b), `{"Int":3,"Float":1.2,"Map":{},"Slice":[]}`)
		b, _ = json.Marshal(wrapper)
		assert.Equal(t, string(b), `{"Wrapper":{"Int":null,"Float":null,"Map":null,"Slice":null}}`)
	}

	{
		p := conf.New()
		p.Set("int", "4")
		p.Set("float", "2.3")
		p.Set("map.a", "1")
		p.Set("map.b", "2")
		p.Set("slice[0]", "3")
		p.Set("slice[1]", "4")
		p.Set("wrapper.int", "3")
		p.Set("wrapper.float", "1.5")
		p.Set("wrapper.map.a", "9")
		p.Set("wrapper.map.b", "8")
		p.Set("wrapper.slice[0]", "4")
		p.Set("wrapper.slice[1]", "6")
		c.RefreshProperties(p)
	}

	{
		b, _ := json.Marshal(cfg)
		assert.Equal(t, string(b), `{"Int":4,"Float":2.3,"Map":{"a":"1","b":"2"},"Slice":["3","4"]}`)
		b, _ = json.Marshal(wrapper)
		assert.Equal(t, string(b), `{"Wrapper":{"Int":null,"Float":null,"Map":null,"Slice":null}}`)
	}

	{
		p := conf.New()
		p.Set("int", "6")
		p.Set("float", "5.1")
		p.Set("map.a", "9")
		p.Set("map.b", "8")
		p.Set("slice[0]", "7")
		p.Set("slice[1]", "6")
		p.Set("wrapper.int", "9")
		p.Set("wrapper.float", "8.4")
		p.Set("wrapper.map.a", "3")
		p.Set("wrapper.map.b", "4")
		p.Set("wrapper.slice[0]", "2")
		p.Set("wrapper.slice[1]", "1")
		err = c.RefreshProperties(p)
		assert.Error(t, err, "validate failed on \"\\$<6\" for value 6")
	}

	{
		b, _ := json.Marshal(cfg)
		assert.Equal(t, string(b), `{"Int":4,"Float":5.1,"Map":{"a":"9","b":"8"},"Slice":["7","6"]}`)
		b, _ = json.Marshal(wrapper)
		assert.Equal(t, string(b), `{"Wrapper":{"Int":null,"Float":null,"Map":null,"Slice":null}}`)
	}

	{
		p := conf.New()
		p.Set("int", "1")
		p.Set("float", "5.1")
		p.Set("map.a", "9")
		p.Set("map.b", "8")
		p.Set("slice[0]", "7")
		p.Set("slice[1]", "6")
		p.Set("wrapper.int", "9")
		p.Set("wrapper.float", "8.4")
		p.Set("wrapper.map.a", "3")
		p.Set("wrapper.map.b", "4")
		p.Set("wrapper.slice[0]", "2")
		p.Set("wrapper.slice[1]", "1")
		err = c.RefreshProperties(p)
		assert.Nil(t, err)
	}

	{
		b, _ := json.Marshal(cfg)
		assert.Equal(t, string(b), `{"Int":1,"Float":5.1,"Map":{"a":"9","b":"8"},"Slice":["7","6"]}`)
		b, _ = json.Marshal(wrapper)
		assert.Equal(t, string(b), `{"Wrapper":{"Int":null,"Float":null,"Map":null,"Slice":null}}`)
	}
}

type ChildBean struct {
	s string
}

type ConfigurationBean struct {
	s string
}

func NewConfigurationBean(s string) *ConfigurationBean {
	return &ConfigurationBean{s}
}

func (c *ConfigurationBean) NewChild() *ChildBean {
	return &ChildBean{c.s}
}

func TestConfiguration(t *testing.T) {
	c := gs_core.New()
	c.Object(&ConfigurationBean{"123"}).Configuration(gs.ConfigurationParam{Excludes: []string{"NewBean"}}).Name("123")
	c.Provide(NewConfigurationBean, gs_arg.Value("456")).Configuration().Name("456")
	if err := c.Refresh(); err != nil {
		t.Fatal(err)
	}
}
