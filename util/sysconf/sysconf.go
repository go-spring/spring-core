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

/*
Package sysconf provides a unified configuration container for the Go programming language.

In the Go programming language, unlike many other languages,
the standard library lacks a unified and general-purpose configuration container.
To address this gap, go-spring introduces a powerful configuration system that supports
layered configuration management and flexible injection.

So sysconf serves as the fallback configuration container within an application,
acting as the lowest-level foundation of the configuration system.
It can be used independently or as a lightweight alternative or supplement to other
configuration sources such as environment variables, command-line arguments, or configuration files.
*/
package sysconf

import (
	"sync"

	"github.com/go-spring/spring-core/conf"
)

var (
	prop = conf.New()
	lock sync.Mutex
)

// Has returns whether the key exists.
func Has(key string) bool {
	lock.Lock()
	defer lock.Unlock()
	return prop.Has(key)
}

// Get returns the property of the key.
func Get(key string) string {
	lock.Lock()
	defer lock.Unlock()
	return prop.Get(key)
}

// Set sets the property of the key.
func Set(key string, val string) error {
	lock.Lock()
	defer lock.Unlock()
	return prop.Set(key, val)
}

// Clear clears all properties.
func Clear() {
	lock.Lock()
	defer lock.Unlock()
	prop = conf.New()
}

// Clone copies all properties into another properties.
func Clone() *conf.MutableProperties {
	lock.Lock()
	defer lock.Unlock()
	p := conf.New()
	err := prop.CopyTo(p)
	_ = err // should no error
	return p
}
