/*
 * Copyright 2012-2024 the original author or authors.
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

package gs

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_app"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/gs/internal/gs_conf"
	"github.com/go-spring/spring-core/gs/internal/gs_core"
	"github.com/go-spring/spring-core/gs/internal/gs_dync"
	"github.com/go-spring/spring-core/gs/sysconf"
)

const (
	Version = "go-spring@v1.1.3"
	Website = "https://go-spring.com/"
)

/************************************ arg ***********************************/

type Arg = gs.Arg

// NilArg return a ValueArg with a value of nil.
func NilArg() gs_arg.ValueArg {
	return gs_arg.Nil()
}

// ValueArg return a ValueArg with a value of v.
func ValueArg(v interface{}) gs_arg.ValueArg {
	return gs_arg.Value(v)
}

// IndexArg returns an IndexArg.
func IndexArg(n int, arg Arg) gs_arg.IndexArg {
	return gs_arg.Index(n, arg)
}

// OptionArg 返回 Option 函数的参数绑定。
func OptionArg(fn interface{}, args ...Arg) *gs_arg.OptionArg {
	return gs_arg.Option(fn, args...)
}

func BindArg(fn interface{}, args []Arg, skip int) (*gs_arg.Callable, error) {
	return gs_arg.Bind(fn, args, skip)
}

// MustBindArg 为 Option 方法绑定运行时参数。
func MustBindArg(fn interface{}, args ...Arg) *gs_arg.Callable {
	return gs_arg.MustBind(fn, args...)
}

/************************************ cond ***********************************/

type (
	Condition      = gs.Condition
	CondContext    = gs.CondContext
	PropertyOption = gs_cond.PropertyOption
	ConditionError = gs_cond.ConditionError
)

func NewCondError(cond gs.Condition, err error) error {
	return gs_cond.NewCondError(cond, err)
}

func OnFunc(fn func(ctx CondContext) (bool, error)) Condition {
	return gs_cond.OnFunc(fn)
}

func MatchIfMissing() PropertyOption {
	return gs_cond.MatchIfMissing()
}

func HavingValue(havingValue string) PropertyOption {
	return gs_cond.HavingValue(havingValue)
}

func OnProperty(name string, options ...PropertyOption) Condition {
	return gs_cond.OnProperty(name, options...)
}

func OnMissingProperty(name string) Condition {
	return gs_cond.OnMissingProperty(name)
}

func OnBean(selector BeanSelector) Condition {
	return gs_cond.OnBean(selector)
}

func OnMissingBean(selector BeanSelector) Condition {
	return gs_cond.OnMissingBean(selector)
}

func OnSingleBean(selector BeanSelector) Condition {
	return gs_cond.OnSingleBean(selector)
}

func CustomFunction(name string, fn interface{}) {
	gs_cond.CustomFunction(name, fn)
}

func OnExpression(expression string) Condition {
	return gs_cond.OnExpression(expression)
}

// Not returns a Condition that returns true when the given Condition returns false.
func Not(c Condition) Condition {
	return gs_cond.Not(c)
}

// Or returns a Condition that returns true when any of the given Conditions returns true.
func Or(cond ...Condition) Condition {
	return gs_cond.Or(cond...)
}

// And returns a Condition that returns true when all the given Conditions return true.
func And(cond ...Condition) Condition {
	return gs_cond.And(cond...)
}

// None returns a Condition that returns true when none of the given Conditions returns true.
func None(cond ...Condition) Condition {
	return gs_cond.None(cond...)
}

func OnProfile(profile string) Condition {
	return OnProperty("spring.profiles.active", HavingValue(profile))
}

/************************************ ioc ************************************/

type (
	BeanSelector = gs.BeanSelector
)

type (
	BeanInit    = gs_bean.BeanInit
	BeanDestroy = gs_bean.BeanDestroy
)

type (
	Context      = gs.Context
	ContextAware = gs.ContextAware
)

type (
	Properties  = gs.Properties
	Dync[T any] = gs_dync.Value[T]
)

type (
	RegisteredBean = gs.RegisteredBean
	BeanDefinition = gs.BeanDefinition
)

func NewBean(objOrCtor interface{}, ctorArgs ...gs.Arg) *BeanDefinition {
	return gs_core.NewBean(objOrCtor, ctorArgs...)
}

/************************************ boot ***********************************/

var boot *gs_app.Boot

