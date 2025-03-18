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

package gs

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_app"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/gs/internal/gs_conf"
	"github.com/go-spring/spring-core/gs/internal/gs_core"
	"github.com/go-spring/spring-core/gs/internal/gs_dync"
)

const (
	Version = "go-spring@v1.2.0.rc"
	Website = "https://go-spring.com/"
)

// As returns the [reflect.Type] of the given interface type.
func As[T any]() reflect.Type {
	return gs.As[T]()
}

/************************************ arg ***********************************/

type Arg = gs.Arg

// TagArg returns a TagArg with the specified tag.
func TagArg(tag string) Arg {
	return gs_arg.TagArg{Tag: tag}
}

// NilArg returns a ValueArg with a nil value.
func NilArg() Arg {
	return gs_arg.Nil()
}

// ValueArg returns a ValueArg with the specified value.
func ValueArg(v interface{}) Arg {
	return gs_arg.Value(v)
}

// IndexArg returns an IndexArg with the specified index and argument.
func IndexArg(n int, arg Arg) Arg {
	return gs_arg.Index(n, arg)
}

// BindArg returns an BindArg for the specified function and arguments.
func BindArg(fn interface{}, args ...Arg) *gs_arg.BindArg {
	return gs_arg.Bind(fn, args...)
}

/************************************ cond ***********************************/

type (
	CondFunc    = gs.CondFunc
	Condition   = gs.Condition
	CondContext = gs.CondContext
)

// OnFunc creates a Condition based on the provided function.
func OnFunc(fn CondFunc) Condition {
	return gs_cond.OnFunc(fn)
}

// OnProperty creates a Condition based on a property name and options.
func OnProperty(name string) gs_cond.OnPropertyInterface {
	return gs_cond.OnProperty(name)
}

// OnMissingProperty creates a Condition that checks for a missing property.
func OnMissingProperty(name string) Condition {
	return gs_cond.OnMissingProperty(name)
}

// OnBean creates a Condition based on a BeanSelector.
func OnBean[T any](name ...string) Condition {
	return gs_cond.OnBeanSelector(BeanSelectorFor[T](name...))
}

// OnBeanSelector creates a Condition based on a BeanSelector.
func OnBeanSelector(s BeanSelector) Condition {
	return gs_cond.OnBeanSelector(s)
}

// OnMissingBean creates a Condition for when a specific bean is missing.
func OnMissingBean[T any](name ...string) Condition {
	return gs_cond.OnMissingBeanSelector(BeanSelectorFor[T](name...))
}

// OnMissingBeanSelector creates a Condition for when a specific bean is missing.
func OnMissingBeanSelector(s BeanSelector) Condition {
	return gs_cond.OnMissingBeanSelector(s)
}

// OnSingleBean creates a Condition for when only one instance of a bean exists.
func OnSingleBean[T any](name ...string) Condition {
	return gs_cond.OnSingleBeanSelector(BeanSelectorFor[T](name...))
}

// OnSingleBeanSelector creates a Condition for when only one instance of a bean exists.
func OnSingleBeanSelector(s BeanSelector) Condition {
	return gs_cond.OnSingleBeanSelector(s)
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
func Or(conditions ...Condition) Condition {
	return gs_cond.Or(conditions...)
}

// And creates a Condition that is true if all the given Conditions are true.
func And(conditions ...Condition) Condition {
	return gs_cond.And(conditions...)
}

// None creates a Condition that is true if none of the given Conditions are true.
func None(conditions ...Condition) Condition {
	return gs_cond.None(conditions...)
}

/************************************ ioc ************************************/

type (
	Refreshable = gs.Refreshable
	Dync[T any] = gs_dync.Value[T]
)

type (
	RegisteredBean = gs.RegisteredBean
	BeanDefinition = gs.BeanDefinition
)

type (
	BeanSelector    = gs.BeanSelector
	BeanInitFunc    = gs.BeanInitFunc
	BeanDestroyFunc = gs.BeanDestroyFunc
)

// NewBean creates a new BeanDefinition.
var NewBean = gs_core.NewBean

// BeanSelectorFor returns a BeanSelector for the given type.
func BeanSelectorFor[T any](name ...string) BeanSelector {
	return gs.BeanSelectorFor[T](name...)
}

/************************************ boot ***********************************/

var boot gs_app.Boot

// Boot initializes and returns a [*gs_app.Boot] instance.
func Boot() gs_app.Boot {
	if boot == nil {
		boot = gs_app.NewBoot()
	}
	return boot
}

/*********************************** app *************************************/

type (
	Runner      = gs.Runner
	Job         = gs.Job
	Server      = gs.Server
	ReadySignal = gs.ReadySignal
)

type FuncRunner func() error

func (f FuncRunner) Run() error {
	return f()
}

type FuncJob func(ctx context.Context) error

func (f FuncJob) Run(ctx context.Context) error {
	return f(ctx)
}

// Run runs the app and waits for an interrupt signal to exit.
func Run() error {
	printBanner()
	if boot != nil {
		if err := boot.(interface{ Run() error }).Run(); err != nil {
			return err
		}
		boot = nil
	}
	return gs_app.GS.Run()
}

// Exiting returns a boolean indicating whether the application is exiting.
func Exiting() bool {
	return gs_app.GS.Exiting()
}

// ShutDown shuts down the app with an optional message.
func ShutDown() {
	gs_app.GS.ShutDown()
}

// Config returns the app configuration.
func Config() *gs_conf.AppConfig {
	return gs_app.GS.P
}

// Object registers a bean definition for a given object.
func Object(i interface{}) *RegisteredBean {
	b := NewBean(reflect.ValueOf(i))
	return gs_app.GS.C.Register(b)
}

// Provide registers a bean definition for a given constructor.
func Provide(ctor interface{}, args ...Arg) *RegisteredBean {
	b := NewBean(ctor, args...)
	return gs_app.GS.C.Register(b)
}

// Register registers a bean definition.
func Register(b *BeanDefinition) *RegisteredBean {
	return gs_app.GS.C.Register(b)
}

// GroupRegister registers a group of bean definitions.
func GroupRegister(fn func(p conf.ReadOnlyProperties) ([]*BeanDefinition, error)) {
	gs_app.GS.C.GroupRegister(fn)
}

// RefreshProperties refreshes the app configuration.
func RefreshProperties() error {
	p, err := Config().Refresh()
	if err != nil {
		return err
	}
	return gs_app.GS.C.RefreshProperties(p)
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
