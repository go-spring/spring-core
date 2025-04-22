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

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/util/assert"
)

func TestContainer(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		c := New()
		c.Object(&http.Server{})
		err := c.Refresh(conf.New())
		assert.Nil(t, err)
		c.Close()
	})

	t.Run("resolve error", func(t *testing.T) {
		c := New()
		c.Object(&http.Server{}).Condition(
			gs_cond.OnFunc(func(ctx gs.CondContext) (bool, error) {
				return false, errors.New("condition error")
			}),
		)
		err := c.Refresh(conf.New())
		assert.Error(t, err, "condition error")
	})

	t.Run("inject error", func(t *testing.T) {
		c := New()
		c.Provide(func(addr string) *http.Server { return nil })
		err := c.Refresh(conf.New())
		assert.Error(t, err, "parse tag .* error: invalid syntax")
	})
}