func bootRun() error {
	if boot != nil {
		if err := boot.Run(); err != nil {
			return err
		}
		boot = nil // Boot 阶段结束，释放资源
	}
	return nil
}

// Boot returns a [gs_app.Boot] instance.
func Boot() *gs_app.Boot {
	if boot == nil {
		boot = gs_app.NewBoot()
	}
	return boot
}

/*********************************** app *************************************/

type (
	AppRunner  = gs_app.AppRunner
	AppServer  = gs_app.AppServer
	AppContext = gs_app.AppContext
)

var app = gs_app.NewApp()

// Start starts the app, usually used in test mode.
func Start() error {
	return app.Start()
}

// Stop stops the app, usually used in test mode.
func Stop() {
	app.Stop()
}

// Run runs the app, and waits to exit by watching the interrupt signal.
func Run() error {
	printBanner()
	if err := bootRun(); err != nil {
		return err
	}
	return app.Run()
}

// ShutDown shuts down the app.
func ShutDown(msg ...string) {
	app.ShutDown(msg...)
}

// Config returns the app configuration.
func Config() *gs_conf.AppConfig {
	return app.P
}

// Object returns a bean definition for the given object.
func Object(i interface{}) *RegisteredBean {
	b := gs_core.NewBean(reflect.ValueOf(i))
	return app.C.Register(b)
}

// Provide returns a bean definition for the given constructor.
func Provide(ctor interface{}, args ...Arg) *RegisteredBean {
	b := gs_core.NewBean(ctor, args...)
	return app.C.Register(b)
}

// Register registers a bean definition.
func Register(b *BeanDefinition) *RegisteredBean {
	return app.C.Register(b)
}

// GroupRegister registers a group of bean definitions.
func GroupRegister(fn func(p Properties) ([]*BeanDefinition, error)) {
	app.C.GroupRegister(fn)
}

// Runner registers a bean definition for the given runner.
func Runner(objOrCtor interface{}, ctorArgs ...Arg) *RegisteredBean {
	b := gs_core.NewBean(objOrCtor, ctorArgs...)
	b.Export((*AppRunner)(nil))
	return app.C.Register(b)
}

// Server registers a bean definition for the given server.
func Server(objOrCtor interface{}, ctorArgs ...Arg) *RegisteredBean {
	b := gs_core.NewBean(objOrCtor, ctorArgs...)
	b.Export((*AppServer)(nil))
	return app.C.Register(b)
}

// RefreshProperties refreshes the app configuration.
func RefreshProperties(p Properties) error {
	return app.C.RefreshProperties(p)
}

/********************************** banner ***********************************/

var appBanner = `
   ____    ___            ____    ____    ____    ___   _   _    ____ 
  / ___|  / _ \          / ___|  |  _ \  |  _ \  |_ _| | \ | |  / ___|
 | |  _  | | | |  _____  \___ \  | |_) | | |_) |  | |  |  \| | | |  _ 
 | |_| | | |_| | |_____|  ___) | |  __/  |  _ <   | |  | |\  | | |_| |
  \____|  \___/          |____/  |_|     |_| \_\ |___| |_| \_|  \____| 
`

// Banner sets the banner of the app.
func Banner(banner string) {
	appBanner = banner
}

// printBanner prints the banner of the app.
func printBanner() {
	if len(appBanner) == 0 {
		return
	}

	if appBanner[0] != '\n' {
		fmt.Println()
	}

	maxLength := 0
	for _, s := range strings.Split(appBanner, "\n") {
		fmt.Printf("\x1b[36m%s\x1b[0m\n", s) // CYAN
		if len(s) > maxLength {
			maxLength = len(s)
		}
	}

	if appBanner[len(appBanner)-1] != '\n' {
		fmt.Println()
	}

	var padding []byte
	if n := (maxLength - len(Version)) / 2; n > 0 {
		padding = make([]byte, n)
		for i := range padding {
			padding[i] = ' '
		}
	}
	fmt.Println(string(padding) + Version + "\n")
}

/********************************** utility **********************************/

// AllowCircularReferences allows circular references between beans.
func AllowCircularReferences(enable bool) {
	err := sysconf.Set("spring.allow-circular-references", enable)
	if err != nil {
		panic(err)
	}
}

// ForceAutowireIsNullable forces autowire is nullable.
func ForceAutowireIsNullable(enable bool) {
	err := sysconf.Set("spring.force-autowire-is-nullable", enable)
	if err != nil {
		panic(err)
	}
}
