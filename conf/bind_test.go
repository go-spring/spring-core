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
	"errors"
	"image"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/go-spring/spring-base/testing/assert"
	"github.com/go-spring/spring-core/conf"
	"github.com/spf13/cast"
)

func init() {
	conf.RegisterConverter(PointConverter)
	conf.RegisterSplitter("PointSplitter", PointSplitter)
}

type funcFilter func(i any, param conf.BindParam) (bool, error)

func (f funcFilter) Do(i any, param conf.BindParam) (bool, error) {
	return f(i, param)
}

func PointConverter(val string) (image.Point, error) {
	ss := strings.Split(val[1:len(val)-1], ",")
	x := cast.ToInt(ss[0])
	y := cast.ToInt(ss[1])
	return image.Point{X: x, Y: y}, nil
}

func PointSplitter(str string) ([]string, error) {
	if !strings.HasPrefix(str, "(") || !strings.HasSuffix(str, ")") {
		return nil, errors.New("split error")
	}
	var ret []string
	var lastIndex int
	for i, c := range str {
		if c == ')' {
			ret = append(ret, str[lastIndex:i+1])
			lastIndex = i + 1
		}
	}
	return ret, nil
}

func TestConverter(t *testing.T) {
	var s struct {
		Time     time.Time     `value:"${time:=2025-02-01}"`
		Duration time.Duration `value:"${duration:=10s}"`
	}

	t.Run("success", func(t *testing.T) {
		err := conf.New().Bind(&s)
		assert.That(t, err).Nil()
		assert.That(t, s.Time).Equal(time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC))
		assert.That(t, s.Duration).Equal(10 * time.Second)
	})

	t.Run("error", func(t *testing.T) {
		p := conf.Map(map[string]any{
			"time": "2025-02-01M00:00:00",
		})
		err := p.Bind(&s)
		assert.ThatError(t, err).Matches("unable to parse date: 2025-02-01M00:00:00")
	})
}

func TestSplitter(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		var points []image.Point
		err := conf.New().Bind(&points, "${:=(1,2)(3,4)}>>PointSplitter")
		assert.That(t, err).Nil()
		assert.That(t, points).Equal([]image.Point{{X: 1, Y: 2}, {X: 3, Y: 4}})
	})

	t.Run("split error", func(t *testing.T) {
		var points []image.Point
		err := conf.New().Bind(&points, "${:=(1}>>PointSplitter")
		assert.ThatError(t, err).Matches("split error")
	})

	t.Run("unknown splitter", func(t *testing.T) {
		var points []image.Point
		err := conf.New().Bind(&points, "${:=(1}>>UnknownSplitter")
		assert.ThatError(t, err).Matches("unknown splitter 'UnknownSplitter'")
	})
}

func TestSplitterError(t *testing.T) {
	conf.RegisterSplitter("ErrorSplitter", func(str string) ([]string, error) {
		return nil, errors.New("splitter error")
	})

	t.Run("splitter returns error", func(t *testing.T) {
		var strs []string
		err := conf.New().Bind(&strs, "${strs:=a,b,c}>>ErrorSplitter")
		assert.ThatError(t, err).Matches("splitter error")
	})
}

