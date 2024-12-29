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
	"testing"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs"
	"github.com/go-spring/spring-core/gs/arg"
	"github.com/go-spring/spring-core/gs/gsutil"
)

var ctx gs.Context

// Init 初始化测试环境
func Init() error {
	c, err := gs.Start()
	if err != nil {
		return err
	}
	ctx = c
	return nil
}

// Run 运行测试用例
func Run(m *testing.M) (code int) {
	defer func() { gs.Stop() }()
	return m.Run()
}

// HasProperty 判断属性是否存在
func HasProperty(key string) bool {
	return ctx.Has(key)
}

// GetProperty 获取属性值
func GetProperty(key string, opts ...conf.GetOption) string {
	return ctx.Prop(key, opts...)
}

// Resolve 解析字符串
func Resolve(s string) (string, error) {
	return ctx.Resolve(s)
}

// Bind 绑定对象
func Bind(i interface{}, opts ...conf.BindArg) error {
	return ctx.Bind(i, opts...)
}

// Get 获取对象
func Get(i interface{}, selectors ...gsutil.BeanSelector) error {
	return ctx.Get(i, selectors...)
}

// Wire 注入对象
func Wire(objOrCtor interface{}, ctorArgs ...arg.Arg) (interface{}, error) {
	return ctx.Wire(objOrCtor, ctorArgs...)
}

// Invoke 调用函数
func Invoke(fn interface{}, args ...arg.Arg) ([]interface{}, error) {
	return ctx.Invoke(fn, args...)
}
