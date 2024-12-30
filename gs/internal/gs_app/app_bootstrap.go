/*
 * Copyright 2012-2019 the original author or authors.
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

package gs_app

import (
	"reflect"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_ctx"
)

type Bootstrapper struct {
	c *gs_ctx.Container
}

func newBootstrap() *Bootstrapper {
	return &Bootstrapper{
		c: gs_ctx.New(),
	}
}

func (b *Bootstrapper) Group(fn gs.GroupFunc) {
	b.c.Group(fn)
}

// Object 参考 Container.Object 的解释。
func (b *Bootstrapper) Object(i interface{}) *gs.BeanDefinition {
	return b.c.Accept(gs_ctx.NewBean(reflect.ValueOf(i)))
}

// Provide 参考 Container.Provide 的解释。
func (b *Bootstrapper) Provide(ctor interface{}, args ...gs.Arg) *gs.BeanDefinition {
	return b.c.Accept(gs_ctx.NewBean(ctor, args...))
}

func (b *Bootstrapper) start() error {

	b.c.Object(b)

	// if err := b.loadBootstrap(e); err != nil {
	// 	return err
	// }
	//
	// // 保存从环境变量和命令行解析的属性
	// for _, k := range e.p.Keys() {
	// 	b.c.initProperties.Set(k, e.p.Get(k))
	// }

	return b.c.Refresh(false)
}
