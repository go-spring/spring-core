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

package gs_util

import (
	"container/list"
	"testing"

	"github.com/go-spring/spring-core/util"
	"github.com/go-spring/spring-core/util/assert"
)

func TestTripleSort(t *testing.T) {

	t.Run("empty list", func(t *testing.T) {
		sorting := list.New()
		sorted, err := TripleSort(sorting, nil)
		assert.Nil(t, err)
		assert.Equal(t, sorted.Len(), 0)
	})

	t.Run("single element", func(t *testing.T) {
		getBefore := func(_ *list.List, _ interface{}) *list.List {
			return list.New()
		}
		sorting := util.ListOf("A")
		sorted, err := TripleSort(sorting, getBefore)
		assert.Nil(t, err)
		assert.Equal(t, sorted.Len(), 1)
		assert.Equal(t, sorted.Front().Value, "A")
	})

	t.Run("independent elements", func(t *testing.T) {
		// A、B、C
		getBefore := func(_ *list.List, _ interface{}) *list.List {
			return list.New()
		}
		sorting := util.ListOf("A", "B", "C")
		sorted, err := TripleSort(sorting, getBefore)
		assert.Nil(t, err)
		assert.Equal(t, util.AllOfList[string](sorted), []string{"A", "B", "C"})
	})

	t.Run("linear dependency", func(t *testing.T) {
		// A -> B -> C
		getBefore := func(_ *list.List, current interface{}) *list.List {
			l := list.New()
			switch current {
			case "A":
				l.PushBack("B")
			case "B":
				l.PushBack("C")
			}
			return l
		}
		sorting := util.ListOf("A", "B", "C")
		sorted, err := TripleSort(sorting, getBefore)
		assert.Nil(t, err)
		assert.Equal(t, util.AllOfList[string](sorted), []string{"C", "B", "A"})
	})

	t.Run("multiple dependencies", func(t *testing.T) {
		// A -> B&C, B -> C
		getBefore := func(_ *list.List, current interface{}) *list.List {
			l := list.New()
			switch current {
			case "A":
				l.PushBack("B")
				l.PushBack("C")
			case "B":
				l.PushBack("C")
			}
			return l
		}
		sorting := util.ListOf("A", "B", "C")
		sorted, err := TripleSort(sorting, getBefore)
		assert.Nil(t, err)
		assert.Equal(t, util.AllOfList[string](sorted), []string{"C", "B", "A"})
	})

	t.Run("cycle", func(t *testing.T) {
		// A -> B -> C -> A
		getBefore := func(_ *list.List, current interface{}) *list.List {
			l := list.New()
			switch current {
			case "A":
				l.PushBack("B")
			case "B":
				l.PushBack("C")
			case "C":
				l.PushBack("A") // cycle
			}
			return l
		}
		sorting := util.ListOf("A", "B", "C")
		_, err := TripleSort(sorting, getBefore)
		assert.Error(t, err, "found sorting cycle")
	})
}
