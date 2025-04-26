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
	"maps"
	"slices"
	"testing"

	"github.com/go-spring/spring-core/util"
	"github.com/lvan100/go-assert"
)

func BenchmarkOrderedMapKeys(b *testing.B) {
	m := map[string]string{
		"a": "1",
		"b": "2",
		"c": "3",
		"d": "4",
		"e": "5",
		"f": "6",
		"g": "7",
		"h": "8",
		"i": "9",
		"j": "10",
		"k": "11",
		"l": "12",
		"m": "13",
		"n": "14",
		"o": "15",
		"p": "16",
		"q": "17",
		"r": "18",
		"s": "19",
		"t": "20",
		"u": "21",
		"v": "22",
		"w": "23",
		"x": "24",
		"y": "25",
		"z": "26",
	}
	b.Run("std", func(b *testing.B) {
		for b.Loop() {
			slices.Sorted(maps.Keys(m))
		}
	})
	b.Run("util", func(b *testing.B) {
		for b.Loop() {
			util.OrderedMapKeys(m)
		}
	})
}

func TestOrderedMapKeys(t *testing.T) {
	assert.That(t, util.OrderedMapKeys(map[string]int{})).Equal([]string{})
	assert.That(t, util.OrderedMapKeys(map[string]int{"a": 1, "b": 2})).Equal([]string{"a", "b"})
	assert.That(t, util.OrderedMapKeys(map[int]string{})).Equal([]int{})
	assert.That(t, util.OrderedMapKeys(map[int]string{1: "a", 2: "b"})).Equal([]int{1, 2})
}
