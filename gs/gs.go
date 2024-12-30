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

	"github.com/go-spring/spring-core/gs/internal/gs_app"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_ctx"
)

var app = gs_app.NewApp()

// Start 启动程序。
func Start() (gs_ctx.Context, error) {
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

// Bootstrap 参考 App.Bootstrap 的解释。
func Bootstrap() *gs_app.Bootstrapper {
	return app.Bootstrap()
}

// OnProperty 参考 App.OnProperty 的解释。
func OnProperty(key string, fn interface{}) {
	app.OnProperty(key, fn)
}

// Property 参考 Container.Property 的解释。
func Property(key string, value interface{}) {
	app.Property(key, value)
}

// Accept 参考 Container.Accept 的解释。
func Accept(b *gs_bean.BeanDefinition) *gs_bean.BeanDefinition {
	return app.Accept(b)
}

// Object 参考 Container.Object 的解释。
func Object(i interface{}) *gs_bean.BeanDefinition {
	return app.Accept(gs_ctx.NewBean(reflect.ValueOf(i)))
}

// Provide 参考 Container.Provide 的解释。
func Provide(ctor interface{}, args ...gs_arg.Arg) *gs_bean.BeanDefinition {
	return app.Accept(gs_ctx.NewBean(ctor, args...))
}
