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

package gs_core

import (
	"errors"
	"reflect"
	"testing"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs_core/resolving"
	"github.com/go-spring/spring-core/gs/internal/gs_core/wiring"
	"github.com/go-spring/spring-core/util/syslog"
)

type Container struct {
	*resolving.Resolving
	wiring *wiring.Wiring
}

// New creates a IoC container.
func New() *Container {
	return &Container{
		Resolving: resolving.New(),
	}
}

// Refresh initializes and wires all beans in the container.
func (c *Container) Refresh(p conf.Properties) error {
	if err := c.Resolving.Refresh(p); err != nil {
		return err
	}
	c.wiring = wiring.New(p)
	if err := c.wiring.Refresh(c.Beans()); err != nil {
		return err
	}
	c.Resolving = nil
	return nil
}

// RefreshProperties updates the properties of the container.
func (c *Container) RefreshProperties(p conf.Properties) error {
	return c.wiring.RefreshProperties(p)
}

// Wire injects dependencies into the given object.
func (c *Container) Wire(obj interface{}) error {
	if !testing.Testing() {
		return errors.New("not allowed to call Wire method in non-test mode")
	}
	stack := wiring.NewStack()
	defer func() {
		if len(stack.Beans) > 0 {
			syslog.Infof("wiring path %s", stack.Path())
		}
	}()
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	return c.wiring.WireBeanValue(v, t, false, stack)
}

// Close closes the container and cleans up resources.
func (c *Container) Close() {
	if c.wiring != nil {
		c.wiring.Close()
	}
}
