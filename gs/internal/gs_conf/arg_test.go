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

func TestCommandArgs(t *testing.T) {

	t.Run("no args - 1", func(t *testing.T) {
		os.Args = nil

		props := conf.New()
		err := NewCommandArgs().CopyTo(props)
		assert.Nil(t, err)
		assert.True(t, len(props.Keys()) == 0)
	})

	t.Run("no args - 2", func(t *testing.T) {
		os.Args = []string{"test"}

		props := conf.New()
		err := NewCommandArgs().CopyTo(props)
		assert.Nil(t, err)
		assert.True(t, len(props.Keys()) == 0)
	})

	t.Run("normal", func(t *testing.T) {
		os.Args = []string{"test", "-D", "name=go-spring", "-D", "debug"}

		p := conf.New()
		err := NewCommandArgs().CopyTo(p)
		assert.Nil(t, err)
		assert.Equal(t, "go-spring", p.Get("name"))
		assert.Equal(t, "true", p.Get("debug"))
	})

	t.Run("missing arg", func(t *testing.T) {
		os.Args = []string{"test", "-D"}

		props := conf.New()
		err := NewCommandArgs().CopyTo(props)
		assert.Error(t, err, "cmd option -D needs arg")
	})

	t.Run("set return error", func(t *testing.T) {
		os.Args = []string{"test", "-D", "name=go-spring", "-D", "debug"}

		p, err := conf.Map(map[string]interface{}{
			"debug": []string{"true"},
		})
		assert.Nil(t, err)
		err = NewCommandArgs().CopyTo(p)
		assert.Error(t, err, "property 'debug' is an array but 'debug' wants other type")
	})

	t.Run("custom prefix", func(t *testing.T) {
		os.Args = []string{"test", "--option", "port=8080"}

		oldEnv := os.Getenv(CommandArgsPrefix)
		defer func() { _ = os.Setenv(CommandArgsPrefix, oldEnv) }()
		_ = os.Setenv(CommandArgsPrefix, "--option")

		props := conf.New()
		err := NewCommandArgs().CopyTo(props)
		assert.Nil(t, err)
		assert.Equal(t, "8080", props.Get("port"))
	})

	t.Run("ignore other args", func(t *testing.T) {
		os.Args = []string{"test", "-v", "-D", "env=prod", "--log-level=info"}

		props := conf.New()
		err := NewCommandArgs().CopyTo(props)
		assert.Nil(t, err)
		assert.Equal(t, "prod", props.Get("env"))
		assert.False(t, props.Has("--log-level"))
		assert.False(t, props.Has("-v"))
	})
}
