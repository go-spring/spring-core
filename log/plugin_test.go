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

package log

import (
	"reflect"
	"testing"

	"github.com/lvan100/go-assert"
)

func TestRegisterPlugin(t *testing.T) {
	assert.Panic(t, func() {
		RegisterPlugin[int]("DummyLayout", PluginTypeLayout)
	}, "T must be struct")
	assert.Panic(t, func() {
		RegisterPlugin[FileAppender]("File", PluginTypeAppender)
	}, "duplicate plugin Appender in .*/plugin_appender.go:30 and .*/plugin_test.go:31")
}

func TestInjectAttribute(t *testing.T) {

	t.Run("no attribute - 1", func(t *testing.T) {
		type ErrorPlugin struct {
			Name string `PluginAttribute:""`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, nil, nil)
		assert.ThatError(t, err).Matches("create plugin log.ErrorPlugin error << found no attribute for struct field Name")
	})

	t.Run("no attribute - 2", func(t *testing.T) {
		type ErrorPlugin struct {
			Name string `PluginAttribute:"name"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, &Node{}, nil)
		assert.ThatError(t, err).Matches("create plugin log.ErrorPlugin error << found no attribute for struct field Name")
	})

	t.Run("property not found", func(t *testing.T) {
		type ErrorPlugin struct {
			Name string `PluginAttribute:"name,default=${not-exist-prop}"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, &Node{}, nil)
		assert.ThatError(t, err).Matches("create plugin log.ErrorPlugin error << property \\${not-exist-prop} not found")
	})

	t.Run("converter error", func(t *testing.T) {
		type ErrorPlugin struct {
			Level Level `PluginAttribute:"level,default=NOT-EXIST-LEVEL"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, &Node{}, nil)
		assert.ThatError(t, err).Matches("create plugin log.ErrorPlugin error << inject struct field Level error << invalid level NOT-EXIST-LEVEL")
	})

	t.Run("uint64 error", func(t *testing.T) {
		type ErrorPlugin struct {
			M uint64 `PluginAttribute:"m,default=111"`
			N uint64 `PluginAttribute:"n,default=abc"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, &Node{}, nil)
		assert.ThatError(t, err).Matches(`create plugin log.ErrorPlugin error << inject struct field N error << strconv.ParseUint: parsing \"abc\": invalid syntax`)
	})

	t.Run("int64 error", func(t *testing.T) {
		type ErrorPlugin struct {
			M int64 `PluginAttribute:"m,default=111"`
			N int64 `PluginAttribute:"n,default=abc"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, &Node{}, nil)
		assert.ThatError(t, err).Matches(`create plugin log.ErrorPlugin error << inject struct field N error << strconv.ParseInt: parsing \"abc\": invalid syntax`)
	})

	t.Run("float64 error", func(t *testing.T) {
		type ErrorPlugin struct {
			M float64 `PluginAttribute:"m,default=111"`
			N float64 `PluginAttribute:"n,default=abc"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, &Node{}, nil)
		assert.ThatError(t, err).Matches(`create plugin log.ErrorPlugin error << inject struct field N error << strconv.ParseFloat: parsing \"abc\": invalid syntax`)
	})

	t.Run("boolean error", func(t *testing.T) {
		type ErrorPlugin struct {
			M bool `PluginAttribute:"m,default=true"`
			N bool `PluginAttribute:"n,default=abc"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, &Node{}, nil)
		assert.ThatError(t, err).Matches(`create plugin log.ErrorPlugin error << inject struct field N error << strconv.ParseBool: parsing \"abc\": invalid syntax`)
	})

	t.Run("type error", func(t *testing.T) {
		type ErrorPlugin struct {
			M chan error `PluginAttribute:"m,default=true"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, &Node{}, nil)
		assert.ThatError(t, err).Matches(`create plugin log.ErrorPlugin error << unsupported inject type chan error for struct field M`)
	})
}

func TestInjectElement(t *testing.T) {

	t.Run("no element - 1", func(t *testing.T) {
		type ErrorPlugin struct {
			Layout Layout `PluginElement:""`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, nil, nil)
		assert.ThatError(t, err).Matches("create plugin log.ErrorPlugin error << found no element for struct field Layout")
	})

	t.Run("plugin not found - 1", func(t *testing.T) {
		type ErrorPlugin struct {
			Layout Layout `PluginElement:"Layout"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, &Node{
			Children: []*Node{
				{Label: "NotExistElement"},
			},
		}, nil)
		assert.ThatError(t, err).Matches("create plugin log.ErrorPlugin error << plugin NotExistElement not found for struct field Layout")
	})

	t.Run("plugin type mismatch", func(t *testing.T) {
		type ErrorPlugin struct {
			Layout Layout `PluginElement:"Layout"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, &Node{
			Children: []*Node{
				{Label: "File"},
			},
		}, nil)
		assert.ThatError(t, err).Matches("create plugin log.ErrorPlugin error << found no plugin elements for struct field Layout")
	})

	t.Run("plugin not found - 2", func(t *testing.T) {
		type ErrorPlugin struct {
			Layout Layout `PluginElement:"Layout,default"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, &Node{
			Children: []*Node{
				{Label: "File"},
			},
		}, nil)
		assert.ThatError(t, err).Matches("create plugin log.ErrorPlugin error << found no plugin elements for struct field Layout")
	})

	t.Run("plugin not found - 3", func(t *testing.T) {
		type ErrorPlugin struct {
			Layout Layout `PluginElement:"Layout,default=DummyLayout"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, &Node{
			Children: []*Node{
				{Label: "File"},
			},
		}, nil)
		assert.ThatError(t, err).Matches("create plugin log.ErrorPlugin error << plugin DummyLayout not found for struct field Layout")
	})

	t.Run("NewPlugin error - 1", func(t *testing.T) {
		type ErrorPlugin struct {
			Layout Layout `PluginElement:"Layout"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, &Node{
			Children: []*Node{
				{
					Label: "TextLayout",
					Attributes: map[string]string{
						"bufferSize": "1GB",
					},
				},
			},
		}, nil)
		assert.ThatError(t, err).Matches(`create plugin log.ErrorPlugin error << create plugin log.TextLayout error ` +
			`<< inject struct field BufferSize error << unhandled size name: \"GB\"`)
	})

	t.Run("NewPlugin error - 2", func(t *testing.T) {
		type ErrorPlugin struct {
			Appender Appender `PluginElement:"Appender,default=File"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, &Node{}, nil)
		assert.ThatError(t, err).Matches(`create plugin log.ErrorPlugin error << create plugin log.FileAppender error ` +
			`<< found no attribute for struct field Name`)
	})

	t.Run("NewPlugin error - 3", func(t *testing.T) {
		type ErrorPlugin struct {
			Layout Layout `PluginElement:"Layout"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, &Node{
			Children: []*Node{
				{Label: "TextLayout"},
				{Label: "TextLayout"},
			},
		}, nil)
		assert.ThatError(t, err).Matches("create plugin log.ErrorPlugin error << found 2 plugin elements for struct field Layout")
	})

	t.Run("NewPlugin error - 4", func(t *testing.T) {
		type ErrorPlugin struct {
			Layout map[string]Layout `PluginElement:"Layout"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, &Node{
			Children: []*Node{
				{Label: "TextLayout"},
				{Label: "TextLayout"},
			},
		}, nil)
		assert.ThatError(t, err).Matches("create plugin log.ErrorPlugin error << unsupported inject type map\\[string]log.Layout for struct field Layout")
	})

	t.Run("NewPlugin success - 1", func(t *testing.T) {
		type ErrorPlugin struct {
			Layout Layout `PluginElement:"Layout"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, &Node{
			Children: []*Node{
				{Label: "TextLayout"},
			},
		}, nil)
		assert.Nil(t, err)
	})

	t.Run("NewPlugin success - 2", func(t *testing.T) {
		type ErrorPlugin struct {
			Layout Layout `PluginElement:"Layout,default=TextLayout"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, &Node{}, nil)
		assert.Nil(t, err)
	})

	t.Run("NewPlugin success - 3", func(t *testing.T) {
		type ErrorPlugin struct {
			Layouts []Layout `PluginElement:"Layout,default=TextLayout"`
		}
		typ := reflect.TypeFor[ErrorPlugin]()
		_, err := NewPlugin(typ, &Node{}, nil)
		assert.Nil(t, err)
	})
}
