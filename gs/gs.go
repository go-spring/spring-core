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
	"reflect"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_app"
	"github.com/go-spring/spring-core/gs/internal/gs_ctx"
)

type (
	Arg          = gs.Arg
	Context      = gs.Context
	BeanSelector = gs.BeanSelector
	GroupFunc    = gs.GroupFunc
)

var app = gs_app.NewApp()

// Start 启动程序。
func Start() (gs.Context, error) {
	return app.Start()
}

// Stop 停止程序。
func Stop() {
	app.Stop()
}

// Run 启动程序。
func Run() error {
	return app.Run()
}

// ShutDown 停止程序。
func ShutDown(msg ...string) {
	app.ShutDown(msg...)
}

// Boot 参考 App.Boot 的解释。
func Boot() *gs_app.Boot {
	return app.Boot()
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
