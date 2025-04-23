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

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/util/assert"
)

func TestReplaceKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"MY_ENV_VAR", "MY.ENV.VAR"},
		{"_MY_ENV_", "_MY.ENV_"},
		{"__PREFIX__KEY__", "__PREFIX__KEY__"},
		{"NO_UNDERSCORES", "NO.UNDERSCORES"},
		{"_LEADING_AND_TRAILING_", "_LEADING.AND.TRAILING_"},
	}
	for _, tt := range tests {
		got := replaceKey(tt.input)
		assert.Equal(t, got, tt.want)
	}
}

func TestEnvironment(t *testing.T) {
	os.Clearenv()

	t.Run("empty", func(t *testing.T) {
		props := conf.New()
		err := NewEnvironment().CopyTo(props)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(props.Keys()))
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
		assert.Nil(t, err)
		assert.Equal(t, props.Get("db.host"), "db1")
		assert.Equal(t, props.Get("API_KEY"), "key123")
	})

	t.Run("custom prefix", func(t *testing.T) {
		_ = os.Setenv(EnvironmentPrefix, "APP_")
		_ = os.Setenv("GS_CACHE_SIZE", "100")
		_ = os.Setenv("APP_DB_HOST", "db1")
		_ = os.Setenv("API_KEY", "key123")
		defer func() {
			_ = os.Unsetenv("GS_DB_HOST")
			_ = os.Unsetenv("API_KEY")
		}()
		props := conf.New()
		err := NewEnvironment().CopyTo(props)
		assert.Nil(t, err)
		assert.Equal(t, props.Get("db.host"), "db1")
		assert.Equal(t, props.Get("API_KEY"), "key123")
	})

	t.Run("custom patterns", func(t *testing.T) {
		_ = os.Setenv(IncludeEnvPatterns, "^TEST_")
		_ = os.Setenv(ExcludeEnvPatterns, "^TEST_INTERNAL")
		_ = os.Setenv("TEST_PUBLIC", "yes")
		_ = os.Setenv("TEST_INTERNAL", "no")
		defer func() {
			_ = os.Unsetenv(IncludeEnvPatterns)
			_ = os.Unsetenv(ExcludeEnvPatterns)
			_ = os.Unsetenv("TEST_PUBLIC")
			_ = os.Unsetenv("TEST_INTERNAL")
		}()
		props := conf.New()
		err := NewEnvironment().CopyTo(props)
		assert.Nil(t, err)
		assert.Equal(t, props.Get("TEST_PUBLIC"), "yes")
		assert.False(t, props.Has("TEST_INTERNAL"))
	})

	t.Run("invalid regex - include", func(t *testing.T) {
		_ = os.Setenv(IncludeEnvPatterns, "[invalid")
		defer func() {
			_ = os.Unsetenv(IncludeEnvPatterns)
		}()
		props := conf.New()
		err := NewEnvironment().CopyTo(props)
		assert.Error(t, err, "error parsing regexp")
	})

	t.Run("invalid regex - exclude", func(t *testing.T) {
		_ = os.Setenv(ExcludeEnvPatterns, "[invalid")
		defer func() {
			_ = os.Unsetenv(ExcludeEnvPatterns)
		}()
		props := conf.New()
		err := NewEnvironment().CopyTo(props)
		assert.Error(t, err, "error parsing regexp")
	})

	t.Run("set return error", func(t *testing.T) {
		_ = os.Setenv("GS_DB_HOST", "db1")
		defer func() {
			_ = os.Unsetenv("GS_DB_HOST")
		}()
		props := conf.Map(map[string]interface{}{
			"db": []string{"db2"},
		})
		err := NewEnvironment().CopyTo(props)
		assert.Error(t, err, "property conflict at path db.host")
	})
}
