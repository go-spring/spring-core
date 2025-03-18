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

package sysconf

import (
	"sync"

	"github.com/go-spring/spring-core/conf"
)

var (
	prop = conf.New()
	lock sync.Mutex
)

// Get returns the property of the key.
func Get(key string) string {
	lock.Lock()
	defer lock.Unlock()
	return prop.Get(key)
}

// MustGet returns the property of the key, if not exist, returns the default value.
func MustGet(key string, def string) string {
	lock.Lock()
	defer lock.Unlock()
	return prop.Get(key, def)
}

// Set sets the property of the key.
func Set(key string, val interface{}) error {
	lock.Lock()
	defer lock.Unlock()
	return prop.Set(key, val)
}

// Delete removes the property.
func Delete(key string) {
	lock.Lock()
	defer lock.Unlock()
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