func TestParseTag(t *testing.T) {

	t.Run("normal", func(t *testing.T) {
		tag, err := conf.ParseTag("${a}")
		assert.That(t, err).Nil()
		assert.That(t, tag.String()).Equal("${a}")
	})

	t.Run("default", func(t *testing.T) {
		tag, err := conf.ParseTag("${a:=123}")
		assert.That(t, err).Nil()
		assert.That(t, tag.String()).Equal("${a:=123}")
	})

	t.Run("splitter", func(t *testing.T) {
		tag, err := conf.ParseTag("${a:=1,2,3}>>splitter")
		assert.That(t, err).Nil()
		assert.That(t, tag.String()).Equal("${a:=1,2,3}>>splitter")
	})

	t.Run("error - 1", func(t *testing.T) {
		_, err := conf.ParseTag(">>splitter")
		assert.ThatError(t, err).Matches("parse tag .* error: invalid syntax")
	})

	t.Run("error - 2", func(t *testing.T) {
		_, err := conf.ParseTag("${a:=1,2,3")
		assert.ThatError(t, err).Matches("parse tag .* error: invalid syntax")
	})

	t.Run("error - 3", func(t *testing.T) {
		_, err := conf.ParseTag("{a:=1,2,3}")
		assert.ThatError(t, err).Matches("parse tag .* error: invalid syntax")
	})

	t.Run("empty key with default", func(t *testing.T) {
		tag, err := conf.ParseTag("${:=default}")
		assert.That(t, err).Nil()
		assert.That(t, tag).Equal(conf.ParsedTag{
			Key:    "",
			Def:    "default",
			HasDef: true,
		})
	})

	t.Run("key with special chars", func(t *testing.T) {
		tag, err := conf.ParseTag("${key-with.dots_and_underscores:=value}")
		assert.That(t, err).Nil()
		assert.That(t, tag).Equal(conf.ParsedTag{
			Key:    "key-with.dots_and_underscores",
			Def:    "value",
			HasDef: true,
		})
	})
}

func TestBindParam(t *testing.T) {

	t.Run("root", func(t *testing.T) {
		var param conf.BindParam
		err := param.BindTag("${ROOT}", "")
		assert.That(t, err).Nil()
		assert.That(t, param).Equal(conf.BindParam{})
	})

	t.Run("normal", func(t *testing.T) {
		var param conf.BindParam
		err := param.BindTag("${a:=1,2,3}>>splitter", "")
		assert.That(t, err).Nil()
		assert.That(t, param).Equal(conf.BindParam{
			Key:  "a",
			Path: "",
			Tag: conf.ParsedTag{
				Key:      "a",
				Def:      "1,2,3",
				HasDef:   true,
				Splitter: "splitter",
			},
			Validate: "",
		})
	})

	t.Run("sub path", func(t *testing.T) {
		var param = conf.BindParam{
			Key:  "s",
			Path: "Struct",
		}
		err := param.BindTag("${a:=1,2,3}>>splitter", "")
		assert.That(t, err).Nil()
		assert.That(t, param).Equal(conf.BindParam{
			Key:  "s.a",
			Path: "Struct",
			Tag: conf.ParsedTag{
				Key:      "a",
				Def:      "1,2,3",
				HasDef:   true,
				Splitter: "splitter",
			},
			Validate: "",
		})
	})

	t.Run("default", func(t *testing.T) {
		var param conf.BindParam
		err := param.BindTag("${:=1,2,3}>>splitter", "")
		assert.That(t, err).Nil()
		assert.That(t, param).Equal(conf.BindParam{
			Key:  "",
			Path: "",
			Tag: conf.ParsedTag{
				Key:      "",
				Def:      "1,2,3",
				HasDef:   true,
				Splitter: "splitter",
			},
			Validate: "",
		})
	})

	t.Run("error - 1", func(t *testing.T) {
		var param conf.BindParam
		err := param.BindTag("a:=123", "")
		assert.ThatError(t, err).Matches("parse tag .* error: invalid syntax")
	})

	t.Run("error - 2", func(t *testing.T) {
		var param conf.BindParam
		err := param.BindTag("${}", "")
		assert.ThatError(t, err).Matches("parse tag .* error: invalid syntax")
	})

	t.Run("empty tag with no default", func(t *testing.T) {
		var param conf.BindParam
		err := param.BindTag("${:=}", "")
		assert.ThatError(t, err).Nil()
	})

	t.Run("nested key", func(t *testing.T) {
		var param = conf.BindParam{
			Key:  "parent",
			Path: "Parent",
		}
		err := param.BindTag("${child.key:=value}", "")
		assert.That(t, err).Nil()
		assert.That(t, param.Key).Equal("parent.child.key")
	})
}

