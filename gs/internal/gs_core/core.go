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

// Package gs_core provides the core implementation of the Inversion of Control (IoC)
// container in the Go-Spring framework. It is responsible for managing the lifecycle,
// dependency resolution, and injection of application beans.
package gs_core

import (
	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs_core/injecting"
	"github.com/go-spring/spring-core/gs/internal/gs_core/resolving"
)

type Container struct {
	*resolving.Resolving
	*injecting.Injecting
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
	c.Injecting = injecting.New(p)
	if err := c.Injecting.Refresh(c.Beans()); err != nil {
		return err
	}
	c.Resolving = nil
	return nil
}
