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
	"strings"
	"testing"
	"time"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/util/assert"
	"github.com/spf13/cast"
)

func init() {
	conf.RegisterConverter(PointConverter)
	conf.RegisterSplitter("PointSplitter", PointSplitter)
}

func PointConverter(val string) (image.Point, error) {
	ss := strings.Split(val[1:len(val)-1], ",")
	x := cast.ToInt(ss[0])
	y := cast.ToInt(ss[1])
	return image.Point{X: x, Y: y}, nil
}

func PointSplitter(str string) ([]string, error) {
	if !(strings.HasPrefix(str, "(") && strings.HasSuffix(str, ")")) {
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
		assert.Nil(t, err)
		assert.Equal(t, s.Time, time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC))
		assert.Equal(t, s.Duration, 10*time.Second)
	})

	t.Run("error", func(t *testing.T) {
		p := conf.Map(map[string]interface{}{
			"time": "2025-02-01M00:00:00",
		})
		err := p.Bind(&s)
		assert.Error(t, err, "unable to parse date: 2025-02-01M00:00:00")
	})
}

func TestSplitter(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		var points []image.Point
		err := conf.New().Bind(&points, "${:=(1,2)(3,4)}>>PointSplitter")
		assert.Nil(t, err)
		assert.Equal(t, points, []image.Point{{X: 1, Y: 2}, {X: 3, Y: 4}})
	})

	t.Run("error", func(t *testing.T) {
		var points []image.Point
		err := conf.New().Bind(&points, "${:=(1}>>PointSplitter")
		assert.Error(t, err, "split error")
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
	Ints     []int          `value:"${:=}"`
	Map      map[string]int `value:"${:=}"`
}

type DBConfig struct {
	DB0   []TaggedNestedDB   `value:"${tagged.db}"`
	DB1   []UntaggedNestedDB `value:"${db}"`
	Extra Extra              `value:"${extra}"`
}

func TestProperties_Bind(t *testing.T) {

	t.Run("target error - 1", func(t *testing.T) {
		err := conf.New().Bind(5)
		assert.Error(t, err, "should be a ptr")
	})

	t.Run("target error - 1", func(t *testing.T) {
		err := conf.New().Bind(new(*int))
		assert.Error(t, err, "target should be value type")
	})

	t.Run("array error", func(t *testing.T) {
		err := conf.New().Bind(new(struct {
			Arr [3]string `value:"${arr:=1,2,3}"`
		}))
		assert.Error(t, err, "use slice instead of array")
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
				Ints:     []int{},
				Map:      map[string]int{},
			},
		}

		p, err := conf.Load("./testdata/config/app.yaml")
		assert.Nil(t, err)

		var c DBConfig
		err = p.Bind(&c)
		assert.Nil(t, err)
		assert.Equal(t, c, expect)

		err = p.Bind(&c, "${prefix}")
		assert.Nil(t, err)
		assert.Equal(t, c, expect)
	})
}

//type NestedDbMapConfig struct {
//	DB map[string]NestedDB `value:"${db_map}"`
//}
//
//func TestProperties_Bind(t *testing.T) {
//
//	t.Run("default", func(t *testing.T) {
//		p := conf.New()
//		v := &struct {
//			S struct {
//				V int `value:"${:=3}"`
//			} `value:"${s:=}"`
//		}{}
//		err := p.Bind(v)
//		assert.Nil(t, err)
//		assert.Equal(t, v.S.V, 3)
//	})
//
//	t.Run("simple bind", func(t *testing.T) {
//		p, err := conf.Load("testdata/config/app.yaml")
//		assert.Nil(t, err)
//
//		dbConfig1 := DbConfig{}
//		err = p.Bind(&dbConfig1)
//		assert.Nil(t, err)
//
//		dbConfig2 := DbConfig{}
//		err = p.Bind(&dbConfig2, "${prefix}")
//		assert.Nil(t, err)
//
//		// 实际上是取的两个节点，只是值是一样的而已
//		assert.Equal(t, dbConfig1, dbConfig2)
//	})
//
//	t.Run("struct bind with tag", func(t *testing.T) {
//
//		p, err := conf.Load("testdata/config/app.yaml")
//		assert.Nil(t, err)
//
//		dbConfig := TagNestedDbConfig{}
//		err = p.Bind(&dbConfig)
//		assert.Nil(t, err)
//
//		fmt.Println(dbConfig)
//	})
//
//	t.Run("struct bind without tag", func(t *testing.T) {
//
//		p, err := conf.Load("testdata/config/app.yaml")
//		assert.Nil(t, err)
//
//		dbConfig1 := NestedDbConfig{}
//		err = p.Bind(&dbConfig1)
//		assert.Nil(t, err)
//
//		dbConfig2 := NestedDbConfig{}
//		err = p.Bind(&dbConfig2, "${prefix}")
//		assert.Nil(t, err)
//
//		// 实际上是取的两个节点，只是值是一样的而已
//		assert.Equal(t, dbConfig1, dbConfig2)
//		assert.Equal(t, len(dbConfig1.DB), 2)
//	})
//
//	t.Run("simple map bind", func(t *testing.T) {
//
//		p := conf.New()
//		err := p.Set("a.b1", "b1")
//		assert.Nil(t, err)
//		err = p.Set("a.b2", "b2")
//		assert.Nil(t, err)
//		err = p.Set("a.b3", "b3")
//		assert.Nil(t, err)
//
//		var m map[string]string
//		err = p.Bind(&m, "${a}")
//		assert.Nil(t, err)
//
//		assert.Equal(t, len(m), 3)
//		assert.Equal(t, m["b1"], "b1")
//	})
//
//	t.Run("simple bind from file", func(t *testing.T) {
//
//		p, err := conf.Load("testdata/config/app.yaml")
//		assert.Nil(t, err)
//
//		var m map[string]string
//		err = p.Bind(&m, "${camera}")
//		assert.Nil(t, err)
//
//		assert.Equal(t, len(m), 3)
//		assert.Equal(t, m["floor1"], "camera_floor1")
//	})
//
//	t.Run("struct bind from file", func(t *testing.T) {
//
//		p, err := conf.Load("testdata/config/app.yaml")
//		assert.Nil(t, err)
//
//		var m map[string]NestedDB
//		err = p.Bind(&m, "${db_map}")
//		assert.Nil(t, err)
//
//		assert.Equal(t, len(m), 2)
//		assert.Equal(t, m["d1"].DB, "db1")
//
//		dbConfig2 := NestedDbMapConfig{}
//		err = p.Bind(&dbConfig2, "${prefix_map}")
//		assert.Nil(t, err)
//
//		assert.Equal(t, len(dbConfig2.DB), 2)
//		assert.Equal(t, dbConfig2.DB["d1"].DB, "db1")
//	})
//
//	t.Run("ignore interface", func(t *testing.T) {
//		p := conf.New()
//		err := p.Bind(&struct{ fmt.Stringer }{})
//		assert.Nil(t, err)
//	})
//
//	//t.Run("", func(t *testing.T) {
//	//	p := conf.New()
//	//	err := p.Bytes([]byte(`m:`), ".yaml")
//	//	if err != nil {
//	//		t.Fatal(err)
//	//	}
//	//	var s struct {
//	//		M map[string]string `value:"${m}"`
//	//	}
//	//	if err = p.Bind(&s); err != nil {
//	//		t.Fatal(err)
//	//	}
//	//	assert.Equal(t, s.M, map[string]string{})
//	//})
//}
