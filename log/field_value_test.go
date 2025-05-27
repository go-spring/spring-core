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
	"bytes"
	"fmt"
	"strconv"
	"testing"

	"github.com/lvan100/go-assert"
)

var _ Encoder = (*MockEncoder)(nil)

type MockEncoder struct {
	AppendEncoderBeginFunc func()
	AppendEncoderEndFunc   func()
	AppendObjectBeginFunc  func()
	AppendObjectEndFunc    func()
	AppendArrayBeginFunc   func()
	AppendArrayEndFunc     func()
	AppendKeyFunc          func(key string)
	AppendBoolFunc         func(v bool)
	AppendInt64Func        func(v int64)
	AppendUint64Func       func(v uint64)
	AppendFloat64Func      func(v float64)
	AppendStringFunc       func(v string)
	AppendReflectFunc      func(v interface{})
}

func (m MockEncoder) AppendEncoderBegin() {
	if m.AppendEncoderBeginFunc != nil {
		m.AppendEncoderBeginFunc()
	}
}

func (m MockEncoder) AppendEncoderEnd() {
	if m.AppendEncoderEndFunc != nil {
		m.AppendEncoderEndFunc()
	}
}

func (m MockEncoder) AppendObjectBegin() {
	if m.AppendObjectBeginFunc != nil {
		m.AppendObjectBeginFunc()
	}
}

func (m MockEncoder) AppendObjectEnd() {
	if m.AppendObjectEndFunc != nil {
		m.AppendObjectEndFunc()
	}
}

func (m MockEncoder) AppendArrayBegin() {
	if m.AppendArrayBeginFunc != nil {
		m.AppendArrayBeginFunc()
	}
}

func (m MockEncoder) AppendArrayEnd() {
	if m.AppendArrayEndFunc != nil {
		m.AppendArrayEndFunc()
	}
}

func (m MockEncoder) AppendKey(key string) {
	if m.AppendKeyFunc != nil {
		m.AppendKeyFunc(key)
	}
}

func (m MockEncoder) AppendBool(v bool) {
	if m.AppendBoolFunc != nil {
		m.AppendBoolFunc(v)
	}
}

func (m MockEncoder) AppendInt64(v int64) {
	if m.AppendInt64Func != nil {
		m.AppendInt64Func(v)
	}
}

func (m MockEncoder) AppendUint64(v uint64) {
	if m.AppendUint64Func != nil {
		m.AppendUint64Func(v)
	}
}

func (m MockEncoder) AppendFloat64(v float64) {
	if m.AppendFloat64Func != nil {
		m.AppendFloat64Func(v)
	}
}

func (m MockEncoder) AppendString(v string) {
	if m.AppendStringFunc != nil {
		m.AppendStringFunc(v)
	}
}

func (m MockEncoder) AppendReflect(v interface{}) {
	if m.AppendReflectFunc != nil {
		m.AppendReflectFunc(v)
	}
}

func TestBoolValue(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{AppendBoolFunc: func(v bool) {
			buf.WriteString(strconv.FormatBool(v))
		}}

		BoolValue(true).Encode(enc)
		assert.ThatString(t, buf.String()).Equal("true")

		buf.Reset()
		BoolValue(false).Encode(enc)
		assert.ThatString(t, buf.String()).Equal("false")
	})
}

func TestInt64Value(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{AppendInt64Func: func(v int64) {
			buf.WriteString(strconv.FormatInt(v, 10))
		}}

		Int64Value(9).Encode(enc)
		assert.ThatString(t, buf.String()).Equal("9")
	})
}

func TestUint64Value(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{AppendUint64Func: func(v uint64) {
			buf.WriteString(strconv.FormatUint(v, 10))
		}}

		Uint64Value(9).Encode(enc)
		assert.ThatString(t, buf.String()).Equal("9")
	})
}

func TestFloat64Value(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{AppendFloat64Func: func(v float64) {
			buf.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
		}}

		Float64Value(9.9).Encode(enc)
		assert.ThatString(t, buf.String()).Equal("9.9")
	})
}

func TestStringValue(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{AppendStringFunc: func(v string) {
			buf.WriteString(v)
		}}

		StringValue("9.9").Encode(enc)
		assert.ThatString(t, buf.String()).Equal("9.9")
	})
}