type DBConnection struct {
	UserName string `value:"${username}"`
	Password string `value:"${password}"`
	Url      string `value:"${url}"`
	Port     string `value:"${port}"`
}

type TaggedNestedDB struct {
	DBConnection `value:"${conn}"`
	DB           string `value:"${db}"`
}

type UntaggedNestedDB struct {
	DBConnection
	DB string `value:"${db}"`
}

type Extra struct {
	Bool     bool           `value:"${bool:=true}" expr:"$"`
	Int      int            `value:"${int:=4}" expr:"$==4"`
	Int8     int8           `value:"${int8:=8}" expr:"$==8"`
	Int16    int16          `value:"${int16:=16}" expr:"$==16"`
	Int32    int32          `value:"${int32:=32}" expr:"$==32"`
	Int64    int64          `value:"${int32:=64}" expr:"$==64"`
	Uint     uint           `value:"${uint:=4}" expr:"$==4"`
	Uint8    uint8          `value:"${uint8:=8}" expr:"$==8"`
	Uint16   uint16         `value:"${uint16:=16}" expr:"$==16"`
	Uint32   uint32         `value:"${uint32:=32}" expr:"$==32"`
	Uint64   uint64         `value:"${uint32:=64}" expr:"$==64"`
	Float32  float32        `value:"${float32:=3.2}" expr:"abs($-3.2)<0.000001"`
	Float64  float64        `value:"${float64:=6.4}" expr:"abs($-6.4)<0.000001"`
	String   string         `value:"${str:=xyz}" expr:"$==\"xyz\""`
	Duration time.Duration  `value:"${duration:=10s}"`
	IntsV0   []int          `value:"${intsV0:=}"`
	IntsV1   []int          `value:"${intsV1:=1,2,3}"`
	IntsV2   []int          `value:"${intsV2}"`
	MapV0    map[string]int `value:"${mapV0:=}"`
	MapV2    map[string]int `value:"${mapV2}"`
}

type DBConfig struct {
	DB0   []TaggedNestedDB   `value:"${tagged.db}"`
	DB1   []UntaggedNestedDB `value:"${db}"`
	Extra Extra              `value:"${extra}"`
}

type UnnamedDefault struct {
	Strs []string       `value:"${:=1,2,3}"`
	Ints []int          `value:"${:=}"`
	Map  map[string]int `value:"${:=}"`
}

type AdvancedTypes struct {
	BoolSlice   []bool          `value:"${boolSlice:=true,false,true}"`
	IntSlice    []int           `value:"${intSlice:=1,2,3}"`
	StringSlice []string        `value:"${stringSlice:=a,b,c}"`
	NestedMap   map[string]Data `value:"${nestedMap}"`
	EmptyStruct Data            `value:"${emptyStruct}"`
}

type Data struct {
	Name string `value:"${name}"`
	Age  int    `value:"${age}"`
}

