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
	"testing"

	"github.com/lvan100/go-assert"
)

func ptr[T any](i T) *T {
	return &i
}

var testFields = []Field{
	Msgf("hello %s", "中国"),
	Msg("hello world\n\\\t\"\r"),
	Any("null", nil),
	Any("bool", false),
	Any("bool_ptr", ptr(true)),
	Any("bool_ptr_nil", (*bool)(nil)),
	Any("bools", []bool{true, true, false}),
	Any("int", int(1)),
	Any("int_ptr", ptr(int(1))),
	Any("int_ptr_nil", (*int)(nil)),
	Any("int_slice", []int{int(1), int(2), int(3)}),
	Any("int8", int8(1)),
	Any("int8_ptr", ptr(int8(1))),
	Any("int8_ptr_nil", (*int8)(nil)),
	Any("int8_slice", []int8{int8(1), int8(2), int8(3)}),
	Any("int16", int16(1)),
	Any("int16_ptr", ptr(int16(1))),
	Any("int16_ptr_nil", (*int16)(nil)),
	Any("int16_slice", []int16{int16(1), int16(2), int16(3)}),
	Any("int32", int32(1)),
	Any("int32_ptr", ptr(int32(1))),
	Any("int32_ptr_nil", (*int32)(nil)),
	Any("int32_slice", []int32{int32(1), int32(2), int32(3)}),
	Any("int64", int64(1)),
	Any("int64_ptr", ptr(int64(1))),
	Any("int64_ptr_nil", (*int64)(nil)),
	Any("int64_slice", []int64{int64(1), int64(2), int64(3)}),
	Any("uint", uint(1)),
	Any("uint_ptr", ptr(uint(1))),
	Any("uint_ptr_nil", (*uint)(nil)),
	Any("uint_slice", []uint{uint(1), uint(2), uint(3)}),
	Any("uint8", uint8(1)),
	Any("uint8_ptr", ptr(uint8(1))),
	Any("uint8_ptr_nil", (*uint8)(nil)),
	Any("uint8_slice", []uint8{uint8(1), uint8(2), uint8(3)}),
	Any("uint16", uint16(1)),
	Any("uint16_ptr", ptr(uint16(1))),
	Any("uint16_ptr_nil", (*uint16)(nil)),
	Any("uint16_slice", []uint16{uint16(1), uint16(2), uint16(3)}),
	Any("uint32", uint32(1)),
	Any("uint32_ptr", ptr(uint32(1))),
	Any("uint32_ptr_nil", (*uint32)(nil)),
	Any("uint32_slice", []uint32{uint32(1), uint32(2), uint32(3)}),
	Any("uint64", uint64(1)),
	Any("uint64_ptr", ptr(uint64(1))),
	Any("uint64_ptr_nil", (*uint64)(nil)),
	Any("uint64_slice", []uint64{uint64(1), uint64(2), uint64(3)}),
	Any("float32", float32(1)),
	Any("float32_ptr", ptr(float32(1))),
	Any("float32_ptr_nil", (*float32)(nil)),
	Any("float32_slice", []float32{float32(1), float32(2), float32(3)}),
	Any("float64", float64(1)),
	Any("float64_ptr", ptr(float64(1))),
	Any("float64_ptr_nil", (*float64)(nil)),
	Any("float64_slice", []float64{float64(1), float64(2), float64(3)}),
	Any("string", "\x80\xC2\xED\xA0\x08"),
	Any("string_ptr", ptr("a")),
	Any("string_ptr_nil", (*string)(nil)),
	Any("string_slice", []string{"a", "b", "c"}),
	Object("object", Any("int64", int64(1)), Any("uint64", uint64(1)), Any("string", "a")),
	Any("struct", struct{ Int64 int64 }{10}),
}

