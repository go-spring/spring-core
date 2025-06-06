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
	"github.com/lvan100/go-assert"
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
		assert.That(t, "go-spring").Equal(p.Get("name"))
		assert.That(t, "true").Equal(p.Get("debug"))
	})

	t.Run("missing arg", func(t *testing.T) {
		os.Args = []string{"test", "-D"}

		props := conf.New()
		err := NewCommandArgs().CopyTo(props)
		assert.ThatError(t, err).Matches("cmd option -D needs arg")
	})

	t.Run("set return error", func(t *testing.T) {
		os.Args = []string{"test", "-D", "name=go-spring", "-D", "debug"}

		p := conf.Map(map[string]any{
			"debug": []string{"true"},
		})
		err := NewCommandArgs().CopyTo(p)
		assert.ThatError(t, err).Matches("property conflict at path debug")
	})

	t.Run("custom prefix", func(t *testing.T) {
		os.Args = []string{"test", "--option", "port=8080"}

		oldEnv := os.Getenv(CommandArgsPrefix)
		defer func() { _ = os.Setenv(CommandArgsPrefix, oldEnv) }()
		_ = os.Setenv(CommandArgsPrefix, "--option")

		props := conf.New()
		err := NewCommandArgs().CopyTo(props)
		assert.Nil(t, err)
		assert.That(t, "8080").Equal(props.Get("port"))
	})

	t.Run("ignore other args", func(t *testing.T) {
		os.Args = []string{"test", "-v", "-D", "env=prod", "--log-level=info"}

		props := conf.New()
		err := NewCommandArgs().CopyTo(props)
		assert.Nil(t, err)
		assert.That(t, "prod").Equal(props.Get("env"))
		assert.False(t, props.Has("--log-level"))
		assert.False(t, props.Has("-v"))
	})
}
