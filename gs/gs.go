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
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/gs/internal/gs_conf"
	"github.com/go-spring/spring-core/gs/internal/gs_dync"
)

const (
	Version = "go-spring@v1.2.3"
	Website = "https://github.com/go-spring/"
)

// Dync is a generic alias for dynamic configuration values.
type Dync[T any] = gs_dync.Value[T]

// BeanSelector is an alias for gs.BeanSelector used to locate beans.
type BeanSelector = gs.BeanSelector

// BeanSelectorFor creates a BeanSelector for the given type and optional name.
func BeanSelectorFor[T any](name ...string) BeanSelector {
	return gs.BeanSelectorFor[T](name...)
}

// As returns the [reflect.Type] for the given generic type.
func As[T any]() reflect.Type {
	return gs.As[T]()
}

/************************************ arg ***********************************/

type Arg = gs.Arg

// TagArg creates an argument that injects a property or object by tag.
func TagArg(tag string) Arg {
	return gs_arg.Tag(tag)
}

// ValueArg creates an argument with a fixed value.
func ValueArg(v any) Arg {
	return gs_arg.Value(v)
}

// IndexArg creates an argument for a specific constructor parameter index.
func IndexArg(n int, arg Arg) Arg {
	return gs_arg.Index(n, arg)
}

// BindArg binds arguments dynamically to an option-style constructor.
func BindArg(fn any, args ...Arg) *gs_arg.BindArg {
	return gs_arg.Bind(fn, args...)
}

/************************************ cond ***********************************/

type (
	Condition           = gs.Condition
	ConditionContext    = gs.ConditionContext
	ConditionOnProperty = gs_cond.ConditionOnProperty
)

// OnOnce wraps a condition so it is evaluated only once.
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

// OnFunc creates a condition from a function.
func OnFunc(fn func(ctx ConditionContext) (bool, error)) Condition {
	return gs_cond.OnFunc(fn)
}

// OnProperty creates a property-based condition.
func OnProperty(name string) ConditionOnProperty {
	return gs_cond.OnProperty(name)
}

// OnBean requires a bean to exist.
func OnBean[T any](name ...string) Condition {
	return gs_cond.OnBean[T](name...)
}

// OnMissingBean requires a bean to be missing.
func OnMissingBean[T any](name ...string) Condition {
	return gs_cond.OnMissingBean[T](name...)
}

// OnSingleBean requires only one instance of a bean.
func OnSingleBean[T any](name ...string) Condition {
	return gs_cond.OnSingleBean[T](name...)
}

// RegisterExpressFunc registers a custom expression function for conditions.
func RegisterExpressFunc(name string, fn any) {
	gs_cond.RegisterExpressFunc(name, fn)
}

// OnExpression creates a condition from an expression.
func OnExpression(expression string) Condition {
	return gs_cond.OnExpression(expression)
}

// Not negates a condition.
func Not(c Condition) Condition {
	return gs_cond.Not(c)
}

// Or combines conditions using logical OR.
func Or(conditions ...Condition) Condition {
	return gs_cond.Or(conditions...)
}

// And combines conditions using logical AND.
func And(conditions ...Condition) Condition {
	return gs_cond.And(conditions...)
}

// None creates a condition that is true if all given conditions are false.
func None(conditions ...Condition) Condition {
	return gs_cond.None(conditions...)
}

// OnEnableJobs checks if job execution is enabled.
func OnEnableJobs() ConditionOnProperty {
	return OnProperty(EnableJobsProp).HavingValue("true").MatchIfMissing()
}

// OnEnableServers checks if servers are enabled.
func OnEnableServers() ConditionOnProperty {
	return OnProperty(EnableServersProp).HavingValue("true").MatchIfMissing()
}

/*********************************** app *************************************/

type (
	Server      = gs.Server
	ReadySignal = gs.ReadySignal
)

var (
	// B is the bootstrapper.
	B = gs_app.NewBoot()
	// app is the application.
	app = gs_app.NewApp()
)

// Config returns the application configuration.
func Config() *gs_conf.AppConfig {
	return app.P
}

// Property sets a system property.
func Property(key string, val string) {
	_, file, _, _ := runtime.Caller(1)
	fileID := gs_conf.SysConf.AddFile(file)
	if err := gs_conf.SysConf.Set(key, val, fileID); err != nil {
		log.Errorf(context.Background(), log.TagAppDef, "failed to set property key=%s, err=%v", key, err)
	}
}

// RefreshProperties reloads application properties.
func RefreshProperties() error {
	p, err := app.P.Refresh()
	if err != nil {
		return err
	}
	return app.C.RefreshProperties(p)
}

// RootBean registers a root bean.
func RootBean(b *gs.RegisteredBean) {
	app.C.RootBean(b)
}

// Object registers a bean definition for a given object.
func Object(i any) *gs.RegisteredBean {
	return app.C.Object(i).Caller(1)
}

// Provide registers a bean definition from a constructor.
func Provide(ctor any, args ...Arg) *gs.RegisteredBean {
	return app.C.Provide(ctor, args...).Caller(1)
}

// Module registers a module with property-based conditions.
func Module(conditions []ConditionOnProperty, fn func(p conf.Properties) error) {
	app.C.Module(conditions, fn)
}

// Group registers beans in a group based on configuration properties.
func Group[T any, R any](key string, fn func(c T) (R, error), d func(R) error) {
	app.C.Module([]ConditionOnProperty{
		OnProperty(key),
	}, func(p conf.Properties) error {
		var m map[string]T
		if err := p.Bind(&m, "${"+key+"}"); err != nil {
			return err
		}
		for name, c := range m {
			b := Provide(fn, ValueArg(c)).Name(name)
			if d != nil {
				b.Destroy(d)
			}
		}
		return nil
	})
}

// Runner registers a function as a runner bean.
func Runner(fn func() error) *gs.RegisteredBean {
	return Object(gs.FuncRunner(fn)).AsRunner().Caller(1)
}

// Job registers a function as a job bean.
func Job(fn func(ctx context.Context) error) *gs.RegisteredBean {
	return Object(gs.FuncJob(fn)).AsJob().Caller(1)
}

// Web enables or disables the built-in HTTP server.
func Web(enable bool) *AppStarter {
	EnableSimpleHttpServer(enable)
	return &AppStarter{}
}

// Run starts the application and waits for exit.
func Run() {
	new(AppStarter).Run()
}

// RunWith starts the application with a custom run function.
func RunWith(fn func(ctx context.Context) error) {
	new(AppStarter).RunWith(fn)
}

// RunAsync starts the application asynchronously.
func RunAsync() (func(), error) {
	return new(AppStarter).RunAsync()
}

// Exiting returns true if the application is shutting down.
func Exiting() bool {
	return app.Exiting()
}

// ShutDown gracefully stops the application.
func ShutDown() {
	app.ShutDown()
}
