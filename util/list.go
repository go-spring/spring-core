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

package util

import (
	"container/list"
)

// ListOf creates a list of the given items.
func ListOf[T any](a ...T) *list.List {
	l := list.New()
	for _, i := range a {
		l.PushBack(i)
	}
	return l
}

// AllOfList returns a slice of all items in the given list.
func AllOfList[T any](l *list.List) []T {
	if l == nil {
		return nil
	}
	if l.Len() == 0 {
		return nil
	}
	ret := make([]T, 0, l.Len())
	for e := l.Front(); e != nil; e = e.Next() {
		ret = append(ret, e.Value.(T))
	}
	return ret
}
