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

package gs_cond

import (
	"strconv"
	"testing"

	"github.com/go-spring/gs-assert/assert"
)

func TestEvalExpr(t *testing.T) {
	_, err := EvalExpr("$", "3")
	assert.ThatError(t, err).Matches("doesn't return bool value")

	ok, err := EvalExpr("int($)==3", "3")
	assert.That(t, err).Nil()
	assert.That(t, ok).True()

	RegisterExpressFunc("equal", func(s string, i int) bool {
		return s == strconv.Itoa(i)
	})
	ok, err = EvalExpr("equal($,9)", "9")
	assert.That(t, err).Nil()
	assert.That(t, ok).True()
}
