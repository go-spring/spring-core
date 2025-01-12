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

package gstest

import (
	"context"
	"testing"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs"
)

var ctx = &gs.ContextAware{}

// Init 初始化测试环境
func Init() error {
	gs.Object(ctx)
	return gs.Start()
}

// Run 运行测试用例
func Run(m *testing.M) (code int) {
	defer func() { gs.Stop() }()
	return m.Run()
}

func Context() context.Context {
	return ctx.GSContext.Context()
}

func Keys() []string {
	return ctx.GSContext.Keys()
}

// Has 判断属性是否存在
func Has(key string) bool {
	return ctx.GSContext.Has(key)
}

func SubKeys(key string) ([]string, error) {
	return ctx.GSContext.SubKeys(key)
}

// Prop 获取属性值
func Prop(key string, opts ...conf.GetOption) string {
	return ctx.GSContext.Prop(key, opts...)
}

// Resolve 解析字符串
func Resolve(s string) (string, error) {
	return ctx.GSContext.Resolve(s)
}

// Bind 绑定对象
func Bind(i interface{}, opts ...conf.BindArg) error {
	return ctx.GSContext.Bind(i, opts...)
}

// Get 获取对象
func Get(i interface{}, selectors ...gs.BeanSelector) error {
	return ctx.GSContext.Get(i, selectors...)
}

// Wire 注入对象
func Wire(objOrCtor interface{}, ctorArgs ...gs.Arg) (interface{}, error) {
	return ctx.GSContext.Wire(objOrCtor, ctorArgs...)
}

// Invoke 调用函数
func Invoke(fn interface{}, args ...gs.Arg) ([]interface{}, error) {
	return ctx.GSContext.Invoke(fn, args...)
}

func RefreshProperties(p gs.Properties) error {
	return gs.RefreshProperties(p)
}