func TestProperties_Bind(t *testing.T) {

	t.Run("unnamed default", func(t *testing.T) {
		var s UnnamedDefault
		err := conf.New().Bind(&s)
		assert.That(t, err).Nil()
		assert.That(t, s).Equal(UnnamedDefault{
			Strs: []string{"1", "2", "3"},
			Ints: []int{},
			Map:  map[string]int{},
		})
	})

	t.Run("BindTag error", func(t *testing.T) {
		var i int
		err := conf.New().Bind(&i, "$")
		assert.ThatError(t, err).Matches("parse tag '\\$' error: invalid syntax")
	})

	t.Run("target error - 1", func(t *testing.T) {
		err := conf.New().Bind(5)
		assert.ThatError(t, err).Matches("should be a ptr")
	})

	t.Run("target error - 2", func(t *testing.T) {
		err := conf.New().Bind(new(*int))
		assert.ThatError(t, err).Matches("target should be value type")
	})

	t.Run("validate error", func(t *testing.T) {
		var s struct {
			Value int `value:"${v}" expr:"$>9"`
		}
		err := conf.Map(map[string]any{
			"v": "1",
		}).Bind(&s)
		assert.ThatError(t, err).Matches("validate failed on .* for value 1")
	})

	t.Run("array error", func(t *testing.T) {
		err := conf.New().Bind(new(struct {
			Arr [3]string `value:"${arr:=1,2,3}"`
		}))
		assert.ThatError(t, err).Matches("use slice instead of array")
	})

	t.Run("type error - 1", func(t *testing.T) {
		var s struct {
			Value int `value:"${v}"`
		}
		err := conf.Map(map[string]any{
			"v": "abc",
		}).Bind(&s)
		assert.ThatError(t, err).Matches("strconv.ParseInt: parsing .*: invalid syntax")
	})

	t.Run("type error - 2", func(t *testing.T) {
		var s struct {
			Value uint `value:"${v}"`
		}
		err := conf.Map(map[string]any{
			"v": "abc",
		}).Bind(&s)
		assert.ThatError(t, err).Matches("strconv.ParseUint: parsing .*: invalid syntax")
	})

	t.Run("type error - 3", func(t *testing.T) {
		var s struct {
			Value float32 `value:"${v}"`
		}
		err := conf.Map(map[string]any{
			"v": "abc",
		}).Bind(&s)
		assert.ThatError(t, err).Matches("strconv.ParseFloat: parsing .*: invalid syntax")
	})

	t.Run("type error - 4", func(t *testing.T) {
		var s struct {
			Value bool `value:"${v}"`
		}
		err := conf.Map(map[string]any{
			"v": "abc",
		}).Bind(&s)
		assert.ThatError(t, err).Matches("strconv.ParseBool: parsing .*: invalid syntax")
	})

	t.Run("slice error - 1", func(t *testing.T) {
		var s struct {
			Value []int `value:"${v}"`
		}
		err := conf.Map(map[string]any{
			"v": []any{
				"1", "2", "a",
			},
		}).Bind(&s)
		assert.ThatError(t, err).Matches("strconv.ParseInt: parsing .*: invalid syntax")
	})

	t.Run("slice error - 2", func(t *testing.T) {
		var s struct {
			Value []int `value:"${v}"`
		}
		err := conf.New().Bind(&s)
		assert.ThatError(t, err).Matches("property \"v\" not exist")
	})

	t.Run("slice error - 3", func(t *testing.T) {
		var s struct {
			Value []image.Rectangle `value:"${v:={(1,2)(3,4)}"`
		}
		err := conf.New().Bind(&s)
		assert.ThatError(t, err).Matches("can't find converter for image.Rectangle")
	})

	t.Run("map error - 1", func(t *testing.T) {
		var s struct {
			Value map[string]int `value:"${v:=a:b,1:2}"`
		}
		err := conf.New().Bind(&s)
		assert.ThatError(t, err).Matches("map can't have a non-empty default value")
	})

	t.Run("map error - 2", func(t *testing.T) {
		var s struct {
			Value map[string]int `value:"${v}"`
		}
		err := conf.Map(map[string]any{
			"v": []any{
				"1", "2", "3",
			},
		}).Bind(&s)
		assert.ThatError(t, err).Matches("property \"v.0\" not exist")
	})

	t.Run("map error - 3", func(t *testing.T) {
		var s struct {
			Value map[string]int `value:"${v}"`
		}
		err := conf.Map(map[string]any{
			"v": "a:b,1:2",
		}).Bind(&s)
		assert.ThatError(t, err).Matches("property conflict at path v")
	})

	t.Run("map error - 4", func(t *testing.T) {
		var s struct {
			Value map[string]int `value:"${v}"`
		}
		err := conf.New().Bind(&s)
		assert.ThatError(t, err).Matches("property \"v\" not exist")
	})

	t.Run("struct error - 1", func(t *testing.T) {
		var s struct {
			Value struct {
				Int int
			} `value:"${v:={123}}"`
		}
		err := conf.New().Bind(&s)
		assert.ThatError(t, err).Matches("struct can't have a non-empty default value")
	})

	t.Run("struct error - 2", func(t *testing.T) {
		var s struct {
			int `value:"${v}"`
		}
		err := conf.Map(map[string]any{
			"v": "123",
		}).Bind(&s)
		assert.That(t, err).Nil()
		assert.That(t, s.int).Equal(0)
	})

	t.Run("struct error - 3", func(t *testing.T) {
		var s struct {
			Value int `value:"v"`
		}
		err := conf.New().Bind(&s)
		assert.ThatError(t, err).Matches("parse tag 'v' error: invalid syntax")
	})

	t.Run("struct error - 4", func(t *testing.T) {
		var s struct {
			io.Reader
		}
		err := conf.New().Bind(&s)
		assert.That(t, err).Nil()
	})

	t.Run("reflect.Value", func(t *testing.T) {
		var i int
		v := reflect.ValueOf(&i).Elem()
		err := conf.New().Bind(v, "${:=3}")
		assert.That(t, err).Nil()
		assert.That(t, i).Equal(3)
	})

	t.Run("success", func(t *testing.T) {
		expect := DBConfig{
			DB0: []TaggedNestedDB{
				{
					DBConnection: DBConnection{
						UserName: "root",
						Password: "123456",
						Url:      "1.1.1.1",
						Port:     "3306",
					},
					DB: "db1",
				},
				{
					DBConnection: DBConnection{
						UserName: "root",
						Password: "123456",
						Url:      "1.1.1.1",
						Port:     "3306",
					},
					DB: "db2",
				},
			},
			DB1: []UntaggedNestedDB{
				{
					DBConnection: DBConnection{
						UserName: "root",
						Password: "123456",
						Url:      "1.1.1.1",
						Port:     "3306",
					},
					DB: "db1",
				},
				{
					DBConnection: DBConnection{
						UserName: "root",
						Password: "123456",
						Url:      "1.1.1.1",
						Port:     "3306",
					},
					DB: "db2",
				},
			},
			Extra: Extra{
				Bool:     true,
				Int:      int(4),
				Int8:     int8(8),
				Int16:    int16(16),
				Int32:    int32(32),
				Int64:    int64(64),
				Uint:     uint(4),
				Uint8:    uint8(8),
				Uint16:   uint16(16),
				Uint32:   uint32(32),
				Uint64:   uint64(64),
				Float32:  float32(3.2),
				Float64:  6.4,
				String:   "xyz",
				Duration: time.Second * 10,
				IntsV0:   []int{},
				IntsV1:   []int{1, 2, 3},
				IntsV2:   []int{1, 2, 3},
				MapV0:    map[string]int{},
				MapV2:    map[string]int{"a": 1, "b": 2},
			},
		}

		p, err := conf.Load("./testdata/config/app.yaml")
		assert.That(t, err).Nil()

		fileID := p.AddFile("bind_test.go")

		err = p.Set("extra.intsV0", "", fileID)
		assert.That(t, err).Nil()
		err = p.Set("extra.intsV2", "1,2,3", fileID)
		assert.That(t, err).Nil()
		err = p.Set("prefix.extra.intsV2", "1,2,3", fileID)
		assert.That(t, err).Nil()

		err = p.Set("extra.mapV2.a", "1", fileID)
		assert.That(t, err).Nil()
		err = p.Set("extra.mapV2.b", "2", fileID)
		assert.That(t, err).Nil()
		err = p.Set("prefix.extra.mapV2.a", "1", fileID)
		assert.That(t, err).Nil()
		err = p.Set("prefix.extra.mapV2.b", "2", fileID)
		assert.That(t, err).Nil()

		var c DBConfig
		err = p.Bind(&c)
		assert.That(t, err).Nil()
		assert.That(t, c).Equal(expect)

		err = p.Bind(&c, "${prefix}")
		assert.That(t, err).Nil()
		assert.That(t, c).Equal(expect)
	})

	t.Run("advanced types", func(t *testing.T) {
		p := conf.Map(map[string]any{
			"boolSlice":   "true,false,true",
			"intSlice":    "1,2,3",
			"stringSlice": "a,b,c",
			"nestedMap": map[string]any{
				"user1": map[string]any{
					"name": "Alice",
					"age":  25,
				},
				"user2": map[string]any{
					"name": "Bob",
					"age":  30,
				},
			},
			"emptyStruct": map[string]any{
				"name": "Empty",
				"age":  0,
			},
			"pointerField": map[string]any{
				"name": "Pointer",
				"age":  40,
			},
		})

		var s AdvancedTypes
		err := p.Bind(&s)
		assert.That(t, err).Nil()

		assert.That(t, s.BoolSlice).Equal([]bool{true, false, true})
		assert.That(t, s.IntSlice).Equal([]int{1, 2, 3})
		assert.That(t, s.StringSlice).Equal([]string{"a", "b", "c"})

		assert.That(t, s.NestedMap).Equal(map[string]Data{
			"user1": {Name: "Alice", Age: 25},
			"user2": {Name: "Bob", Age: 30},
		})

		assert.That(t, s.EmptyStruct).Equal(Data{Name: "Empty", Age: 0})
	})

	t.Run("empty collections with defaults", func(t *testing.T) {
		var s struct {
			EmptySlice []string       `value:"${emptySlice:=}"`
			EmptyMap   map[string]int `value:"${emptyMap:=}"`
		}
		err := conf.New().Bind(&s)
		assert.That(t, err).Nil()
		assert.That(t, s.EmptySlice).Equal([]string{})
		assert.That(t, s.EmptyMap).Equal(map[string]int{})
	})

	t.Run("filter false", func(t *testing.T) {
		var param conf.BindParam
		err := param.BindTag("${ROOT}", "")
		assert.That(t, err).Nil()

		var s struct {
			Value int `value:"${v:=3}"`
		}

		v := reflect.ValueOf(&s).Elem()
		err = conf.BindValue(conf.New(), v, v.Type(), param,
			funcFilter(func(i any, param conf.BindParam) (bool, error) {
				return false, nil
			}))
		assert.That(t, err).Nil()
		assert.That(t, s.Value).Equal(3)
	})

	t.Run("filter true", func(t *testing.T) {
		var param conf.BindParam
		err := param.BindTag("${ROOT}", "")
		assert.That(t, err).Nil()

		var s struct {
			Value int `value:"${v:=3}"`
		}

		v := reflect.ValueOf(&s).Elem()
		err = conf.BindValue(conf.New(), v, v.Type(), param,
			funcFilter(func(i any, param conf.BindParam) (bool, error) {
				return true, nil
			}))
		assert.That(t, err).Nil()
		assert.That(t, s.Value).Equal(0)
	})

	t.Run("filter error", func(t *testing.T) {
		var param conf.BindParam
		err := param.BindTag("${ROOT}", "")
		assert.That(t, err).Nil()

		var s struct {
			Value int `value:"${v:=3}"`
		}

		v := reflect.ValueOf(&s).Elem()
		err = conf.BindValue(conf.New(), v, v.Type(), param,
			funcFilter(func(i any, param conf.BindParam) (bool, error) {
				return false, errors.New("filter error")
			}))
		assert.ThatError(t, err).Matches("filter error")
		assert.That(t, s.Value).Equal(0)
	})

	t.Run("property reference resolution", func(t *testing.T) {
		p := conf.Map(map[string]any{
			"host": "localhost",
			"port": "8080",
			"url":  "http://${host}:${port}",
		})

		var s struct {
			URL string `value:"${url}"`
		}

		err := p.Bind(&s)
		assert.That(t, err).Nil()
		assert.That(t, s.URL).Equal("http://localhost:8080")
	})

	t.Run("nested property reference", func(t *testing.T) {
		p := conf.Map(map[string]any{
			"protocol": "https",
			"host":     "example.com",
			"port":     "443",
			"path":     "api",
			"url":      "${protocol}://${host}:${port}/${path}",
		})

		var s struct {
			URL string `value:"${url}"`
		}

		err := p.Bind(&s)
		assert.That(t, err).Nil()
		assert.That(t, s.URL).Equal("https://example.com:443/api")
	})
}

