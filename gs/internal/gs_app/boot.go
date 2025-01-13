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

// Boot is the bootstrapper of the application.
type Boot struct {
	c gs.Container
	p *gs_conf.BootConfig

	Runners []AppRunner `autowire:"${spring.boot.runners:=*?}"`
}

// NewBoot creates a new boot instance.
func NewBoot() *Boot {
	b := &Boot{
		c: gs_core.New(),
		p: gs_conf.NewBootConfig(),
	}
	b.c.Object(b)
	return b
}

// Config returns the boot config.
func (b *Boot) Config() *gs_conf.BootConfig {
	return b.p
}

// Object registers a bean by instance.
func (b *Boot) Object(i interface{}) *gs.RegisteredBean {
	bd := gs_core.NewBean(reflect.ValueOf(i))
	return b.c.Register(bd)
}

// Provide registers a bean by constructor.
func (b *Boot) Provide(ctor interface{}, args ...gs.Arg) *gs.RegisteredBean {
	bd := gs_core.NewBean(ctor, args...)
	return b.c.Register(bd)
}

// Register registers a [gs.BeanDefinition].
func (b *Boot) Register(bd *gs.BeanDefinition) *gs.RegisteredBean {
	return b.c.Register(bd)
}

// Group registers a group of [gs.BeanDefinition].
func (b *Boot) Group(fn func(p gs.Properties) ([]*gs.BeanDefinition, error)) {
	b.c.Group(fn)
}

// Runner registers an [AppRunner] by instance or constructor.
func (b *Boot) Runner(objOrCtor interface{}, ctorArgs ...gs.Arg) *gs.RegisteredBean {
	bd := gs_core.NewBean(objOrCtor, ctorArgs...)
	bd.Export((*AppRunner)(nil))
	return b.c.Register(bd)
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
		r.Run(&AppContext{b.c.(gs.Context)})
	}

	b.c.Close()
	return nil
}
