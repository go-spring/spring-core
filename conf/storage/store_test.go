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

package storage_test

import (
	"testing"

	"github.com/go-spring/spring-core/conf/storage"
	"github.com/go-spring/spring-core/util/assert"
)

func TestStorage(t *testing.T) {

	var s *storage.Storage

	{
		s = storage.NewStorage()
		assert.Equal(t, s.Data(), map[string]string{})
		assert.Equal(t, s.Keys(), []string{})

		subKeys, err := s.SubKeys("m")
		assert.Nil(t, err)
		assert.Nil(t, subKeys)

		assert.False(t, s.Has("m[b]"))

		subKeys, err = s.SubKeys("m[b]")
		assert.Error(t, err, "invalid key 'm\\[b]'")

		err = s.Set("m[b]", "123")
		assert.Error(t, err, "invalid key 'm\\[b]'")

		err = s.Set("[0].x", "123")
		assert.Nil(t, err)
	}

	// 初始值是简单的 KV 值
	{
		s = storage.NewStorage()

		err := s.Set("a", "b")
		assert.Nil(t, err)
		assert.True(t, s.Has("a"))
		// assert.Equal(t, s.Get("a"), "b")
		err = s.Set("a[0]", "x")
		assert.Error(t, err, "property 'a' is a value but 'a\\[0]' wants other type")
		err = s.Set("a.y", "x")
		assert.Error(t, err, "property 'a' is a value but 'a\\.y' wants other type")
		assert.Equal(t, s.Keys(), []string{"a"})

		_, err = s.SubKeys("a")
		assert.Error(t, err, "property 'a' is value")

		err = s.Set("a", "c")
		assert.Nil(t, err)
		assert.True(t, s.Has("a"))
		// assert.Equal(t, s.Get("a"), "c")
		err = s.Set("a", "")
		assert.Nil(t, err)
		assert.True(t, s.Has("a"))
		// assert.Equal(t, s.Get("a"), "")
		err = s.Set("a[0]", "x")
		assert.Error(t, err, "property 'a' is a value but 'a\\[0]' wants other type")
		err = s.Set("a.y", "x")
		assert.Error(t, err, "property 'a' is a value but 'a\\.y' wants other type")
		assert.Equal(t, s.Keys(), []string{"a"})

		err = s.Set("a", "c")
		assert.Nil(t, err)
		assert.True(t, s.Has("a"))
		// assert.Equal(t, s.Get("a"), "c")
		err = s.Set("a[0]", "x")
		assert.Error(t, err, "property 'a' is a value but 'a\\[0]' wants other type")
		err = s.Set("a.y", "x")
		assert.Error(t, err, "property 'a' is a value but 'a\\.y' wants other type")
		assert.Equal(t, s.Keys(), []string{"a"})
	}

	// 初始值是嵌套的 KV 值
	{
		s = storage.NewStorage()

		err := s.Set("m.x", "y")
		assert.Nil(t, err)
		assert.True(t, s.Has("m"))
		assert.True(t, s.Has("m.x"))
		// assert.Equal(t, s.Get("m.x"), "y")
		err = s.Set("m", "w")
		assert.Error(t, err, "property 'm' is a map but 'm' wants other type")
		err = s.Set("m[0]", "f")
		assert.Error(t, err, "property 'm' is a map but 'm\\[0]' wants other type")
		assert.Equal(t, s.Keys(), []string{"m.x"})

		subKeys, err := s.SubKeys("m")
		assert.Nil(t, err)
		assert.Equal(t, subKeys, []string{"x"})

		err = s.Set("m.x", "z")
		assert.Nil(t, err)
		assert.True(t, s.Has("m"))
		assert.True(t, s.Has("m.x"))
		// assert.Equal(t, s.Get("m.x"), "z")
		err = s.Set("m", "")
		assert.Error(t, err, "invalid node type 2")
		assert.True(t, s.Has("m"))
		assert.True(t, s.Has("m.x"))
		err = s.Set("m", "w")
		assert.Error(t, err, "property 'm' is a map but 'm' wants other type")
		err = s.Set("m[0]", "f")
		assert.Error(t, err, "property 'm' is a map but 'm\\[0]' wants other type")
		assert.Equal(t, s.Keys(), []string{"m.x"})

		subKeys, err = s.SubKeys("m")
		assert.Nil(t, err)
		assert.Equal(t, subKeys, []string{"x"})

		err = s.Set("m.t", "q")
		assert.Nil(t, err)
		assert.True(t, s.Has("m"))
		assert.True(t, s.Has("m.x"))
		assert.True(t, s.Has("m.t"))
		// assert.Equal(t, s.Get("m.x"), "")
		// assert.Equal(t, s.Get("m.t"), "q")
		err = s.Set("m", "w")
		assert.Error(t, err, "property 'm' is a map but 'm' wants other type")
		err = s.Set("m[0]", "f")
		assert.Error(t, err, "property 'm' is a map but 'm\\[0]' wants other type")
		err = s.Set("m.t[0]", "f")
		assert.Error(t, err, "property 'm.t' is a value but 'm.t\\[0]' wants other type")
		assert.Equal(t, s.Keys(), []string{"m.t", "m.x"})

		subKeys, err = s.SubKeys("m")
		assert.Nil(t, err)
		assert.Equal(t, subKeys, []string{"t", "x"})
	}

	// 初始值是数组 KV 值
	{
		s = storage.NewStorage()

		err := s.Set("s[0]", "p")
		assert.Nil(t, err)
		assert.True(t, s.Has("s"))
		assert.True(t, s.Has("s[0]"))
		// assert.Equal(t, s.Get("s[0]"), "p")
		err = s.Set("s", "w")
		assert.Error(t, err, "property 's' is an array but 's' wants other type")
		err = s.Set("s.x", "f")
		assert.Error(t, err, "property 's' is an array but 's\\.x' wants other type")
		assert.Equal(t, s.Keys(), []string{"s[0]"})

		subKeys, err := s.SubKeys("s")
		assert.Nil(t, err)
		assert.Equal(t, subKeys, []string{"0"})

		err = s.Set("s[0]", "q")
		assert.Nil(t, err)
		assert.True(t, s.Has("s"))
		assert.True(t, s.Has("s[0]"))
		// assert.Equal(t, s.Get("s[0]"), "q")
		err = s.Set("s", "")
		assert.Error(t, err, "invalid node type 3")
		assert.True(t, s.Has("s"))
		assert.True(t, s.Has("s[0]"))
		err = s.Set("s", "w")
		assert.Error(t, err, "property 's' is an array but 's' wants other type")
		err = s.Set("s.x", "f")
		assert.Error(t, err, "property 's' is an array but 's\\.x' wants other type")
		assert.Equal(t, s.Keys(), []string{"s[0]"})

		subKeys, err = s.SubKeys("s")
		assert.Nil(t, err)
		assert.Equal(t, subKeys, []string{"0"})

		err = s.Set("s[1]", "o")
		assert.Nil(t, err)
		assert.True(t, s.Has("s"))
		assert.True(t, s.Has("s[0]"))
		assert.True(t, s.Has("s[1]"))
		// assert.Equal(t, s.Get("s[0]"), "")
		// assert.Equal(t, s.Get("s[1]"), "o")
		err = s.Set("s", "w")
		assert.Error(t, err, "property 's' is an array but 's' wants other type")
		err = s.Set("s.x", "f")
		assert.Error(t, err, "property 's' is an array but 's\\.x' wants other type")
		assert.Equal(t, s.Keys(), []string{"s[0]", "s[1]"})

		subKeys, err = s.SubKeys("s")
		assert.Nil(t, err)
		assert.Equal(t, subKeys, []string{"0", "1"})
	}

	{
		s = storage.NewStorage()

		err := s.Set("a.b[0].c", "")
		assert.Nil(t, err)
		assert.Equal(t, s.Keys(), []string{"a.b[0].c"})

		err = s.Set("a.b[0].c[0]", "123")
		assert.Nil(t, err)
		assert.Equal(t, s.Keys(), []string{"a.b[0].c[0]"})

		err = s.Set("a.b[0].d", "")
		assert.Nil(t, err)
		assert.Equal(t, s.Keys(), []string{"a.b[0].c[0]", "a.b[0].d"})

		err = s.Set("a.b[0].d.e", "123")
		assert.Nil(t, err)
		assert.Equal(t, s.Keys(), []string{"a.b[0].c[0]", "a.b[0].d.e"})
	}
}
