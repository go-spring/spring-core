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

	"github.com/go-spring/gs-assert/assert"
	"github.com/go-spring/spring-core/conf"
)

func TestEnvironment(t *testing.T) {
	os.Clearenv()

	t.Run("empty", func(t *testing.T) {
		props := conf.New()
		err := NewEnvironment().CopyTo(props)
		assert.That(t, err).Nil()
		assert.That(t, 0).Equal(len(props.Keys()))
	})

	t.Run("normal", func(t *testing.T) {
		_ = os.Setenv("GS_DB_HOST", "db1")
		_ = os.Setenv("API_KEY", "key123")
		defer func() {
			_ = os.Unsetenv("GS_DB_HOST")
			_ = os.Unsetenv("API_KEY")
		}()
		props := conf.New()
		err := NewEnvironment().CopyTo(props)
		assert.That(t, err).Nil()
		assert.That(t, props.Get("db.host")).Equal("db1")
		assert.That(t, props.Get("API_KEY")).Equal("key123")
	})

	t.Run("set return error", func(t *testing.T) {
		_ = os.Setenv("GS_DB_HOST", "db1")
		defer func() {
			_ = os.Unsetenv("GS_DB_HOST")
		}()
		props := conf.Map(map[string]any{
			"db": []string{"db2"},
		})
		err := NewEnvironment().CopyTo(props)
		assert.ThatError(t, err).Matches("property conflict at path db.host")
	})
}
