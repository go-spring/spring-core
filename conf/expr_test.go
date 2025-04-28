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

package conf_test

import (
	"testing"

	"github.com/go-spring/spring-core/conf"
	"github.com/lvan100/go-assert"
)

func TestExpr(t *testing.T) {
	conf.RegisterValidateFunc("checkInt", func(i int) bool {
		return i < 5
	})

	t.Run("success", func(t *testing.T) {
		var v struct {
			A int `value:"${a}" expr:"checkInt($)"`
		}
		p := conf.Map(map[string]any{
			"a": 4,
		})
		err := p.Bind(&v)
		assert.Nil(t, err)
		assert.That(t, 4).Equal(v.A)
	})

	t.Run("return false", func(t *testing.T) {
		var v struct {
			A int `value:"${a}" expr:"checkInt($)"`
		}
		p := conf.Map(map[string]any{
			"a": 14,
		})
		err := p.Bind(&v)
		assert.ThatError(t, err).Matches("validate failed on .* for value 14")
	})

	t.Run("return not bool", func(t *testing.T) {
		var v struct {
			A int `value:"${a}" expr:"$+$"`
		}
		p := conf.Map(map[string]any{
			"a": 4,
		})
		err := p.Bind(&v)
		assert.ThatError(t, err).Matches("eval .* doesn't return bool value")
	})
}
