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

package iterutil_test

import (
	"testing"

	"github.com/go-spring/spring-core/util/iterutil"
	"github.com/lvan100/go-assert"
)

func TestTimes(t *testing.T) {
	var arr []int
	iterutil.Times(5, func(i int) {
		arr = append(arr, i)
	})
	assert.Equal(t, arr, []int{0, 1, 2, 3, 4})
}

func TestRanges(t *testing.T) {
	var arr []int
	iterutil.Ranges(1, 5, func(i int) {
		arr = append(arr, i)
	})
	assert.Equal(t, arr, []int{1, 2, 3, 4})
	arr = nil
	iterutil.Ranges(5, 1, func(i int) {
		arr = append(arr, i)
	})
	assert.Equal(t, arr, []int{5, 4, 3, 2})
}

func TestStepRanges(t *testing.T) {
	var arr []int
	iterutil.StepRanges(1, 5, 2, func(i int) {
		arr = append(arr, i)
	})
	assert.Equal(t, arr, []int{1, 3})
	arr = nil
	iterutil.StepRanges(5, 1, -2, func(i int) {
		arr = append(arr, i)
	})
	assert.Equal(t, arr, []int{5, 3})
}
