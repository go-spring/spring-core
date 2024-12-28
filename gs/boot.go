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
	"os"
	"reflect"

	"github.com/go-spring/spring-base/util"
	"github.com/go-spring/spring-core/gs/arg"
)

var app = NewApp()

// Setenv 封装 os.Setenv 函数，如果发生 error 会 panic 。
func Setenv(key string, value string) {
	err := os.Setenv(key, value)
	util.Panic(err).When(err != nil)
}

// Run 启动程序。
func Run() error {
	return app.Run()
}

// ShutDown 停止程序。
func ShutDown(msg ...string) {
	app.ShutDown(msg...)
}

// Banner 参考 App.Banner 的解释。
func Banner(banner string) {
	app.Banner(banner)
}

// Bootstrap 参考 App.Bootstrap 的解释。
func Bootstrap() *bootstrap {
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
func Accept(b *BeanDefinition) *BeanDefinition {
	return app.c.Accept(b)
}

// Object 参考 Container.Object 的解释。
func Object(i interface{}) *BeanDefinition {
	return app.c.Accept(NewBean(reflect.ValueOf(i)))
}

// Provide 参考 Container.Provide 的解释。
func Provide(ctor interface{}, args ...arg.Arg) *BeanDefinition {
	return app.c.Accept(NewBean(ctor, args...))
}