func TestResolveString(t *testing.T) {
	t.Run("unbalanced braces", func(t *testing.T) {
		_, err := conf.ParseTag("${key")
		assert.ThatError(t, err).Matches("parse tag .* error: invalid syntax")
	})

	t.Run("missing property", func(t *testing.T) {
		p := conf.New()

		var s struct {
			Value string `value:"${missing}"`
		}

		err := p.Bind(&s)
		assert.ThatError(t, err).Matches("property \"missing\" not exist")
	})

	t.Run("missing property with default", func(t *testing.T) {
		p := conf.New()

		var s struct {
			Value string `value:"${missing:=default}"`
		}

		err := p.Bind(&s)
		assert.That(t, err).Nil()
		assert.That(t, s.Value).Equal("default")
	})
}

func TestMapBinding(t *testing.T) {
	t.Run("map with string keys and int values", func(t *testing.T) {
		p := conf.Map(map[string]any{
			"config": map[string]any{
				"a": 1,
				"b": 2,
				"c": 3,
			},
		})

		var s struct {
			Config map[string]int `value:"${config}"`
		}

		err := p.Bind(&s)
		assert.That(t, err).Nil()
		assert.That(t, s.Config).Equal(map[string]int{
			"a": 1,
			"b": 2,
			"c": 3,
		})
	})

	t.Run("empty map", func(t *testing.T) {
		p := conf.Map(map[string]any{
			"config": map[string]any{},
		})

		var s struct {
			Config map[string]int `value:"${config}"`
		}

		err := p.Bind(&s)
		assert.That(t, err).Nil()
		assert.That(t, s.Config).Equal(map[string]int{})
	})
}

