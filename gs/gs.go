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
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/gs/internal/gs_conf"
	"github.com/go-spring/spring-core/gs/internal/gs_dync"
	"github.com/go-spring/spring-core/log"
)

const (
	Version = "go-spring@v1.2.0"
	Website = "https://go-spring.com/"
)

// As returns the [reflect.Type] of the given interface type.
func As[T any]() reflect.Type {
	return gs.As[T]()
}

/************************************ arg ***********************************/

type Arg = gs.Arg

// TagArg returns a TagArg with the specified tag.
// Used for property binding or object injection when providing constructor parameters.
func TagArg(tag string) Arg {
	return gs_arg.Tag(tag)
}

// ValueArg returns a ValueArg with the specified value.
// Used to provide specific values for constructor parameters.
func ValueArg(v any) Arg {
	return gs_arg.Value(v)
}

// IndexArg returns an IndexArg with the specified index and argument.
// When most constructor parameters can use default values, IndexArg helps reduce configuration effort.
func IndexArg(n int, arg Arg) Arg {
	return gs_arg.Index(n, arg)
}

// BindArg returns a BindArg for the specified function and arguments.
// Used to provide argument binding for option-style constructor parameters.
func BindArg(fn any, args ...Arg) *gs_arg.BindArg {
	return gs_arg.Bind(fn, args...)
}

/************************************ cond ***********************************/

type (
	Condition   = gs.Condition
	CondContext = gs.CondContext
)

// OnFunc creates a Condition based on the provided function.
func OnFunc(fn func(ctx CondContext) (bool, error)) Condition {
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

// OnBean creates a Condition for when a specific bean exists.
func OnBean[T any](name ...string) Condition {
	return gs_cond.OnBean[T](name...)
}

// OnMissingBean creates a Condition for when a specific bean is missing.
func OnMissingBean[T any](name ...string) Condition {
	return gs_cond.OnMissingBean[T](name...)
}

// OnSingleBean creates a Condition for when only one instance of a bean exists.
func OnSingleBean[T any](name ...string) Condition {
	return gs_cond.OnSingleBean[T](name...)
}

// RegisterExpressFunc registers a custom expression function.
func RegisterExpressFunc(name string, fn any) {
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
	BeanID   = gs.BeanID
	BeanMock = gs.BeanMock
)

type (
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
func NewBean(objOrCtor any, ctorArgs ...gs.Arg) *gs.BeanDefinition {
	return gs_bean.NewBean(objOrCtor, ctorArgs...).Caller(1)
}

// BeanSelectorFor returns a BeanSelector for the given type.
func BeanSelectorFor[T any](name ...string) BeanSelector {
	return gs.BeanSelectorFor[T](name...)
}

/*********************************** app *************************************/

// Property sets a system property.
func Property(key string, val string) {
	if err := gs_conf.SysConf.Set(key, val); err != nil {
		log.Errorf(context.Background(), log.TagGS, "failed to set property key=%s, err=%v", key, err)
	}
}

type (
	Runner      = gs.Runner
	Job         = gs.Job
	Server      = gs.Server
	ReadySignal = gs.ReadySignal
)

var B = gs_app.NewBoot()

// funcRunner is a function type that implements the Runner interface.
type funcRunner func() error

func (f funcRunner) Run() error {
	return f()
}

// FuncRunner creates a Runner from a function.
func FuncRunner(fn func() error) *RegisteredBean {
	return Object(funcRunner(fn)).AsRunner().Caller(1)
}

// funcJob is a function type that implements the Job interface.
type funcJob func(ctx context.Context) error

func (f funcJob) Run(ctx context.Context) error {
	return f(ctx)
}

// FuncJob creates a Job from a function.
func FuncJob(fn func(ctx context.Context) error) *RegisteredBean {
	return Object(funcJob(fn)).AsJob().Caller(1)
}

type AppStarter struct{}

// Web enables or disables the built-in web server.
func Web(enable bool) *AppStarter {
	EnableSimpleHttpServer(enable)
	return &AppStarter{}
}

// Run runs the app and waits for an interrupt signal to exit.
func (s *AppStarter) Run() {
	s.RunWith(nil)
}

// RunWith runs the app with a given function and waits for an interrupt signal to exit.
func (s *AppStarter) RunWith(fn func(ctx context.Context) error) {
	var err error
	defer func() {
		if err != nil {
			log.Errorf(context.Background(), log.TagGS, "app run failed: %v", err)
		}
	}()
	printBanner()
	if err = B.(*gs_app.BootImpl).Run(); err != nil {
		return
	}
	B = nil
	err = gs_app.GS.RunWith(fn)
}

// RunAsync runs the app asynchronously and returns a function to stop the app.
func (s *AppStarter) RunAsync() (func(), error) {
	printBanner()
	if err := B.(*gs_app.BootImpl).Run(); err != nil {
		return nil, err
	}
	B = nil
	if err := gs_app.GS.Start(); err != nil {
		return nil, err
	}
	return func() { gs_app.GS.Stop() }, nil
}

// Run runs the app and waits for an interrupt signal to exit.
func Run() {
	new(AppStarter).Run()
}

// RunWith runs the app with a given function and waits for an interrupt signal to exit.
func RunWith(fn func(ctx context.Context) error) {
	new(AppStarter).RunWith(fn)
}

// RunAsync runs the app asynchronously and returns a function to stop the app.
func RunAsync() (func(), error) {
	return new(AppStarter).RunAsync()
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

// Component registers a bean definition for a given object.
func Component[T any](i T) T {
	b := gs_bean.NewBean(reflect.ValueOf(i))
	gs_app.GS.C.Register(b).Caller(1)
	return i
}

// Object registers a bean definition for a given object.
func Object(i any) *RegisteredBean {
	b := gs_bean.NewBean(reflect.ValueOf(i))
	return gs_app.GS.C.Register(b).Caller(1)
}

// Provide registers a bean definition for a given constructor.
func Provide(ctor any, args ...Arg) *RegisteredBean {
	b := gs_bean.NewBean(ctor, args...)
	return gs_app.GS.C.Register(b).Caller(1)
}

// Register registers a bean definition.
func Register(b *BeanDefinition) *RegisteredBean {
	return gs_app.GS.C.Register(b)
}

// GroupRegister registers a group of bean definitions.
func GroupRegister(fn func(p conf.Properties) ([]*BeanDefinition, error)) {
	gs_app.GS.C.GroupRegister(fn)
}

// RefreshProperties refreshes the app configuration.
func RefreshProperties() error {
	p, err := gs_app.GS.P.Refresh()
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
	for s := range strings.SplitSeq(appBanner, "\n") {
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
