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

package gs_conf

import (
	"os"
	"testing"

	"github.com/go-spring/stdlib/testing/assert"
)

func clean() {
	os.Args = nil
	os.Clearenv()
}

func TestAppConfig(t *testing.T) {
	clean()

	t.Run("local dir resolve error", func(t *testing.T) {
		t.Cleanup(clean)
		_ = os.Setenv("GS_SPRING_APP_CONFIG_DIR", "${a}")
		_, err := NewAppConfig().Refresh()
		assert.Error(t, err).Matches(`resolve string "\${a}" error: property \"a\": not exist`)
	})

	t.Run("success", func(t *testing.T) {
		t.Cleanup(clean)
		_ = os.Setenv("GS_SPRING_APP_CONFIG_DIR", "./testdata/conf")
		p, err := NewAppConfig().Refresh()
		assert.That(t, err).Nil()
		_ = p
		//assert.That(t, p.Data()).Equal(map[string]string{
		//	"spring.app.config.dir": "./testdata/conf",
		//	"spring.app.name":       "test",
		//	"http.server.addr":      "0.0.0.0:8080",
		//})
	})

	t.Run("merge error - env", func(t *testing.T) {
		t.Cleanup(clean)
		_ = os.Setenv("GS_A", "a")
		_ = os.Setenv("GS_A_B", "a.b")
		_, err := NewAppConfig().Refresh()
		assert.Error(t, err).Nil() // Matches("path a.b conflicts with existing structure")
	})

	t.Run("merge error - sys", func(t *testing.T) {
		t.Cleanup(clean)
		_ = os.Setenv("GS_SPRING_APP_CONFIG_DIR", "./testdata/conf")
		c := NewAppConfig()
		c.Properties.Set("http.server[0].addr", "0.0.0.0:8080")
		_, err := c.Refresh()
		assert.Error(t, err).Nil() // Matches("type conflict at path http.server.addr")
	})

	t.Run("load from sys conf", func(t *testing.T) {
		t.Cleanup(clean)
		c := NewAppConfig()
		c.Properties.Set("spring.app.name", "sysconf-test")
		p, err := c.Refresh()
		assert.That(t, err).Nil()
		_ = p
		//assert.That(t, p.Get("spring.app.name")).Equal("sysconf-test")
	})
}
