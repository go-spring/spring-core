/*
 * Copyright 2024 The Go-Spring Authors.
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

// Bootstrap is the bootstrapper of the application.
type Bootstrap struct {
	c *gs_core.Container
	p *gs_conf.BootConfig

	Runners []gs.Runner `autowire:"${spring.boot.runners:=*?}"`
}

// NewBootstrap creates a new Bootstrap instance.
func NewBootstrap() *Bootstrap {
	b := &Bootstrap{
		c: gs_core.New(),
		p: gs_conf.NewBootConfig(),
	}
	b.c.Register(gs_core.NewBean(b))
	return b
}

// Config returns the boot configuration.
func (b *Bootstrap) Config() *gs_conf.BootConfig {
	return b.p
}

// Object registers an object bean.
func (b *Bootstrap) Object(i interface{}) *gs.RegisteredBean {
	bd := gs_core.NewBean(reflect.ValueOf(i))
	return b.c.Register(bd)
}

// Provide registers a bean using a constructor function.
func (b *Bootstrap) Provide(ctor interface{}, args ...gs.Arg) *gs.RegisteredBean {
	bd := gs_core.NewBean(ctor, args...)
	return b.c.Register(bd)
}

// Register registers a BeanDefinition instance.
func (b *Bootstrap) Register(bd *gs.BeanDefinition) *gs.RegisteredBean {
	return b.c.Register(bd)
}

// GroupRegister registers a group of BeanDefinitions.
func (b *Bootstrap) GroupRegister(fn func(p gs.Properties) ([]*gs.BeanDefinition, error)) {
	b.c.GroupRegister(fn)
}

// Run executes the application's bootstrap process.
func (b *Bootstrap) Run() error {
	// Refresh the boot configuration.
	p, err := b.p.Refresh()
	if err != nil {
		return err
	}

	// Refresh properties in the container.
	err = b.c.RefreshProperties(p)
	if err != nil {
		return err
	}

	// Refresh the container.
	err = b.c.Refresh()
	if err != nil {
		return err
	}

	// Execute all registered AppRunners.
	for _, r := range b.Runners {
		r.Run()
	}

	b.c.Close()
	return nil
}
