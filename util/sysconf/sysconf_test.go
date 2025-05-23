/*
 * Copyright 2025 The Go-Spring Authors.
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

package sysconf_test

import (
	"testing"

	"github.com/go-spring/spring-core/util/sysconf"
	"github.com/lvan100/go-assert"
)

func TestSysConf(t *testing.T) {
	assert.False(t, sysconf.Has("name"))

	sysconf.Set("name", "Alice")
	assert.True(t, sysconf.Has("name"))
	assert.That(t, "Alice").Equal(sysconf.Get("name"))

	sysconf.Clear()
	assert.False(t, sysconf.Has("name"))

	sysconf.Set("name", "Alice")
	sysconf.Set("name.first", "Alice")

	p := sysconf.Clone()
	assert.That(t, p.Data()).Equal(map[string]string{"name": "Alice"})
}