func TestSliceBinding(t *testing.T) {
	t.Run("int slice from comma separated string", func(t *testing.T) {
		p := conf.Map(map[string]any{
			"numbers": "1,2,3,4,5",
		})

		var s struct {
			Numbers []int `value:"${numbers}"`
		}

		err := p.Bind(&s)
		assert.That(t, err).Nil()
		assert.That(t, s.Numbers).Equal([]int{1, 2, 3, 4, 5})
	})

	t.Run("string slice with whitespace", func(t *testing.T) {
		p := conf.Map(map[string]any{
			"values": " a , b , c ",
		})

		var s struct {
			Values []string `value:"${values}"`
		}

		err := p.Bind(&s)
		assert.That(t, err).Nil()
		assert.That(t, s.Values).Equal([]string{"a", "b", "c"})
	})

	t.Run("bool slice", func(t *testing.T) {
		p := conf.Map(map[string]any{
			"flags": "true,false,true",
		})

		var s struct {
			Flags []bool `value:"${flags}"`
		}

		err := p.Bind(&s)
		assert.That(t, err).Nil()
		assert.That(t, s.Flags).Equal([]bool{true, false, true})
	})
}

func TestStructBinding(t *testing.T) {
	t.Run("nested struct", func(t *testing.T) {
		p := conf.Map(map[string]any{
			"user": map[string]any{
				"name": "Alice",
				"age":  25,
			},
		})

		var s struct {
			User Data `value:"${user}"`
		}

		err := p.Bind(&s)
		assert.That(t, err).Nil()
		assert.That(t, s.User).Equal(Data{
			Name: "Alice",
			Age:  25,
		})
	})

	t.Run("embedded struct", func(t *testing.T) {
		p := conf.Map(map[string]any{
			"name": "Bob",
			"age":  30,
		})

		var s struct {
			Data
		}

		err := p.Bind(&s)
		assert.That(t, err).Nil()
		assert.That(t, s.Data).Equal(Data{
			Name: "Bob",
			Age:  30,
		})
	})
}
