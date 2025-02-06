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

	"github.com/go-spring/spring-core/conf/sysconf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_app"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/gs/internal/gs_conf"
	"github.com/go-spring/spring-core/gs/internal/gs_core"
	"github.com/go-spring/spring-core/gs/internal/gs_dync"
)

const (
	Version = "go-spring@v1.1.3"
	Website = "https://go-spring.com/"
)

/************************************ arg ***********************************/

type Arg = gs.Arg

// NilArg returns a ValueArg with a nil value.
func NilArg() gs_arg.ValueArg {
	return gs_arg.Nil()
}

// TagArg returns a TagArg with the specified tag.
func TagArg(tag string) gs_arg.TagArg {
	return gs_arg.Tag(tag)
}

// TypeArg returns a TagArg with the specified bean type.
func TypeArg[T any]() gs_arg.TagArg {
	return gs_arg.BeanTag[T]()
}

// ValueArg returns a ValueArg with the specified value.
func ValueArg(v interface{}) gs_arg.ValueArg {
	return gs_arg.Value(v)
}

// IndexArg returns an IndexArg with the specified index and argument.
func IndexArg(n int, arg Arg) gs_arg.IndexArg {
	return gs_arg.Index(n, arg)
}

// BindArg binds runtime arguments to a given function.
func BindArg(fn interface{}, args ...Arg) *gs_arg.Callable {
	return gs_arg.MustBind(fn, args...)
}

// OptionArg returns an OptionArg for the specified function and arguments.
func OptionArg(fn interface{}, args ...Arg) *gs_arg.OptionArg {
	return gs_arg.Option(fn, args...)
}

/************************************ cond ***********************************/

type (
	CondBean    = gs.CondBean
	CondFunc    = gs.CondFunc
	Condition   = gs.Condition
	CondContext = gs.CondContext
)

// OnFunc creates a Condition based on the provided function.
func OnFunc(fn CondFunc) Condition {
	return gs_cond.OnFunc(fn)
}

// OnProperty creates a Condition based on a property name and options.
func OnProperty(name string) *gs_cond.CondOnProperty {
	return gs_cond.OnProperty(name)
}

// OnMissingProperty creates a Condition that checks for a missing property.
func OnMissingProperty(name string) Condition {
	return gs_cond.OnMissingProperty(name)
}

// OnBean creates a Condition based on a BeanSelector.
func OnBean(s BeanSelector) Condition {
	return gs_cond.OnBean(s)
}

// OnMissingBean creates a Condition for when a specific bean is missing.
func OnMissingBean(s BeanSelector) Condition {
	return gs_cond.OnMissingBean(s)
}

// OnSingleBean creates a Condition for when only one instance of a bean exists.
func OnSingleBean(s BeanSelector) Condition {
	return gs_cond.OnSingleBean(s)
}

// RegisterExpressFunc registers a custom expression function.
func RegisterExpressFunc(name string, fn interface{}) {
	gs_cond.RegisterExpressFunc(name, fn)
}

// OnExpression creates a Condition based on a custom expression.
func OnExpression(expression string) Condition {
	return gs_cond.OnExpression(expression)
}

// Not creates a Condition that negates the given Condition.
func Not(c Condition) Condition {
	return gs_cond.Not(c)
}

// Or creates a Condition that is true if any of the given Conditions are true.
func Or(cond ...Condition) Condition {
	return gs_cond.Or(cond...)
}

// And creates a Condition that is true if all of the given Conditions are true.
func And(cond ...Condition) Condition {
	return gs_cond.And(cond...)
}

// None creates a Condition that is true if none of the given Conditions are true.
func None(cond ...Condition) Condition {
	return gs_cond.None(cond...)
}

// OnProfile creates a Condition based on the active profile.
func OnProfile(profile string) Condition {
	return OnProperty("spring.profiles.active").HavingValue(profile)
}

/************************************ ioc ************************************/

type (
	BeanSelector = gs.BeanSelector
)

type (
	Properties = gs.Properties
)

type (
	Context      = gs.Context
	ContextAware = gs.ContextAware
)

type (
	Refreshable = gs.Refreshable
	Dync[T any] = gs_dync.Value[T]
)

type (
	BeanInit    = gs_bean.BeanInit
	BeanDestroy = gs_bean.BeanDestroy
)

type (
	RegisteredBean = gs.RegisteredBean
	BeanDefinition = gs.BeanDefinition
)

// NewBean creates a new BeanDefinition.
var NewBean = gs_core.NewBean

// TagBeanSelector creates a BeanSelector based on a tag.
func TagBeanSelector(tag string) BeanSelector {
	return BeanSelector{Tag: tag}
}

// TypeBeanSelector creates a BeanSelector based on a type.
func TypeBeanSelector[T any]() BeanSelector {
	return BeanSelector{Type: reflect.TypeFor[T]()}
}

/************************************ boot ***********************************/

var boot *gs_app.Boot

// bootRun runs the boot process.
func bootRun() error {
	if boot != nil {
		if err := boot.Run(); err != nil {
			return err
		}
		boot = nil
	}
	return nil
}

// Boot initializes and returns a [gs_app.Boot] instance.
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

// Start starts the app, usually for testing purposes.
func Start() error {
	return app.Start()
}

// Stop stops the app, usually for testing purposes.
func Stop() {
	app.Stop()
}

// Run runs the app and waits for an interrupt signal to exit.
func Run() error {
	printBanner()
	if err := bootRun(); err != nil {
		return err
	}
	return app.Run()
}

// ShutDown shuts down the app with an optional message.
func ShutDown(msg ...string) {
	app.ShutDown(msg...)
}

// Config returns the app configuration.
func Config() *gs_conf.AppConfig {
	return app.P
}

// Object registers a bean definition for a given object.
func Object(i interface{}) *RegisteredBean {
	b := NewBean(reflect.ValueOf(i))
	return app.C.Register(b)
}

// Provide registers a bean definition for a given constructor.
func Provide(ctor interface{}, args ...Arg) *RegisteredBean {
	b := NewBean(ctor, args...)
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

// Runner registers a bean definition for an [AppRunner].
func Runner(objOrCtor interface{}, ctorArgs ...Arg) *RegisteredBean {
	b := NewBean(objOrCtor, ctorArgs...).Export(
		reflect.TypeFor[AppRunner](),
	)
	return app.C.Register(b)
}

// Server registers a bean definition for an [AppServer].
func Server(objOrCtor interface{}, ctorArgs ...Arg) *RegisteredBean {
	b := NewBean(objOrCtor, ctorArgs...).Export(
		reflect.TypeFor[AppServer](),
	)
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

// Banner sets a custom app banner.
func Banner(banner string) {
	appBanner = banner
}

// printBanner prints the app banner.
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

// AllowCircularReferences enables or disables circular references between beans.
func AllowCircularReferences(enable bool) {
	err := sysconf.Set("spring.allow-circular-references", enable)
	_ = err // Ignore error
}

// ForceAutowireIsNullable forces autowire to be nullable.
func ForceAutowireIsNullable(enable bool) {
	err := sysconf.Set("spring.force-autowire-is-nullable", enable)
	_ = err // Ignore error
}