func TestJSONEncoder(t *testing.T) {

	t.Run("chan error", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		enc := NewJSONEncoder(buf)
		enc.AppendEncoderBegin()
		enc.AppendKey("chan")
		enc.AppendReflect(make(chan error))
		enc.AppendEncoderEnd()
		assert.ThatString(t, buf.String()).Equal(`{"chan":"json: unsupported type: chan error"}`)
	})

	t.Run("success", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		enc := NewJSONEncoder(buf)
		enc.AppendEncoderBegin()
		WriteFields(enc, testFields)
		enc.AppendEncoderEnd()
		assert.ThatString(t, buf.String()).JsonEqual(`{
	    "msg": "hello world\n\\\t\"\r",
	    "null": null,
	    "bool": false,
	    "bool_ptr": true,
	    "bool_ptr_nil": null,
	    "bools": [
	        true,
	        true,
	        false
	    ],
	    "int": 1,
	    "int_ptr": 1,
	    "int_ptr_nil": null,
	    "int_slice": [
	        1,
	        2,
	        3
	    ],
	    "int8": 1,
	    "int8_ptr": 1,
	    "int8_ptr_nil": null,
	    "int8_slice": [
	        1,
	        2,
	        3
	    ],
	    "int16": 1,
	    "int16_ptr": 1,
	    "int16_ptr_nil": null,
	    "int16_slice": [
	        1,
	        2,
	        3
	    ],
	    "int32": 1,
	    "int32_ptr": 1,
	    "int32_ptr_nil": null,
	    "int32_slice": [
	        1,
	        2,
	        3
	    ],
	    "int64": 1,
	    "int64_ptr": 1,
	    "int64_ptr_nil": null,
	    "int64_slice": [
	        1,
	        2,
	        3
	    ],
	    "uint": 1,
	    "uint_ptr": 1,
	    "uint_ptr_nil": null,
	    "uint_slice": [
	        1,
	        2,
	        3
	    ],
	    "uint8": 1,
	    "uint8_ptr": 1,
	    "uint8_ptr_nil": null,
	    "uint8_slice": [
	        1,
	        2,
	        3
	    ],
	    "uint16": 1,
	    "uint16_ptr": 1,
	    "uint16_ptr_nil": null,
	    "uint16_slice": [
	        1,
	        2,
	        3
	    ],
	    "uint32": 1,
	    "uint32_ptr": 1,
	    "uint32_ptr_nil": null,
	    "uint32_slice": [
	        1,
	        2,
	        3
	    ],
	    "uint64": 1,
	    "uint64_ptr": 1,
	    "uint64_ptr_nil": null,
	    "uint64_slice": [
	        1,
	        2,
	        3
	    ],
	    "float32": 1,
	    "float32_ptr": 1,
	    "float32_ptr_nil": null,
	    "float32_slice": [
	        1,
	        2,
	        3
	    ],
	    "float64": 1,
	    "float64_ptr": 1,
	    "float64_ptr_nil": null,
	    "float64_slice": [
	        1,
	        2,
	        3
	    ],
	    "string": "\ufffd\ufffd\ufffd\ufffd\u0008",
	    "string_ptr": "a",
	    "string_ptr_nil": null,
	    "string_slice": [
	        "a",
	        "b",
	        "c"
	    ],
	    "object": {
	        "int64": 1,
	        "uint64": 1,
	        "string": "a"
	    },
		"struct": {
			"Int64": 10
		}
	}`)
	})
}

func TestTextEncoder(t *testing.T) {

	t.Run("chan error", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		enc := NewTextEncoder(buf, "||")
		enc.AppendEncoderBegin()
		enc.AppendKey("chan")
		enc.AppendReflect(make(chan error))
		enc.AppendEncoderEnd()
		assert.ThatString(t, buf.String()).Equal("chan=json: unsupported type: chan error")
	})

	t.Run("success", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		enc := NewTextEncoder(buf, "||")
		enc.AppendEncoderBegin()
		WriteFields(enc, testFields)
		{
			enc.AppendKey("object_2")
			enc.AppendObjectBegin()
			enc.AppendKey("map")
			enc.AppendReflect(map[string]int{"a": 1})
			enc.AppendObjectEnd()
		}
		{
			enc.AppendKey("array_2")
			enc.AppendArrayBegin()
			enc.AppendReflect(map[string]int{"a": 1})
			enc.AppendReflect(map[string]int{"a": 1})
			enc.AppendArrayEnd()
		}
		enc.AppendEncoderEnd()
		const expect = `msg=hello 中国||msg=hello world\n\\\t\"\r||null=null||` +
			`bool=false||bool_ptr=true||bool_ptr_nil=null||bools=[true,true,false]||` +
			`int=1||int_ptr=1||int_ptr_nil=null||int_slice=[1,2,3]||` +
			`int8=1||int8_ptr=1||int8_ptr_nil=null||int8_slice=[1,2,3]||` +
			`int16=1||int16_ptr=1||int16_ptr_nil=null||int16_slice=[1,2,3]||` +
			`int32=1||int32_ptr=1||int32_ptr_nil=null||int32_slice=[1,2,3]||` +
			`int64=1||int64_ptr=1||int64_ptr_nil=null||int64_slice=[1,2,3]||` +
			`uint=1||uint_ptr=1||uint_ptr_nil=null||uint_slice=[1,2,3]||` +
			`uint8=1||uint8_ptr=1||uint8_ptr_nil=null||uint8_slice=[1,2,3]||` +
			`uint16=1||uint16_ptr=1||uint16_ptr_nil=null||uint16_slice=[1,2,3]||` +
			`uint32=1||uint32_ptr=1||uint32_ptr_nil=null||uint32_slice=[1,2,3]||` +
			`uint64=1||uint64_ptr=1||uint64_ptr_nil=null||uint64_slice=[1,2,3]||` +
			`float32=1||float32_ptr=1||float32_ptr_nil=null||float32_slice=[1,2,3]||` +
			`float64=1||float64_ptr=1||float64_ptr_nil=null||float64_slice=[1,2,3]||` +
			`string=\ufffd\ufffd\ufffd\ufffd\u0008||string_ptr=a||string_ptr_nil=null||string_slice=["a","b","c"]||` +
			`object={"int64":1,"uint64":1,"string":"a"}||struct={"Int64":10}||` +
			`object_2={"map":{"a":1}}||array_2=[{"a":1},{"a":1}]`
		assert.ThatString(t, buf.String()).Equal(expect)
	})
}