func TestReflectValue(t *testing.T) {
	type A struct {
		V string
	}

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{AppendReflectFunc: func(v any) {
			buf.WriteString(fmt.Sprint(v))
		}}

		v := &ReflectValue{Val: &A{V: "a"}}
		v.Encode(enc)
		assert.ThatString(t, buf.String()).Equal("&{a}")
	})
}

func TestBoolsValue(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() {
				buf.WriteString("[")
			},
			AppendArrayEndFunc: func() {
				buf.WriteString("]")
			},
			AppendBoolFunc: func(v bool) {
				if buf.Len() > 1 {
					buf.WriteString(",")
				}
				buf.WriteString(strconv.FormatBool(v))
			},
		}

		BoolsValue([]bool{true, false, true}).Encode(enc)
		assert.ThatString(t, buf.String()).Equal("[true,false,true]")
	})
}

func TestIntsValue(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() {
				buf.WriteString("[")
			},
			AppendArrayEndFunc: func() {
				buf.WriteString("]")
			},
			AppendInt64Func: func(v int64) {
				if buf.Len() > 1 {
					buf.WriteString(",")
				}
				buf.WriteString(strconv.FormatInt(v, 10))
			},
		}

		IntsValue([]int{1, 2, 3}).Encode(enc)
		assert.ThatString(t, buf.String()).Equal("[1,2,3]")
	})
}

func TestInt8sValue(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() {
				buf.WriteString("[")
			},
			AppendArrayEndFunc: func() {
				buf.WriteString("]")
			},
			AppendInt64Func: func(v int64) {
				if buf.Len() > 1 {
					buf.WriteString(",")
				}
				buf.WriteString(strconv.FormatInt(v, 10))
			},
		}

		Int8sValue([]int8{1, 2, 3}).Encode(enc)
		assert.ThatString(t, buf.String()).Equal("[1,2,3]")
	})
}

func TestInt16sValue(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() {
				buf.WriteString("[")
			},
			AppendArrayEndFunc: func() {
				buf.WriteString("]")
			},
			AppendInt64Func: func(v int64) {
				if buf.Len() > 1 {
					buf.WriteString(",")
				}
				buf.WriteString(strconv.FormatInt(v, 10))
			},
		}

		Int16sValue([]int16{1, 2, 3}).Encode(enc)
		assert.ThatString(t, buf.String()).Equal("[1,2,3]")
	})
}

func TestInt32sValue(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() {
				buf.WriteString("[")
			},
			AppendArrayEndFunc: func() {
				buf.WriteString("]")
			},
			AppendInt64Func: func(v int64) {
				if buf.Len() > 1 {
					buf.WriteString(",")
				}
				buf.WriteString(strconv.FormatInt(v, 10))
			},
		}

		Int32sValue([]int32{1, 2, 3}).Encode(enc)
		assert.ThatString(t, buf.String()).Equal("[1,2,3]")
	})
}

func TestInt64sValue(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() {
				buf.WriteString("[")
			},
			AppendArrayEndFunc: func() {
				buf.WriteString("]")
			},
			AppendInt64Func: func(v int64) {
				if buf.Len() > 1 {
					buf.WriteString(",")
				}
				buf.WriteString(strconv.FormatInt(v, 10))
			},
		}

		Int64sValue([]int64{1, 2, 3}).Encode(enc)
		assert.ThatString(t, buf.String()).Equal("[1,2,3]")
	})
}

func TestUintsValue(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() {
				buf.WriteString("[")
			},
			AppendArrayEndFunc: func() {
				buf.WriteString("]")
			},
			AppendUint64Func: func(v uint64) {
				if buf.Len() > 1 {
					buf.WriteString(",")
				}
				buf.WriteString(strconv.FormatUint(v, 10))
			},
		}

		UintsValue([]uint{1, 2, 3}).Encode(enc)
		assert.ThatString(t, buf.String()).Equal("[1,2,3]")
	})
}

func TestUint8sValue(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() {
				buf.WriteString("[")
			},
			AppendArrayEndFunc: func() {
				buf.WriteString("]")
			},
			AppendUint64Func: func(v uint64) {
				if buf.Len() > 1 {
					buf.WriteString(",")
				}
				buf.WriteString(strconv.FormatUint(v, 10))
			},
		}

		Uint8sValue([]uint8{1, 2, 3}).Encode(enc)
		assert.ThatString(t, buf.String()).Equal("[1,2,3]")
	})
}

