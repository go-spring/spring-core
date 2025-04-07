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
	"fmt"
	"testing"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/util/assert"
)

type DB struct {
	UserName string `value:"${username}" expr:"len($)>=4"`
	Password string `value:"${password}"`
	Url      string `value:"${url}"`
	Port     string `value:"${port}"`
	DB       string `value:"${db}"`
}

type DbConfig struct {
	DB []DB `value:"${db}"`
}

type DBConnection struct {
	UserName string `value:"${username}"`
	Password string `value:"${password}"`
	Url      string `value:"${url}"`
	Port     string `value:"${port}"`
}

type UntaggedNestedDB struct {
	DBConnection
	DB string `value:"${db}"`
}

type TaggedNestedDB struct {
	DBConnection `value:"${tag}"`
	DB           string `value:"${db}"`
}

type TagNestedDbConfig struct {
	DB0 []TaggedNestedDB   `value:"${tagged.db}"`
	DB1 []UntaggedNestedDB `value:"${db}"`
}

type NestedDB struct {
	DBConnection
	DB string `value:"${db}"`
}

type NestedDbConfig struct {
	DB      []NestedDB     `value:"${db}"`
	Ints    []int          `value:"${:=}"`
	Map     map[string]int `value:"${:=}"`
	MapDef  map[string]int `value:"${x:=}"`
	Structs []struct {
		V string `value:"${v:=#v}"`
	} `value:"${:=}"`
}

type NestedDbMapConfig struct {
	DB map[string]NestedDB `value:"${db_map}"`
}

