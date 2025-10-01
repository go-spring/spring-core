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

package gs_core

import (
	"errors"
	"net/http"
	"testing"

	"github.com/go-spring/spring-base/testing/assert"
	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
)

func TestContainer(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		c := New()
		c.Root(c.Object(&http.Server{}))
		err := c.Refresh(conf.New())
		assert.That(t, err).Nil()
		c.Close()
	})

	t.Run("resolve error", func(t *testing.T) {
		c := New()
		c.Root(c.Object(&http.Server{}).Condition(
			gs_cond.OnFunc(func(ctx gs.ConditionContext) (bool, error) {
				return false, errors.New("condition error")
			}),
		))
		err := c.Refresh(conf.New())
		assert.Error(t, err).Matches("condition error")
	})

	t.Run("inject error", func(t *testing.T) {
		c := New()
		c.Root(c.Provide(func(addr string) *http.Server { return nil }))
		err := c.Refresh(conf.New())
		assert.Error(t, err).Matches("property \"\" not exist")
	})

	t.Run("duplicate object registration", func(t *testing.T) {
		c := New()
		c.Root(c.Object(&http.Server{}))
		c.Root(c.Object(&http.Server{}))
		err := c.Refresh(conf.New())
		assert.Error(t, err).Matches("found duplicate beans")
	})

	t.Run("provide with dependency", func(t *testing.T) {
		c := New()

		c.Root(c.Provide(func(addr string) *http.Server {
			return &http.Server{Addr: addr}
		}, gs_arg.Tag("${server.address:=:9090}")))

		err := c.Refresh(conf.Map(map[string]any{
			"server.address": ":8080",
		}))
		assert.That(t, err).Nil()
	})

	t.Run("provide with missing dependency", func(t *testing.T) {
		c := New()

		c.Root(c.Provide(func(addr string) *http.Server {
			return &http.Server{Addr: addr}
		}, gs_arg.Tag("${server.address}")))

		err := c.Refresh(conf.New())
		assert.Error(t, err).Matches("property \"server.address\" not exist")
	})
}
