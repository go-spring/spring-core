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
	"github.com/go-spring/spring-core/gs/internal/gs_core"
)

type BootRunner interface {
	Run(ctx gs.Context)
}

type Boot struct {
	c gs.Container
	p *gs_conf.BootConfig

	Runners []AppRunner `autowire:"${spring.boot.runners:=*?}"`
}

func NewBoot() *Boot {
	b := &Boot{
		c: gs_core.New(),
		p: gs_conf.NewBootConfig(),
	}
	b.c.Object(b)
	return b
}

func (b *Boot) Config() *gs_conf.BootConfig {
	return b.p
}

// Object 参考 Container.Object 的解释。
func (b *Boot) Object(i interface{}) *gs.BeanDefinition {
	return b.c.Accept(gs_core.NewBean(reflect.ValueOf(i)))
}

// Provide 参考 Container.Provide 的解释。
func (b *Boot) Provide(ctor interface{}, args ...gs.Arg) *gs.BeanDefinition {
	return b.c.Accept(gs_core.NewBean(ctor, args...))
}

// Accept 参考 Container.Accept 的解释。
func (b *Boot) Accept(bd *gs.BeanDefinition) *gs.BeanDefinition {
	return b.c.Accept(bd)
}

func (b *Boot) Group(fn gs.GroupFunc) {
	b.c.Group(fn)
}

func (b *Boot) Run() error {

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

	// 执行命令行启动器
	for _, r := range b.Runners {
		r.Run(b.c.(gs.Context))
	}

	b.c.Close()
	return nil
}
