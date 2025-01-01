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
	"github.com/go-spring/spring-core/gs/internal/gs_conf"
	"github.com/go-spring/spring-core/gs/internal/gs_ctx"
)

type BootRunner interface {
	Run(ctx gs.Context)
}

type Boot struct {
	c *gs_ctx.Container
	p *gs_conf.Bootstrap
}

func newBoot() *Boot {
	b := &Boot{
		c: gs_ctx.New(),
		p: gs_conf.NewBootstrap(),
	}
	b.c.Object(b)
	return b
}

// Object 参考 Container.Object 的解释。
func (b *Boot) Object(i interface{}) *gs.BeanDefinition {
	return b.c.Accept(gs_ctx.NewBean(reflect.ValueOf(i)))
}

// Provide 参考 Container.Provide 的解释。
func (b *Boot) Provide(ctor interface{}, args ...gs.Arg) *gs.BeanDefinition {
	return b.c.Accept(gs_ctx.NewBean(ctor, args...))
}

// Accept 参考 Container.Accept 的解释。
func (b *Boot) Accept(bd *gs.BeanDefinition) *gs.BeanDefinition {
	return b.c.Accept(bd)
}

func (b *Boot) Group(fn gs.GroupFunc) {
	b.c.Group(fn)
}

func (b *Boot) run() error {

	p, err := b.p.Refresh()
	if err != nil {
		return err
	}

	err = b.c.RefreshProperties(p)
	if err != nil {
		return err
	}

	err = b.c.Refresh()
	if err != nil {
		return err
	}

	var runners []AppRunner
	err = b.c.Get(&runners, "${spring.boot.runners:=*?}")
	if err != nil {
		return err
	}

	// 执行命令行启动器
	for _, r := range runners {
		r.Run(b.c)
	}

	b.c.Close()
	return nil
}