func TestUint16sValue(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() {
				buf.WriteString("[")
			},
			AppendArrayEndFunc: func() {
				buf.WriteString("]")
			},
			AppendUint64Func: func(v uint64) {
				if buf.Len() > 1 {
					buf.WriteString(",")
				}
				buf.WriteString(strconv.FormatUint(v, 10))
			},
		}

		Uint16sValue([]uint16{1, 2, 3}).Encode(enc)
		assert.ThatString(t, buf.String()).Equal("[1,2,3]")
	})
}

func TestUint32sValue(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() {
				buf.WriteString("[")
			},
			AppendArrayEndFunc: func() {
				buf.WriteString("]")
			},
			AppendUint64Func: func(v uint64) {
				if buf.Len() > 1 {
					buf.WriteString(",")
				}
				buf.WriteString(strconv.FormatUint(v, 10))
			},
		}

		Uint32sValue([]uint32{1, 2, 3}).Encode(enc)
		assert.ThatString(t, buf.String()).Equal("[1,2,3]")
	})
}

func TestUint64sValue(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() {
				buf.WriteString("[")
			},
			AppendArrayEndFunc: func() {
				buf.WriteString("]")
			},
			AppendUint64Func: func(v uint64) {
				if buf.Len() > 1 {
					buf.WriteString(",")
				}
				buf.WriteString(strconv.FormatUint(v, 10))
			},
		}

		Uint64sValue([]uint64{1, 2, 3}).Encode(enc)
		assert.ThatString(t, buf.String()).Equal("[1,2,3]")
	})
}

func TestFloat32sValue(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() {
				buf.WriteString("[")
			},
			AppendArrayEndFunc: func() {
				buf.WriteString("]")
			},
			AppendFloat64Func: func(v float64) {
				if buf.Len() > 1 {
					buf.WriteString(",")
				}
				buf.WriteString(strconv.FormatFloat(v, 'f', 1, 64))
			},
		}

		Float32sValue([]float32{1.1, 2.2, 3.3}).Encode(enc)
		assert.ThatString(t, buf.String()).Equal("[1.1,2.2,3.3]")
	})
}

func TestFloat64sValue(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() {
				buf.WriteString("[")
			},
			AppendArrayEndFunc: func() {
				buf.WriteString("]")
			},
			AppendFloat64Func: func(v float64) {
				if buf.Len() > 1 {
					buf.WriteString(",")
				}
				buf.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
			},
		}

		Float64sValue([]float64{1.1, 2.2, 3.3}).Encode(enc)
		assert.ThatString(t, buf.String()).Equal("[1.1,2.2,3.3]")
	})
}

func TestStringsValue(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() {
				buf.WriteString("[")
			},
			AppendArrayEndFunc: func() {
				buf.WriteString("]")
			},
			AppendStringFunc: func(v string) {
				if buf.Len() > 1 {
					buf.WriteString(",")
				}
				buf.WriteString(v)
			},
		}

		StringsValue([]string{"a", "b", "c"}).Encode(enc)
		assert.ThatString(t, buf.String()).Equal("[a,b,c]")
	})
}

func TestObjectValue(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendObjectBeginFunc: func() {
				buf.WriteString("{")
			},
			AppendObjectEndFunc: func() {
				buf.WriteString("}")
			},
			AppendKeyFunc: func(key string) {
				if buf.Len() > 1 {
					buf.WriteString(",")
				}
				buf.WriteString(strconv.Quote(key))
			},
			AppendStringFunc: func(v string) {
				buf.WriteString(":" + strconv.Quote(v))
			},
		}

		fields := []Field{{Key: "key", Val: StringValue("value")}}
		ObjectValue(fields).Encode(enc)
		assert.ThatString(t, buf.String()).Equal(`{"key":"value"}`)
	})
}

func TestArrayValue(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		var buf bytes.Buffer
		enc := MockEncoder{
			AppendArrayBeginFunc: func() {
				buf.WriteString("[")
			},
			AppendArrayEndFunc: func() {
				buf.WriteString("]")
			},
			AppendStringFunc: func(v string) {
				if buf.Len() > 1 {
					buf.WriteString(",")
				}
				buf.WriteString(v)
			},
		}

		values := []Value{StringValue("a"), StringValue("b"), StringValue("c")}
		ArrayValue(values).Encode(enc)
		assert.ThatString(t, buf.String()).Equal(`[a,b,c]`)
	})
}
