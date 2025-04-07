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

package conf_test

import (
	"testing"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/util/assert"
)

func TestProperties_Load(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		p, err := conf.Load("./testdata/config/app.properties")
		assert.Nil(t, err)
		assert.Equal(t, p.Data(), map[string]string{
			"properties.list[0]":          "1",
			"properties.list[1]":          "2",
			"properties.obj.list[0].age":  "4",
			"properties.obj.list[0].name": "tom",
			"properties.obj.list[1].age":  "2",
			"properties.obj.list[1].name": "jerry",
		})
	})

	t.Run("file not exist", func(t *testing.T) {
		_, err := conf.Load("./testdata/config/app.tcl")
		assert.Error(t, err, "no such file or directory")
	})

	t.Run("unsupported ext", func(t *testing.T) {
		_, err := conf.Load("./testdata/config/app.ini")
		assert.Error(t, err, "unsupported file type .ini")
	})

	t.Run("syntax error", func(t *testing.T) {
		_, err := conf.Load("./testdata/config/err.yaml")
		assert.Error(t, err, "did not find expected node content")
	})
}

func TestProperties_Resolve(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		p := conf.Map(map[string]interface{}{
			"a.b.c": []string{"3"},
		})
		s, err := p.Resolve("${a.b.c[0]}")
		assert.Nil(t, err)
		assert.Equal(t, s, "3")
	})

	t.Run("default", func(t *testing.T) {
		p := conf.New()
		s, err := p.Resolve("${a.b.c:=123}")
		assert.Nil(t, err)
		assert.Equal(t, s, "123")
	})

	t.Run("key not exist", func(t *testing.T) {
		p := conf.New()
		_, err := p.Resolve("${a.b.c}")
		assert.Error(t, err, "property a.b.c not exist")
	})

	t.Run("syntax error - 1", func(t *testing.T) {
		p := conf.Map(map[string]interface{}{
			"a.b.c": []string{"3"},
		})
		_, err := p.Resolve("${a.b.c}")
		assert.Error(t, err, "property a.b.c isn't simple value")
	})

	t.Run("syntax error - 2", func(t *testing.T) {
		p := conf.Map(map[string]interface{}{
			"a.b.c": []string{"3"},
		})
		_, err := p.Resolve("${a.b.c")
		assert.Error(t, err, "resolve string .* error: invalid syntax")
	})

	t.Run("syntax error - 3", func(t *testing.T) {
		p := conf.Map(map[string]interface{}{
			"a.b.c": []string{"3"},
		})
		_, err := p.Resolve("${a.b.c[0]}==${a.b.c}")
		assert.Error(t, err, "property a.b.c isn't simple value")
	})
}

func TestProperties_CopyTo(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		p := conf.Map(map[string]interface{}{
			"a.b.c": []string{"3"},
		})
		assert.Equal(t, p.Keys(), []string{
			"a.b.c[0]",
		})

		assert.True(t, p.Has("a.b.c"))
		assert.True(t, p.Has("a.b.c[0]"))
		assert.Equal(t, p.Get("a.b.c[0]"), "3")
		assert.Equal(t, p.Data(), map[string]string{
			"a.b.c[0]": "3",
		})

		s := conf.Map(map[string]interface{}{
			"a.b.c": []string{"4", "5"},
		})
		assert.Equal(t, s.Keys(), []string{
			"a.b.c[0]",
			"a.b.c[1]",
		})

		assert.True(t, s.Has("a.b.c"))
		assert.True(t, s.Has("a.b.c[0]"))
		assert.True(t, s.Has("a.b.c[1]"))
		assert.Equal(t, s.Data(), map[string]string{
			"a.b.c[0]": "4",
			"a.b.c[1]": "5",
		})

		err := p.CopyTo(s)
		assert.Nil(t, err)
		assert.Equal(t, s.Data(), map[string]string{
			"a.b.c[0]": "3",
			"a.b.c[1]": "5",
		})
	})

	t.Run("error", func(t *testing.T) {
		p := conf.Map(map[string]interface{}{
			"a.b.c": []string{"3"},
		})
		assert.Equal(t, p.Data(), map[string]string{
			"a.b.c[0]": "3",
		})

		s := conf.Map(map[string]interface{}{
			"a.b.c": "3",
		})
		assert.Equal(t, s.Get("a.b.c"), "3")

		err := p.CopyTo(s)
		assert.Error(t, err, "property conflict at path a.b.c\\[0]")
	})
}
