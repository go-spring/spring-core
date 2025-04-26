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

package storage

import (
	"testing"

	"github.com/lvan100/go-assert"
)

func TestStorage(t *testing.T) {

	t.Run("empty", func(t *testing.T) {
		s := NewStorage()
		assert.That(t, s.RawData()).Equal(map[string]string{})

		subKeys, err := s.SubKeys("a")
		assert.Nil(t, err)
		assert.Nil(t, subKeys)

		subKeys, err = s.SubKeys("a.b")
		assert.Nil(t, err)
		assert.Nil(t, subKeys)

		subKeys, err = s.SubKeys("a[0]")
		assert.Nil(t, err)
		assert.Nil(t, subKeys)

		assert.False(t, s.Has("a"))
		assert.False(t, s.Has("a.b"))
		assert.False(t, s.Has("a[0]"))

		err = s.Set("", "abc")
		assert.ThatError(t, err).Matches("key is empty")
	})

	t.Run("map-0", func(t *testing.T) {
		s := NewStorage()

		err := s.Set("a", "b")
		assert.Nil(t, err)
		assert.True(t, s.Has("a"))
		assert.That(t, s.RawData()).Equal(map[string]string{
			"a": "b",
		})

		err = s.Set("a.y", "x")
		assert.ThatError(t, err).Matches("property conflict at path a.y")
		err = s.Set("a[0]", "x")
		assert.ThatError(t, err).Matches("property conflict at path a\\[0]")

		assert.False(t, s.Has(""))
		assert.False(t, s.Has("a["))
		assert.False(t, s.Has("a.y"))
		assert.False(t, s.Has("a[0]"))

		subKeys, err := s.SubKeys("")
		assert.Nil(t, err)
		assert.That(t, subKeys).Equal([]string{"a"})

		_, err = s.SubKeys("a")
		assert.ThatError(t, err).Matches("property conflict at path a")
		_, err = s.SubKeys("a[")
		assert.ThatError(t, err).Matches("invalid key 'a\\['")

		err = s.Set("a", "c")
		assert.Nil(t, err)
		assert.True(t, s.Has("a"))
		assert.That(t, s.RawData()).Equal(map[string]string{
			"a": "c",
		})
	})

	t.Run("map-1", func(t *testing.T) {
		s := NewStorage()

		err := s.Set("m.x", "y")
		assert.Nil(t, err)
		assert.True(t, s.Has("m"))
		assert.True(t, s.Has("m.x"))
		assert.That(t, s.RawData()).Equal(map[string]string{
			"m.x": "y",
		})

		assert.False(t, s.Has(""))
		assert.False(t, s.Has("m.t"))
		assert.False(t, s.Has("m.x.y"))
		assert.False(t, s.Has("m[0]"))
		assert.False(t, s.Has("m.x[0]"))

		err = s.Set("m", "a")
		assert.ThatError(t, err).Matches("property conflict at path m")
		err = s.Set("m.x.z", "w")
		assert.ThatError(t, err).Matches("property conflict at path m")
		err = s.Set("m[0]", "f")
		assert.ThatError(t, err).Matches("property conflict at path m\\[0]")

		_, err = s.SubKeys("m.t")
		assert.Nil(t, err)
		subKeys, err := s.SubKeys("m")
		assert.Nil(t, err)
		assert.That(t, subKeys).Equal([]string{"x"})

		_, err = s.SubKeys("m.x")
		assert.ThatError(t, err).Matches("property conflict at path m.x")
		_, err = s.SubKeys("m[0]")
		assert.ThatError(t, err).Matches("property conflict at path m\\[0]")

		err = s.Set("m.x", "z")
		assert.Nil(t, err)
		assert.True(t, s.Has("m"))
		assert.True(t, s.Has("m.x"))
		assert.That(t, s.RawData()).Equal(map[string]string{
			"m.x": "z",
		})

		err = s.Set("m.t", "q")
		assert.Nil(t, err)
		assert.True(t, s.Has("m"))
		assert.True(t, s.Has("m.x"))
		assert.True(t, s.Has("m.t"))
		assert.That(t, s.RawData()).Equal(map[string]string{
			"m.x": "z",
			"m.t": "q",
		})

		subKeys, err = s.SubKeys("m")
		assert.Nil(t, err)
		assert.That(t, subKeys).Equal([]string{"t", "x"})
	})

	t.Run("arr-0", func(t *testing.T) {
		s := NewStorage()

		err := s.Set("[0]", "p")
		assert.Nil(t, err)
		assert.True(t, s.Has("[0]"))
		assert.That(t, s.RawData()).Equal(map[string]string{
			"[0]": "p",
		})

		err = s.Set("[0]x", "f")
		assert.ThatError(t, err).Matches("invalid key '\\[0]x'")
		err = s.Set("[0].x", "f")
		assert.ThatError(t, err).Matches("property conflict at path \\[0].x")

		err = s.Set("[0]", "w")
		assert.Nil(t, err)
		assert.That(t, s.RawData()).Equal(map[string]string{
			"[0]": "w",
		})

		subKeys, err := s.SubKeys("")
		assert.Nil(t, err)
		assert.That(t, subKeys).Equal([]string{"0"})

		err = s.Set("[1]", "p")
		assert.Nil(t, err)
		assert.True(t, s.Has("[0]"))
		assert.That(t, s.RawData()).Equal(map[string]string{
			"[0]": "w",
			"[1]": "p",
		})

		subKeys, err = s.SubKeys("")
		assert.Nil(t, err)
		assert.That(t, subKeys).Equal([]string{"0", "1"})
	})

	t.Run("arr-1", func(t *testing.T) {
		s := NewStorage()

		err := s.Set("s[0]", "p")
		assert.Nil(t, err)
		assert.True(t, s.Has("s"))
		assert.True(t, s.Has("s[0]"))
		assert.That(t, s.RawData()).Equal(map[string]string{
			"s[0]": "p",
		})

		err = s.Set("s[1]", "o")
		assert.Nil(t, err)
		assert.True(t, s.Has("s"))
		assert.True(t, s.Has("s[0]"))
		assert.True(t, s.Has("s[1]"))
		assert.That(t, s.RawData()).Equal(map[string]string{
			"s[0]": "p",
			"s[1]": "o",
		})

		subKeys, err := s.SubKeys("s")
		assert.Nil(t, err)
		assert.That(t, subKeys).Equal([]string{"0", "1"})

		err = s.Set("s", "w")
		assert.ThatError(t, err).Matches("property conflict at path s")
		err = s.Set("s.x", "f")
		assert.ThatError(t, err).Matches("property conflict at path s.x")
	})

	t.Run("map && array", func(t *testing.T) {
		s := NewStorage()

		err := s.Set("a.b[0].c", "123")
		assert.Nil(t, err)
		assert.True(t, s.Has("a"))
		assert.True(t, s.Has("a.b"))
		assert.True(t, s.Has("a.b[0]"))
		assert.True(t, s.Has("a.b[0].c"))
		assert.That(t, s.RawData()).Equal(map[string]string{
			"a.b[0].c": "123",
		})

		err = s.Set("a.b[0].d[0]", "123")
		assert.Nil(t, err)
		assert.True(t, s.Has("a"))
		assert.True(t, s.Has("a.b"))
		assert.True(t, s.Has("a.b[0]"))
		assert.True(t, s.Has("a.b[0].d"))
		assert.True(t, s.Has("a.b[0].d[0]"))
		assert.That(t, s.RawData()).Equal(map[string]string{
			"a.b[0].c":    "123",
			"a.b[0].d[0]": "123",
		})
	})
}
