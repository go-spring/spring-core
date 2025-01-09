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

type (
	Arg              = gs.Arg
	BeanDefinition   = gs.BeanDefinition
	BeanInit         = gs.BeanInit
	BeanDestroy      = gs.BeanDestroy
	BeanRegistration = gs.BeanRegistration
	BeanSelector     = gs.BeanSelector
	CondContext      = gs.CondContext
	Condition        = gs.Condition
	Context          = gs.Context
	ContextAware     = gs_core.ContextAware
	Dync[T any]      = gs_dync.Value[T]
	Runner           = gs_app.Runner
	Server           = gs_app.Server
)

/************************************ arg ***********************************/

// IndexArg returns an IndexArg.
func IndexArg(n int, arg Arg) gs_arg.IndexArg {
	return gs_arg.Index(n, arg)
}

// R0 returns an IndexArg with index 0.
func R0(arg Arg) gs_arg.IndexArg { return gs_arg.R0(arg) }

// R1 returns an IndexArg with index 1.
func R1(arg Arg) gs_arg.IndexArg { return gs_arg.R1(arg) }

// R2 returns an IndexArg with index 2.
func R2(arg Arg) gs_arg.IndexArg { return gs_arg.R2(arg) }

// R3 returns an IndexArg with index 3.
func R3(arg Arg) gs_arg.IndexArg { return gs_arg.R3(arg) }

// R4 returns an IndexArg with index 4.
func R4(arg Arg) gs_arg.IndexArg { return gs_arg.R4(arg) }

// R5 returns an IndexArg with index 5.
func R5(arg Arg) gs_arg.IndexArg { return gs_arg.R5(arg) }

// R6 returns an IndexArg with index 6.
func R6(arg Arg) gs_arg.IndexArg { return gs_arg.R6(arg) }

// NilArg return a ValueArg with a value of nil.
func NilArg() gs_arg.ValueArg {
	return gs_arg.Nil()
}

// ValueArg return a ValueArg with a value of v.
func ValueArg(v interface{}) gs_arg.ValueArg {
	return gs_arg.Value(v)
}

// OptionArg 返回 Option 函数的参数绑定。
func OptionArg(fn interface{}, args ...Arg) *gs_arg.OptionArg {
	return gs_arg.Option(fn, args...)
}

// MustBindArg 为 Option 方法绑定运行时参数。
func MustBindArg(fn interface{}, args ...Arg) *gs_arg.Callable {
	return gs_arg.MustBind(fn, args...)
}

func BindArg(fn interface{}, args []Arg, skip int) (*gs_arg.Callable, error) {
	return gs_arg.Bind(fn, args, skip)
}

/************************************ cond ***********************************/

type (
	Conditional    = gs_cond.Conditional
	PropertyOption = gs_cond.PropertyOption
)

// OK returns a Condition that always returns true.
func OK() Condition {
	return gs_cond.OK()
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

func MatchIfMissing() PropertyOption {
	return gs_cond.MatchIfMissing()
}

func HavingValue(havingValue string) PropertyOption {
	return gs_cond.HavingValue(havingValue)
}

func OnProperty(name string, options ...PropertyOption) *Conditional {
	return gs_cond.OnProperty(name, options...)
}

func OnMissingProperty(name string) *Conditional {
	return gs_cond.OnMissingProperty(name)
}

func OnBean(selector BeanSelector) *Conditional {
	return gs_cond.OnBean(selector)
}

func OnMissingBean(selector BeanSelector) *Conditional {
	return gs_cond.OnMissingBean(selector)
}

func OnSingleBean(selector BeanSelector) *Conditional {
	return gs_cond.OnSingleBean(selector)
}

func OnExpression(expression string) *Conditional {
	return gs_cond.OnExpression(expression)
}

func OnMatches(fn func(ctx CondContext) (bool, error)) *Conditional {
	return gs_cond.OnMatches(fn)
}

func OnProfile(profile string) *Conditional {
	return gs_cond.OnProfile(profile)
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

// Boot 参考 App.Boot 的解释。
func Boot() *gs_app.Boot {
	if boot == nil {
		boot = gs_app.NewBoot()
	}
	return boot
}

/*********************************** app *************************************/

var app = gs_app.NewApp()

// Start 启动程序。
func Start() error {
	return app.Start()
}

// Stop 停止程序。
func Stop() {
	app.Stop()
}

// Run 启动程序。
func Run() error {
	printBanner()
	if err := bootRun(); err != nil {
		return err
	}
	return app.Run()
}

// ShutDown 停止程序。
func ShutDown(msg ...string) {
	app.ShutDown(msg...)
}

func AppConfig() *gs_conf.AppConfig {
	return app.P
}

func AppRunner(objOrCtor interface{}, ctorArgs ...gs.Arg) *BeanRegistration {
	b := gs_core.NewBean(objOrCtor, ctorArgs...)
	b.Export((*gs_app.Runner)(nil))
	return app.C.Accept(b)
}

func AppServer(objOrCtor interface{}, ctorArgs ...gs.Arg) *BeanRegistration {
	b := gs_core.NewBean(objOrCtor, ctorArgs...)
	b.Export((*gs_app.Server)(nil))
	return app.C.Accept(b)
}

// Object 参考 Container.Object 的解释。
func Object(i interface{}) *BeanRegistration {
	b := gs_core.NewBean(reflect.ValueOf(i))
	return app.C.Accept(b)
}

// Provide 参考 Container.Provide 的解释。
func Provide(ctor interface{}, args ...Arg) *BeanRegistration {
	b := gs_core.NewBean(ctor, args...)
	return app.C.Accept(b)
}

// Accept 参考 Container.Accept 的解释。
func Accept(b *BeanDefinition) *BeanRegistration {
	return app.C.Accept(b)
}

func Group(fn func(p gs.Properties) ([]*BeanDefinition, error)) {
	app.C.Group(fn)
}

func RefreshProperties(p gs.Properties) error {
	return app.C.RefreshProperties(p)
}

/********************************** banner ***********************************/

var appBanner = `
                                              (_)              
  __ _    ___             ___   _ __    _ __   _   _ __     __ _ 
 / _' |  / _ \   ______  / __| | '_ \  | '__| | | | '_ \   / _' |
| (_| | | (_) | |______| \__ \ | |_) | | |    | | | | | | | (_| |
 \__, |  \___/           |___/ | .__/  |_|    |_| |_| |_|  \__, |
  __/ |                        | |                          __/ |
 |___/                         |_|                         |___/ 
`

// Banner 自定义 banner 字符串。
func Banner(banner string) {
	appBanner = banner
}

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

func AllowCircularReferences(allow bool) {
	err := sysconf.Set("spring.allow-circular-references", allow)
	_ = err // ignore error
}
