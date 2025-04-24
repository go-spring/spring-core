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

	"github.com/go-spring/spring-core/util/assert"
	"github.com/go-spring/spring-core/util/iterutil"
)

func TestTimes(t *testing.T) {
	var arr []int
	for i := range iterutil.Times(5) {
		arr = append(arr, i)
	}
	assert.Equal(t, arr, []int{0, 1, 2, 3, 4})
}

func TestRanges(t *testing.T) {
	var arr []int
	for i := range iterutil.Ranges(1, 5) {
		arr = append(arr, i)
	}
	assert.Equal(t, arr, []int{1, 2, 3, 4})
	arr = nil
	for i := range iterutil.Ranges(5, 1) {
		arr = append(arr, i)
	}
	assert.Equal(t, arr, []int{5, 4, 3, 2})
	arr = nil
	for i := range iterutil.Ranges(1, 1) {
		arr = append(arr, i)
	}
	assert.Equal(t, arr, ([]int)(nil))
}

func TestStepRanges(t *testing.T) {
	var arr []int
	for i := range iterutil.StepRanges(1, 5, 2) {
		arr = append(arr, i)
	}
	assert.Equal(t, arr, []int{1, 3})
	arr = nil
	for i := range iterutil.StepRanges(5, 1, -2) {
		arr = append(arr, i)
	}
	assert.Equal(t, arr, []int{5, 3})
	arr = nil
	for i := range iterutil.StepRanges(5, 5, -2) {
		arr = append(arr, i)
	}
	assert.Equal(t, arr, ([]int)(nil))
}
