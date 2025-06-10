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
	"fmt"
	"math"
	"unsafe"

	"github.com/go-spring/spring-core/util"
)

const MsgKey = "msg"

// ValueType represents the type of value stored in a Field.
type ValueType int

const (
	ValueTypeBool = ValueType(iota)
	ValueTypeInt64
	ValueTypeUint64
	ValueTypeFloat64
	ValueTypeString
	ValueTypeReflect
	ValueTypeArray
	ValueTypeObject
)

// Field represents a structured log field with a key and a value.
type Field struct {
	Key  string
	Type ValueType
	Num  uint64
	Any  any
}

// Msg creates a string Field with the "msg" key.
func Msg(msg string) Field {
	return String(MsgKey, msg)
}

// Msgf formats a message using fmt.Sprintf and creates a string Field with "msg" key.
func Msgf(format string, args ...any) Field {
	return String(MsgKey, fmt.Sprintf(format, args...))
}

// Nil creates a Field with a nil value.
func Nil(key string) Field {
	return Reflect(key, nil)
}

// Bool creates a Field for a boolean value.
func Bool(key string, val bool) Field {
	if val {
		return Field{
			Key:  key,
			Type: ValueTypeBool,
			Num:  1,
		}
	}
	return Field{
		Key:  key,
		Type: ValueTypeBool,
		Num:  0,
	}
}

// BoolPtr creates a Field from a *bool, or a nil Field if the pointer is nil.
func BoolPtr(key string, val *bool) Field {
	if val == nil {
		return Nil(key)
	}
	return Bool(key, *val)
}

// Int creates a Field for an int value.
func Int[T util.IntType](key string, val T) Field {
	return Field{
		Key:  key,
		Type: ValueTypeInt64,
		Num:  uint64(val),
	}
}

// IntPtr creates a Field from a *int, or returns Nil if pointer is nil.
func IntPtr[T util.IntType](key string, val *T) Field {
	if val == nil {
		return Nil(key)
	}
	return Int(key, *val)
}

// Uint creates a Field for an uint value.
func Uint[T util.UintType](key string, val T) Field {
	return Field{
		Key:  key,
		Type: ValueTypeUint64,
		Num:  uint64(val),
	}
}

// UintPtr creates a Field from a *uint, or returns Nil if pointer is nil.
func UintPtr[T util.UintType](key string, val *T) Field {
	if val == nil {
		return Nil(key)
	}
	return Uint(key, *val)
}

// Float creates a Field for a float32 value.
func Float[T util.FloatType](key string, val T) Field {
	return Field{
		Key:  key,
		Type: ValueTypeFloat64,
		Num:  math.Float64bits(float64(val)),
	}
}

// FloatPtr creates a Field from a *float32, or returns Nil if pointer is nil.
func FloatPtr[T util.FloatType](key string, val *T) Field {
	if val == nil {
		return Nil(key)
	}
	return Float(key, *val)
}

// String creates a Field for a string value.
func String(key string, val string) Field {
	return Field{
		Key:  key,
		Type: ValueTypeString,
		Num:  uint64(len(val)),       // Store the length of the string
		Any:  unsafe.StringData(val), // Store the pointer to string data
	}
}

// StringPtr creates a Field from a *string, or returns Nil if pointer is nil.
func StringPtr(key string, val *string) Field {
	if val == nil {
		return Nil(key)
	}
	return String(key, *val)
}

// Reflect wraps any value into a Field using reflection.
func Reflect(key string, val interface{}) Field {
	return Field{
		Key:  key,
		Type: ValueTypeReflect,
		Any:  val,
	}
}

type bools []bool

// EncodeArray encodes a slice of bools using the Encoder interface.
func (arr bools) EncodeArray(enc Encoder) {
	for _, v := range arr {
		enc.AppendBool(v)
	}
}

// Bools creates a Field with a slice of booleans.
func Bools(key string, val []bool) Field {
	return Array(key, bools(val))
}

type sliceOfInt[T util.IntType] []T

// EncodeArray encodes a slice of ints using the Encoder interface.
func (arr sliceOfInt[T]) EncodeArray(enc Encoder) {
	for _, v := range arr {
		enc.AppendInt64(int64(v))
	}
}

// Ints creates a Field with a slice of integers.
func Ints[T util.IntType](key string, val []T) Field {
	return Array(key, sliceOfInt[T](val))
}

type sliceOfUint[T util.UintType] []T

// EncodeArray encodes a slice of uints using the Encoder interface.
func (arr sliceOfUint[T]) EncodeArray(enc Encoder) {
	for _, v := range arr {
		enc.AppendUint64(uint64(v))
	}
}

// Uints creates a Field with a slice of unsigned integers.
func Uints[T util.UintType](key string, val []T) Field {
	return Array(key, sliceOfUint[T](val))
}

type sliceOfFloat[T util.FloatType] []T

