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
	"reflect"
	"runtime"

	"github.com/go-spring/log"
	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_app"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/gs/internal/gs_conf"
	"github.com/go-spring/spring-core/gs/internal/gs_dync"
)

const (
	Version = "go-spring@v1.2.2"
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
	Condition           = gs.Condition
	ConditionContext    = gs.ConditionContext
	ConditionOnProperty = gs_cond.ConditionOnProperty
)

// OnOnce creates a Condition that wraps another Condition and ensures
// its Matches method is called only once. Subsequent calls will return
// the same result as the first call without re-evaluating the condition.
func OnOnce(conditions ...Condition) Condition {
	var (
		done   bool
		result bool
	)
	return OnFunc(func(ctx ConditionContext) (_ bool, err error) {
		if done {
			return result, nil
		}
		done = true
		result, err = gs_cond.And(conditions...).Matches(ctx)
		return result, err
	})
}

// OnFunc creates a Condition based on the provided function.
func OnFunc(fn func(ctx ConditionContext) (bool, error)) Condition {
	return gs_cond.OnFunc(fn)
}

// OnProperty creates a Condition based on a property name and options.
func OnProperty(name string) ConditionOnProperty {
	return gs_cond.OnProperty(name)
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

// OnEnableJobs creates a Condition that checks whether the EnableJobsProp property is true.
func OnEnableJobs() ConditionOnProperty {
	return OnProperty(EnableJobsProp).HavingValue("true").MatchIfMissing()
}

// OnEnableServers creates a Condition that checks whether the EnableServersProp property is true.
func OnEnableServers() ConditionOnProperty {
	return OnProperty(EnableServersProp).HavingValue("true").MatchIfMissing()
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
	_, file, _, _ := runtime.Caller(1)
	fileID := gs_conf.SysConf.AddFile(file)
	if err := gs_conf.SysConf.Set(key, val, fileID); err != nil {
		log.Errorf(context.Background(), log.TagAppDef, "failed to set property key=%s, err=%v", key, err)
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

// Web enables or disables the built-in web server.
func Web(enable bool) *AppStarter {
	EnableSimpleHttpServer(enable)
	return &AppStarter{}
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

// Module registers a module.
func Module(conditions []ConditionOnProperty, fn func(p conf.Properties) error) {
	gs_app.GS.C.Module(conditions, fn)
}

// RefreshProperties refreshes the app configuration.
func RefreshProperties() error {
	p, err := gs_app.GS.P.Refresh()
	if err != nil {
		return err
	}
	return gs_app.GS.C.RefreshProperties(p)
}