func TestProperties_Bind(t *testing.T) {

	t.Run("default", func(t *testing.T) {
		p := conf.New()
		v := &struct {
			S struct {
				V int `value:"${:=3}"`
			} `value:"${s:=}"`
		}{}
		err := p.Bind(v)
		assert.Nil(t, err)
		assert.Equal(t, v.S.V, 3)
	})

	t.Run("simple bind", func(t *testing.T) {
		p, err := conf.Load("testdata/config/app.yaml")
		assert.Nil(t, err)

		dbConfig1 := DbConfig{}
		err = p.Bind(&dbConfig1)
		assert.Nil(t, err)

		dbConfig2 := DbConfig{}
		err = p.Bind(&dbConfig2, "${prefix}")
		assert.Nil(t, err)

		// 实际上是取的两个节点，只是值是一样的而已
		assert.Equal(t, dbConfig1, dbConfig2)
	})

	t.Run("struct bind with tag", func(t *testing.T) {

		p, err := conf.Load("testdata/config/app.yaml")
		assert.Nil(t, err)

		dbConfig := TagNestedDbConfig{}
		err = p.Bind(&dbConfig)
		assert.Nil(t, err)

		fmt.Println(dbConfig)
	})

	t.Run("struct bind without tag", func(t *testing.T) {

		p, err := conf.Load("testdata/config/app.yaml")
		assert.Nil(t, err)

		dbConfig1 := NestedDbConfig{}
		err = p.Bind(&dbConfig1)
		assert.Nil(t, err)

		dbConfig2 := NestedDbConfig{}
		err = p.Bind(&dbConfig2, "${prefix}")
		assert.Nil(t, err)

		// 实际上是取的两个节点，只是值是一样的而已
		assert.Equal(t, dbConfig1, dbConfig2)
		assert.Equal(t, len(dbConfig1.DB), 2)
	})

	t.Run("simple map bind", func(t *testing.T) {

		p := conf.New()
		err := p.Set("a.b1", "b1")
		assert.Nil(t, err)
		err = p.Set("a.b2", "b2")
		assert.Nil(t, err)
		err = p.Set("a.b3", "b3")
		assert.Nil(t, err)

		var m map[string]string
		err = p.Bind(&m, "${a}")
		assert.Nil(t, err)

		assert.Equal(t, len(m), 3)
		assert.Equal(t, m["b1"], "b1")
	})

	t.Run("simple bind from file", func(t *testing.T) {

		p, err := conf.Load("testdata/config/app.yaml")
		assert.Nil(t, err)

		var m map[string]string
		err = p.Bind(&m, "${camera}")
		assert.Nil(t, err)

		assert.Equal(t, len(m), 3)
		assert.Equal(t, m["floor1"], "camera_floor1")
	})

	t.Run("struct bind from file", func(t *testing.T) {

		p, err := conf.Load("testdata/config/app.yaml")
		assert.Nil(t, err)

		var m map[string]NestedDB
		err = p.Bind(&m, "${db_map}")
		assert.Nil(t, err)

		assert.Equal(t, len(m), 2)
		assert.Equal(t, m["d1"].DB, "db1")

		dbConfig2 := NestedDbMapConfig{}
		err = p.Bind(&dbConfig2, "${prefix_map}")
		assert.Nil(t, err)

		assert.Equal(t, len(dbConfig2.DB), 2)
		assert.Equal(t, dbConfig2.DB["d1"].DB, "db1")
	})

	t.Run("ignore interface", func(t *testing.T) {
		p := conf.New()
		err := p.Bind(&struct{ fmt.Stringer }{})
		assert.Nil(t, err)
	})

	//t.Run("", func(t *testing.T) {
	//	p := conf.New()
	//	err := p.Bytes([]byte(`m:`), ".yaml")
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	var s struct {
	//		M map[string]string `value:"${m}"`
	//	}
	//	if err = p.Bind(&s); err != nil {
	//		t.Fatal(err)
	//	}
	//	assert.Equal(t, s.M, map[string]string{})
	//})
}

//func TestProperties_Load(t *testing.T) {
//
//	p := conf.New()
//	err := p.Load("testdata/config/app.yaml")
//	assert.Nil(t, err)
//	err = p.Load("testdata/config/app.properties")
//	assert.Nil(t, err)
//
//	for _, k := range p.Keys() {
//		fmt.Println(k+":", p.Get(k))
//	}
//
//	assert.True(t, p.Has("yaml.list"))
//	assert.True(t, p.Has("properties.list"))
//}

//func TestProperties_Get(t *testing.T) {
//
//	t.Run("base", func(t *testing.T) {
//
//		p := conf.Map(map[string]interface{}{
//			"a.b.d":       []string{"3"},
//			"Bool":        true,
//			"StringSlice": []string{"3", "4"},
//		})
//
//		err := p.Set("a.b.c", "3")
//		assert.Nil(t, err)
//		//err = p.Merge()
//		//assert.Nil(t, err)
//
//		v := p.Get("a.b.c")
//		assert.Equal(t, v, "3")
//		v = p.Get("a.b.d[0]")
//		assert.Equal(t, v, "3")
//
//		//err = p.Merge(map[string]interface{}{})
//		//assert.Nil(t, err)
//		err = p.Set("Int", "3")
//		assert.Nil(t, err)
//		err = p.Set("Uint", "3")
//		assert.Nil(t, err)
//		err = p.Set("Float", "3.0")
//		assert.Nil(t, err)
//		err = p.Set("String", "3")
//		assert.Nil(t, err)
//		err = p.Set("Duration", "3s")
//		assert.Nil(t, err)
//		//err = p.Merge(map[string]interface{}{})
//		//assert.Nil(t, err)
//		err = p.Set("Time", "2020-02-04 20:02:04")
//		assert.Nil(t, err)
//		//err = p.Merge(map[string]interface{}{
//		//	"MapStringInterface": []interface{}{
//		//		map[interface{}]interface{}{
//		//			"1": 2,
//		//		},
//		//	},
//		//})
//		//assert.Nil(t, err)
//
//		assert.False(t, p.Has("NULL"))
//		assert.Equal(t, p.Get("NULL"), "")
//
//		v = p.Get("NULL", "OK")
//		assert.Equal(t, v, "OK")
//
//		v = p.Get("Int")
//		assert.Equal(t, v, "3")
//
//		var v2 int
//		err = p.Bind(&v2, "${Int}")
//		assert.Nil(t, err)
//		assert.Equal(t, v2, 3)
//
//		var u2 uint
//		err = p.Bind(&u2, "${Uint}")
//		assert.Nil(t, err)
//		assert.Equal(t, u2, uint(3))
//
//		var u3 uint32
//		err = p.Bind(&u3, "${Uint}")
//		assert.Nil(t, err)
//		assert.Equal(t, u3, uint32(3))
//
//		var f2 float32
//		err = p.Bind(&f2, "${Float}")
//		assert.Nil(t, err)
//		assert.Equal(t, f2, float32(3))
//
//		v = p.Get("Bool")
//		b := cast.ToBool(v)
//		assert.Equal(t, b, true)
//
//		var b2 bool
//		err = p.Bind(&b2, "${Bool}")
//		assert.Nil(t, err)
//		assert.Equal(t, b2, true)
//
//		v = p.Get("Int")
//		i := cast.ToInt64(v)
//		assert.Equal(t, i, int64(3))
//
//		v = p.Get("Uint")
//		u := cast.ToUint64(v)
//		assert.Equal(t, u, uint64(3))
//
//		v = p.Get("Float")
//		f := cast.ToFloat64(v)
//		assert.Equal(t, f, 3.0)
//
//		v = p.Get("String")
//		s := cast.ToString(v)
//		assert.Equal(t, s, "3")
//
//		assert.False(t, p.Has("string"))
//		assert.Equal(t, p.Get("string"), "")
//
//		v = p.Get("Duration")
//		d := cast.ToDuration(v)
//		assert.Equal(t, d, time.Second*3)
//
//		var ti time.Time
//		err = p.Bind(&ti, "${Time}")
//		assert.Nil(t, err)
//		assert.Equal(t, ti, time.Date(2020, 02, 04, 20, 02, 04, 0, time.UTC))
//
//		var ss2 []string
//		err = p.Bind(&ss2, "${StringSlice}")
//		assert.Nil(t, err)
//		assert.Equal(t, ss2, []string{"3", "4"})
//	})
//
//	t.Run("slice slice", func(t *testing.T) {
//		p := conf.Map(map[string]interface{}{
//			"a": []interface{}{
//				[]interface{}{1, 2},
//				[]interface{}{3, 4},
//				map[string]interface{}{
//					"b": "c",
//					"d": []interface{}{5, 6},
//				},
//			},
//		})
//		v := p.Get("a[0][0]")
//		assert.Equal(t, v, "1")
//		v = p.Get("a[0][1]")
//		assert.Equal(t, v, "2")
//		v = p.Get("a[1][0]")
//		assert.Equal(t, v, "3")
//		v = p.Get("a[1][1]")
//		assert.Equal(t, v, "4")
//		v = p.Get("a[2].b")
//		assert.Equal(t, v, "c")
//		v = p.Get("a[2].d[0]")
//		assert.Equal(t, v, "5")
//		v = p.Get("a[2].d[1]")
//		assert.Equal(t, v, "6")
//	})
//}
//
//func TestProperties_Ref(t *testing.T) {
//
//	type FileLog struct {
//		Dir             string `value:"${dir:=${app.dir}}"`
//		NestedDir       string `value:"${nested.dir:=${nested.app.dir:=./log}}"`
//		NestedEmptyDir  string `value:"${nested.dir:=${nested.app.dir:=}}"`
//		NestedNestedDir string `value:"${nested.dir:=${nested.app.dir:=${nested.nested.app.dir:=./log}}}"`
//	}
//
//	var mqLog struct{ FileLog }
//	var httpLog struct{ FileLog }
//
//	t.Run("not config", func(t *testing.T) {
//		p := conf.New()
//		err := p.Bind(&httpLog)
//		assert.Error(t, err, "property app.dir not exist")
//	})
//
//	t.Run("config", func(t *testing.T) {
//		p := conf.New()
//
//		appDir := "/home/log"
//		err := p.Set("app.dir", appDir)
//		assert.Nil(t, err)
//
//		err = p.Bind(&httpLog)
//		assert.Nil(t, err)
//		assert.Equal(t, httpLog.Dir, appDir)
//		assert.Equal(t, httpLog.NestedDir, "./log")
//		assert.Equal(t, httpLog.NestedEmptyDir, "")
//		assert.Equal(t, httpLog.NestedNestedDir, "./log")
//
//		err = p.Bind(&mqLog)
//		assert.Nil(t, err)
//		assert.Equal(t, mqLog.Dir, appDir)
//		assert.Equal(t, mqLog.NestedDir, "./log")
//		assert.Equal(t, mqLog.NestedEmptyDir, "")
//		assert.Equal(t, mqLog.NestedNestedDir, "./log")
//	})
//
//	t.Run("empty key", func(t *testing.T) {
//		p := conf.New()
//		var s struct {
//			KeyIsEmpty string `value:"${:=kie}"`
//		}
//		err := p.Bind(&s)
//		assert.Nil(t, err)
//		assert.Equal(t, s.KeyIsEmpty, "kie")
//	})
//}
//
//func TestBindSlice(t *testing.T) {
//	p := conf.Map(map[string]interface{}{
//		"a": []string{"1", "2"},
//	})
//	var ss []string
//	err := p.Bind(&ss, "${a}")
//	if err != nil {
//		t.Fatal(err)
//	}
//	assert.Equal(t, ss, []string{"1", "2"})
//}
//
//func TestBindMap(t *testing.T) {
//
//	t.Run("", func(t *testing.T) {
//		var r [3]map[string]string
//		err := conf.New().Bind(&r)
//		assert.Error(t, err, "bind path=.* type=.* error: target should be value type")
//	})
//
//	t.Run("", func(t *testing.T) {
//		var r []map[string]string
//		err := conf.New().Bind(&r)
//		assert.Error(t, err, "bind path=.* type=.* error: target should be value type")
//	})
//
//	t.Run("", func(t *testing.T) {
//		var r map[string]map[string]string
//		err := conf.New().Bind(&r)
//		assert.Error(t, err, "bind path=.* type=.* error: target should be value type")
//	})
//
//	m := map[string]interface{}{
//		"a.b1": "ab1",
//		"a.b2": "ab2",
//		"a.b3": "ab3",
//		"b.b1": "bb1",
//		"b.b2": "bb2",
//		"b.b3": "bb3",
//	}
//
//	t.Run("", func(t *testing.T) {
//		var r map[string][3]map[string]string
//		p := conf.Map(m)
//		err := p.Bind(&r)
//		assert.Error(t, err, "bind path=.* type=.* error: target should be value type")
//	})
//
//	t.Run("", func(t *testing.T) {
//		var r map[string][]map[string]string
//		p := conf.Map(m)
//		err := p.Bind(&r)
//		assert.Error(t, err, "bind path=.* type=.* error: target should be value type")
//	})
//
//	t.Run("", func(t *testing.T) {
//		var r map[string]map[string]map[string]string
//		p := conf.Map(m)
//		err := p.Bind(&r)
//		assert.Error(t, err, "bind path=.* type=.* error: target should be value type")
//	})
//
//	t.Run("", func(t *testing.T) {
//		var r map[string]struct {
//			B1 string `value:"${b1}"`
//			B2 string `value:"${b2}"`
//			B3 string `value:"${b3}"`
//		}
//		p := conf.Map(m)
//		err := p.Bind(&r)
//		assert.Nil(t, err)
//		assert.Equal(t, r["a"].B1, "ab1")
//	})
//
//	t.Run("", func(t *testing.T) {
//		p := conf.Map(map[string]interface{}{"a.b1": "ab1"})
//		var r map[string]string
//		err := p.Bind(&r)
//		assert.Error(t, err, "bind path=map\\[string]string type=string error << property a isn't simple value")
//	})
//
//	t.Run("", func(t *testing.T) {
//		type S struct {
//			A map[string]string `value:"${a}"`
//			B map[string]string `value:"${b}"`
//		}
//		var r S
//		p := conf.Map(map[string]interface{}{
//			"a": "1", "b": 2,
//		})
//		err := p.Bind(&r)
//		assert.Error(t, err, "bind path=S\\.A type=map\\[string]string error << property conflict at path a")
//	})
//
//	t.Run("", func(t *testing.T) {
//		var r struct {
//			A map[string]string `value:"${a}"`
//			B map[string]string `value:"${b}"`
//		}
//		p := conf.Map(m)
//		err := p.Bind(&r)
//		assert.Nil(t, err)
//		assert.Equal(t, r.A["b1"], "ab1")
//	})
//}
//
//func TestResolve(t *testing.T) {
//	p := conf.New()
//	err := p.Set("name", "Jim")
//	assert.Nil(t, err)
//	_, err = p.Resolve("my name is ${name")
//	assert.Error(t, err, "resolve string \"my name is \\${name\" error: invalid syntax")
//	str, _ := p.Resolve("my name is ${name}")
//	assert.Equal(t, str, "my name is Jim")
//	str, _ = p.Resolve("my name is ${name}${name}")
//	assert.Equal(t, str, "my name is JimJim")
//	_, err = p.Resolve("my name is ${name} my name is ${name")
//	assert.NotNil(t, err)
//	str, _ = p.Resolve("my name is ${name} my name is ${name}")
//	assert.Equal(t, str, "my name is Jim my name is Jim")
//}
//
//func TestProperties_Has(t *testing.T) {
//	p := conf.Map(map[string]interface{}{
//		"a.b.c": "3",
//		"a.b.d": []string{"7", "8"},
//	})
//	assert.True(t, p.Has("a"))
//	assert.False(t, p.Has("a[0]"))
//	assert.True(t, p.Has("a.b"))
//	assert.False(t, p.Has("a.c"))
//	assert.True(t, p.Has("a.b.c"))
//	assert.True(t, p.Has("a.b.d"))
//	assert.True(t, p.Has("a.b.d[0]"))
//	assert.True(t, p.Has("a.b.d[1]"))
//	assert.False(t, p.Has("a.b.d[2]"))
//	assert.False(t, p.Has("a.b.e"))
//}
//
//func TestProperties_Set(t *testing.T) {
//
//	t.Run("map nil", func(t *testing.T) {
//
//		p := conf.Map(map[string]interface{}{
//			"m": map[string]string{"a": "b"},
//		})
//		//assert.False(t, p.Has("m"))
//		//assert.Equal(t, p.Get("m"), "")
//
//		//err = p.Merge(map[string]interface{}{
//		//	"m": map[string]string{"a": "b"},
//		//})
//		//assert.Nil(t, err)
//		//assert.Equal(t, p.Keys(), []string{"m.a"})
//
//		err := p.Set("m", "1")
//		assert.Error(t, err, "property conflict at path m")
//
//		//err = p.Merge(map[string]interface{}{
//		//	"m": []string{"b"},
//		//})
//		//assert.Error(t, err, "property conflict at path m\\[0]")
//	})
//
//	t.Run("map empty", func(t *testing.T) {
//		//p := conf.New()
//		//err := p.Merge(map[string]interface{}{
//		//	"m": map[string]string{},
//		//})
//		//assert.Nil(t, err)
//		//assert.True(t, p.Has("m"))
//		//assert.Equal(t, p.Get("m"), "")
//		//err := p.Merge(map[string]interface{}{
//		//	"m": map[string]string{"a": "b"},
//		//})
//		//assert.Nil(t, err)
//		//assert.Equal(t, p.Keys(), []string{"m.a"})
//	})
//
//	t.Run("list nil", func(t *testing.T) {
//		//p := conf.New()
//		//err := p.Merge(map[string]interface{}{
//		//	"a": nil,
//		//})
//		//assert.Nil(t, err)
//		//assert.False(t, p.Has("a"))
//		//assert.Equal(t, p.Get("a"), "")
//		//err = p.Merge(map[string]interface{}{
//		//	"a": []string{"b"},
//		//})
//		//assert.Nil(t, err)
//		//assert.Equal(t, p.Keys(), []string{"a[0]"})
//	})
//
//	t.Run("list empty", func(t *testing.T) {
//		//p := conf.New()
//		//err := p.Merge(map[string]interface{}{
//		//	"a": []string{},
//		//})
//		//assert.Nil(t, err)
//		//assert.True(t, p.Has("a"))
//		//assert.Equal(t, p.Get("a"), "")
//		//err := p.Merge(map[string]interface{}{
//		//	"a": []string{"b"},
//		//})
//		//assert.Nil(t, err)
//		//assert.Equal(t, p.Keys(), []string{"a[0]"})
//	})
//
//	t.Run("list", func(t *testing.T) {
//		p := conf.Map(map[string]interface{}{
//			"a": []string{"a", "aa", "aaa"},
//			"b": []int{1, 11, 111},
//			"c": []float32{1, 1.1, 1.11},
//		})
//		//err := p.Merge(map[string]interface{}{
//		//	"a": []string{"a", "aa", "aaa"},
//		//})
//		//assert.Nil(t, err)
//		//err = p.Merge(map[string]interface{}{
//		//	"b": []int{1, 11, 111},
//		//})
//		//assert.Nil(t, err)
//		//err = p.Merge(map[string]interface{}{
//		//	"c": []float32{1, 1.1, 1.11},
//		//})
//		//assert.Nil(t, err)
//		assert.Equal(t, p.Get("a[0]"), "a")
//		assert.Equal(t, p.Get("a[1]"), "aa")
//		assert.Equal(t, p.Get("a[2]"), "aaa")
//		assert.Equal(t, p.Get("b[0]"), "1")
//		assert.Equal(t, p.Get("b[1]"), "11")
//		assert.Equal(t, p.Get("b[2]"), "111")
//		assert.Equal(t, p.Get("c[0]"), "1")
//		assert.Equal(t, p.Get("c[1]"), "1.1")
//		assert.Equal(t, p.Get("c[2]"), "1.11")
//	})
//}
//
//func PointConverter(val string) (image.Point, error) {
//	if !(strings.HasPrefix(val, "(") && strings.HasSuffix(val, ")")) {
//		return image.Point{}, errors.New("数据格式错误")
//	}
//	ss := strings.Split(val[1:len(val)-1], ",")
//	x := cast.ToInt(ss[0])
//	y := cast.ToInt(ss[1])
//	return image.Point{X: x, Y: y}, nil
//}
//
//func PointSplitter(str string) ([]string, error) {
//	var ret []string
//	var lastIndex int
//	for i, c := range str {
//		if c == ')' {
//			ret = append(ret, str[lastIndex:i+1])
//			lastIndex = i + 1
//		}
//	}
//	return ret, nil
//}
//
//func TestSplitter(t *testing.T) {
//	conf.RegisterConverter(PointConverter)
//	conf.RegisterSplitter("PointSplitter", PointSplitter)
//	var points []image.Point
//	err := conf.New().Bind(&points, "${:=(1,2)(3,4)}>>PointSplitter")
//	assert.Nil(t, err)
//	assert.Equal(t, points, []image.Point{{X: 1, Y: 2}, {X: 3, Y: 4}})
//}
