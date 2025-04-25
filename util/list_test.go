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

package util_test

import (
	"testing"

	"github.com/go-spring/spring-core/util"
	"github.com/lvan100/go-assert"
)

func TestListOf(t *testing.T) {
	assert.Nil(t, util.AllOfList[string](nil))
	l := util.ListOf[string]()
	assert.Nil(t, util.AllOfList[string](l))
	l = util.ListOf("a", "b", "c")
	assert.Equal(t, []string{"a", "b", "c"}, util.AllOfList[string](l))
}
