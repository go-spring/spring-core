/*
 * Copyright 2012-2024 the original author or authors.
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
	prop = conf.NewProperties()
	lock sync.Mutex
)

// Get returns the property of the key.
func Get(key string, opts ...conf.GetOption) string {
	lock.Lock()
	defer lock.Unlock()
	return prop.Get(key, opts...)
}

// Set sets the property of the key.
func Set(key string, val interface{}) error {
	lock.Lock()
	defer lock.Unlock()
	return prop.Set(key, val)
}

// Unset removes the property.
func Unset(key string) {
	lock.Lock()
	defer lock.Unlock()
}

// Clear clears all properties.
func Clear() {
	lock.Lock()
	defer lock.Unlock()
	prop = conf.NewProperties()
}

// Clone copies all properties into another properties.
func Clone() *conf.Properties {
	lock.Lock()
	m := prop.Data()
	lock.Unlock()
	p := conf.NewProperties()
	for k, v := range m {
		_ = p.Set(k, v)
	}
	return p
}
