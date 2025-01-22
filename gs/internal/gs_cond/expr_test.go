/*
 * Copyright 2012-2024 the original author or authors.
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

package gs_cond_test

import (
	"strconv"
	"testing"

	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/util/assert"
)

func TestEvalExpr(t *testing.T) {
	ok, err := gs_cond.EvalExpr("$==3", 3)
	assert.Nil(t, err)
	assert.True(t, ok)

	gs_cond.RegisterExpressFunc("check", func(s string, i int) bool {
		return s == strconv.Itoa(i)
	})
	ok, err = gs_cond.EvalExpr("check($,9)", "9")
	assert.Nil(t, err)
	assert.True(t, ok)
}
