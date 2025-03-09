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

// Boot defines the interface for application bootstrapping.
type Boot interface {
	Config() *gs_conf.BootConfig
	Object(i interface{}) *gs.RegisteredBean
	Provide(ctor interface{}, args ...gs.Arg) *gs.RegisteredBean
	Register(bd *gs.BeanDefinition) *gs.RegisteredBean
}

// boot is the bootstrapper of the application.
type boot struct {
	c *gs_core.Container
	p *gs_conf.BootConfig

	Runners []gs.Runner `autowire:"${spring.boot.runners:=*?}"`
}

// NewBoot creates a new Boot instance.
func NewBoot() Boot {
	return &boot{
		c: gs_core.New(),
		p: gs_conf.NewBootConfig(),
	}
}

// Config returns the boot configuration.
func (b *boot) Config() *gs_conf.BootConfig {
	return b.p
}

// Object registers an object bean.
func (b *boot) Object(i interface{}) *gs.RegisteredBean {
	bd := gs_core.NewBean(reflect.ValueOf(i))
	return b.c.Register(bd)
}

// Provide registers a bean using a constructor function.
func (b *boot) Provide(ctor interface{}, args ...gs.Arg) *gs.RegisteredBean {
	bd := gs_core.NewBean(ctor, args...)
	return b.c.Register(bd)
}

// Register registers a BeanDefinition instance.
func (b *boot) Register(bd *gs.BeanDefinition) *gs.RegisteredBean {
	return b.c.Register(bd)
}

// Run executes the application's boot process.
func (b *boot) Run() error {
	b.c.Object(b)

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

	// Execute all registered runners.
	for _, r := range b.Runners {
		if err := r.Run(); err != nil {
			return err
		}
	}

	b.c.Close()
	return nil
}
