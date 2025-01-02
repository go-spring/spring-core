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
	"github.com/go-spring/spring-core/gs/internal/gs_ctx"
)

const (
	Version = "go-spring@v1.1.3"
	Website = "https://go-spring.com/"
)

type (
	BeanSelector = gs.BeanSelector
	Condition    = gs.Condition
	CondContext  = gs.CondContext
	Arg          = gs.Arg
	Context      = gs.Context
	GroupFunc    = gs.GroupFunc
)

/************************************ boot ***********************************/

var boot *gs_app.Boot

func runBoot() error {
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
func Start() (gs.Context, error) {
	printBanner()
	err := runBoot()
	if err != nil {
		return nil, err
	}
	return app.Start()
}

// Stop 停止程序。
func Stop() {
	app.Stop()
}

// Run 启动程序。
func Run() error {
	printBanner()
	err := runBoot()
	if err != nil {
		return err
	}
	return app.Run()
}

// ShutDown 停止程序。
func ShutDown(msg ...string) {
	app.ShutDown(msg...)
}

// Object 参考 Container.Object 的解释。
func Object(i interface{}) *gs.BeanDefinition {
	return app.Accept(gs_ctx.NewBean(reflect.ValueOf(i)))
}

// Provide 参考 Container.Provide 的解释。
func Provide(ctor interface{}, args ...gs.Arg) *gs.BeanDefinition {
	return app.Accept(gs_ctx.NewBean(ctor, args...))
}

// Accept 参考 Container.Accept 的解释。
func Accept(b *gs.BeanDefinition) *gs.BeanDefinition {
	return app.Accept(b)
}

func Group(fn GroupFunc) {
	app.Group(fn)
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