// EncodeArray encodes a slice of float32s using the Encoder interface.
func (arr sliceOfFloat[T]) EncodeArray(enc Encoder) {
	for _, v := range arr {
		enc.AppendFloat64(float64(v))
	}
}

// Floats creates a Field with a slice of float32 values.
func Floats[T util.FloatType](key string, val []T) Field {
	return Array(key, sliceOfFloat[T](val))
}

type sliceOfString []string

// EncodeArray encodes a slice of strings using the Encoder interface.
func (arr sliceOfString) EncodeArray(enc Encoder) {
	for _, v := range arr {
		enc.AppendString(v)
	}
}

// Strings creates a Field with a slice of strings.
func Strings(key string, val []string) Field {
	return Array(key, sliceOfString(val))
}

// ArrayValue is an interface for types that can be encoded as array.
type ArrayValue interface {
	EncodeArray(enc Encoder)
}

// Array creates a Field with array type, using the ArrayValue interface.
func Array(key string, val ArrayValue) Field {
	return Field{
		Key:  key,
		Type: ValueTypeArray,
		Any:  val,
	}
}

// Object creates a Field containing a variadic slice of Fields, treated as a nested object.
func Object(key string, fields ...Field) Field {
	return Field{
		Key:  key,
		Type: ValueTypeObject,
		Any:  fields,
	}
}

// Any creates a Field from a value of any type by inspecting its dynamic type.
// It dispatches to the appropriate typed constructor based on the actual value.
// If the type is not explicitly handled, it falls back to using Reflect.
func Any(key string, value interface{}) Field {
	switch val := value.(type) {
	case nil:
		return Nil(key)

	case bool:
		return Bool(key, val)
	case *bool:
		return BoolPtr(key, val)
	case []bool:
		return Bools(key, val)

	case int:
		return Int(key, val)
	case *int:
		return IntPtr(key, val)
	case []int:
		return Ints(key, val)

	case int8:
		return Int(key, val)
	case *int8:
		return IntPtr(key, val)
	case []int8:
		return Ints(key, val)

	case int16:
		return Int(key, val)
	case *int16:
		return IntPtr(key, val)
	case []int16:
		return Ints(key, val)

	case int32:
		return Int(key, val)
	case *int32:
		return IntPtr(key, val)
	case []int32:
		return Ints(key, val)

	case int64:
		return Int(key, val)
	case *int64:
		return IntPtr(key, val)
	case []int64:
		return Ints(key, val)

	case uint:
		return Uint(key, val)
	case *uint:
		return UintPtr(key, val)
	case []uint:
		return Uints(key, val)

	case uint8:
		return Uint(key, val)
	case *uint8:
		return UintPtr(key, val)
	case []uint8:
		return Uints(key, val)

	case uint16:
		return Uint(key, val)
	case *uint16:
		return UintPtr(key, val)
	case []uint16:
		return Uints(key, val)

	case uint32:
		return Uint(key, val)
	case *uint32:
		return UintPtr(key, val)
	case []uint32:
		return Uints(key, val)

	case uint64:
		return Uint(key, val)
	case *uint64:
		return UintPtr(key, val)
	case []uint64:
		return Uints(key, val)

	case float32:
		return Float(key, val)
	case *float32:
		return FloatPtr(key, val)
	case []float32:
		return Floats(key, val)

	case float64:
		return Float(key, val)
	case *float64:
		return FloatPtr(key, val)
	case []float64:
		return Floats(key, val)

	case string:
		return String(key, val)
	case *string:
		return StringPtr(key, val)
	case []string:
		return Strings(key, val)

	default:
		return Reflect(key, val)
	}
}

// Encode encodes the Field into the Encoder based on its type.
func (f *Field) Encode(enc Encoder) {
	enc.AppendKey(f.Key)
	switch f.Type {
	case ValueTypeBool:
		enc.AppendBool(f.Num != 0)
	case ValueTypeInt64:
		enc.AppendInt64(int64(f.Num))
	case ValueTypeUint64:
		enc.AppendUint64(f.Num)
	case ValueTypeFloat64:
		enc.AppendFloat64(math.Float64frombits(f.Num))
	case ValueTypeString:
		enc.AppendString(unsafe.String(f.Any.(*byte), f.Num))
	case ValueTypeReflect:
		enc.AppendReflect(f.Any)
	case ValueTypeArray:
		enc.AppendArrayBegin()
		f.Any.(ArrayValue).EncodeArray(enc)
		enc.AppendArrayEnd()
	case ValueTypeObject:
		enc.AppendObjectBegin()
		WriteFields(enc, f.Any.([]Field))
		enc.AppendObjectEnd()
	default: // for linter
	}
}

// WriteFields writes a slice of Field objects to the encoder.
func WriteFields(enc Encoder, fields []Field) {
	for _, f := range fields {
		f.Encode(enc)
	}
}
