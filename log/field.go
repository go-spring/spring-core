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
)

// Field represents a log entry field with a key and a corresponding value.
type Field struct {
	Key string // The name of the log field.
	Val Value  // The value of the log field.
}

// Msg creates a Field with the key "msg" and a string message value.
func Msg(msg string) Field {
	return String("msg", msg)
}

// Msgf formats the message using fmt.Sprintf and returns a Field with key "msg".
func Msgf(format string, args ...any) Field {
	return String("msg", fmt.Sprintf(format, args...))
}

// Reflect wraps any value into a Field using reflection.
func Reflect(key string, val interface{}) Field {
	return Field{Key: key, Val: ReflectValue{Val: val}}
}

// Nil creates a Field for a given key with a nil value.
func Nil(key string) Field {
	return Reflect(key, nil)
}

// Bool creates a boolean Field.
func Bool(key string, val bool) Field {
	return Field{Key: key, Val: BoolValue(val)}
}

// BoolPtr creates a boolean Field from a pointer; returns Nil if pointer is nil.
func BoolPtr(key string, val *bool) Field {
	if val == nil {
		return Nil(key)
	}
	return Bool(key, *val)
}

// Int creates an integer Field.
func Int(key string, val int) Field {
	return Field{Key: key, Val: Int64Value(val)}
}

// IntPtr creates an integer Field from a pointer; returns Nil if pointer is nil.
func IntPtr(key string, val *int) Field {
	if val == nil {
		return Nil(key)
	}
	return Int(key, *val)
}

// Int8 creates an int8 Field.
func Int8(key string, val int8) Field {
	return Field{Key: key, Val: Int64Value(val)}
}

// Int8Ptr creates an int8 Field from a pointer; returns Nil if pointer is nil.
func Int8Ptr(key string, val *int8) Field {
	if val == nil {
		return Nil(key)
	}
	return Int8(key, *val)
}

// Int16 creates an int16 Field.
func Int16(key string, val int16) Field {
	return Field{Key: key, Val: Int64Value(val)}
}

// Int16Ptr creates an int16 Field from a pointer; returns Nil if pointer is nil.
func Int16Ptr(key string, val *int16) Field {
	if val == nil {
		return Nil(key)
	}
	return Int16(key, *val)
}

// Int32 creates an int32 Field.
func Int32(key string, val int32) Field {
	return Field{Key: key, Val: Int64Value(val)}
}

// Int32Ptr creates an int32 Field from a pointer; returns Nil if pointer is nil.
func Int32Ptr(key string, val *int32) Field {
	if val == nil {
		return Nil(key)
	}
	return Int32(key, *val)
}

// Int64 creates an int64 Field.
func Int64(key string, val int64) Field {
	return Field{Key: key, Val: Int64Value(val)}
}

// Int64Ptr creates an int64 Field from a pointer; returns Nil if pointer is nil.
func Int64Ptr(key string, val *int64) Field {
	if val == nil {
		return Nil(key)
	}
	return Int64(key, *val)
}

// Uint creates a uint Field.
func Uint(key string, val uint) Field {
	return Field{Key: key, Val: Uint64Value(val)}
}

// UintPtr creates a uint Field from a pointer; returns Nil if pointer is nil.
func UintPtr(key string, val *uint) Field {
	if val == nil {
		return Nil(key)
	}
	return Uint(key, *val)
}

// Uint8 creates a uint8 Field.
func Uint8(key string, val uint8) Field {
	return Field{Key: key, Val: Uint64Value(val)}
}

// Uint8Ptr creates a uint8 Field from a pointer; returns Nil if pointer is nil.
func Uint8Ptr(key string, val *uint8) Field {
	if val == nil {
		return Nil(key)
	}
	return Uint8(key, *val)
}

// Uint16 creates a uint16 Field.
func Uint16(key string, val uint16) Field {
	return Field{Key: key, Val: Uint64Value(val)}
}

// Uint16Ptr creates a uint16 Field from a pointer; returns Nil if pointer is nil.
func Uint16Ptr(key string, val *uint16) Field {
	if val == nil {
		return Nil(key)
	}
	return Uint16(key, *val)
}

// Uint32 creates a uint32 Field.
func Uint32(key string, val uint32) Field {
	return Field{Key: key, Val: Uint64Value(val)}
}

// Uint32Ptr creates a uint32 Field from a pointer; returns Nil if pointer is nil.
func Uint32Ptr(key string, val *uint32) Field {
	if val == nil {
		return Nil(key)
	}
	return Uint32(key, *val)
}

// Uint64 creates a uint64 Field.
func Uint64(key string, val uint64) Field {
	return Field{Key: key, Val: Uint64Value(val)}
}

// Uint64Ptr creates a uint64 Field from a pointer; returns Nil if pointer is nil.
func Uint64Ptr(key string, val *uint64) Field {
	if val == nil {
		return Nil(key)
	}
	return Uint64(key, *val)
}

// Float32 creates a float32 Field.
func Float32(key string, val float32) Field {
	return Field{Key: key, Val: Float64Value(val)}
}

// Float32Ptr creates a float32 Field from a pointer; returns Nil if pointer is nil.
func Float32Ptr(key string, val *float32) Field {
	if val == nil {
		return Nil(key)
	}
	return Float32(key, *val)
}

// Float64 creates a float64 Field.
func Float64(key string, val float64) Field {
	return Field{Key: key, Val: Float64Value(val)}
}

// Float64Ptr creates a float64 Field from a pointer; returns Nil if pointer is nil.
func Float64Ptr(key string, val *float64) Field {
	if val == nil {
		return Nil(key)
	}
	return Float64(key, *val)
}

// String creates a string Field.
func String(key string, val string) Field {
	return Field{Key: key, Val: StringValue(val)}
}

// StringPtr creates a string Field from a pointer; returns Nil if pointer is nil.
func StringPtr(key string, val *string) Field {
	if val == nil {
		return Nil(key)
	}
	return String(key, *val)
}

// Array creates a Field containing a variadic slice of Values, wrapped as an array.
func Array(key string, val ...Value) Field {
	return Field{Key: key, Val: ArrayValue(val)}
}

// Object creates a Field containing a variadic slice of Fields, treated as a nested object.
func Object(key string, fields ...Field) Field {
	return Field{Key: key, Val: ObjectValue(fields)}
}

// Bools creates a Field containing a slice of boolean values.
func Bools(key string, val []bool) Field {
	return Field{Key: key, Val: BoolsValue(val)}
}

// Ints creates a Field containing a slice of int values.
func Ints(key string, val []int) Field {
	return Field{Key: key, Val: IntsValue(val)}
}

// Int8s creates a Field containing a slice of int8 values.
func Int8s(key string, val []int8) Field {
	return Field{Key: key, Val: Int8sValue(val)}
}

// Int16s creates a Field containing a slice of int16 values.
func Int16s(key string, val []int16) Field {
	return Field{Key: key, Val: Int16sValue(val)}
}

// Int32s creates a Field containing a slice of int32 values.
func Int32s(key string, val []int32) Field {
	return Field{Key: key, Val: Int32sValue(val)}
}

// Int64s creates a Field containing a slice of int64 values.
func Int64s(key string, val []int64) Field {
	return Field{Key: key, Val: Int64sValue(val)}
}

// Uints creates a Field containing a slice of uint values.
func Uints(key string, val []uint) Field {
	return Field{Key: key, Val: UintsValue(val)}
}

// Uint8s creates a Field containing a slice of uint8 values.
func Uint8s(key string, val []uint8) Field {
	return Field{Key: key, Val: Uint8sValue(val)}
}

// Uint16s creates a Field containing a slice of uint16 values.
func Uint16s(key string, val []uint16) Field {
	return Field{Key: key, Val: Uint16sValue(val)}
}

// Uint32s creates a Field containing a slice of uint32 values.
func Uint32s(key string, val []uint32) Field {
	return Field{Key: key, Val: Uint32sValue(val)}
}

// Uint64s creates a Field containing a slice of uint64 values.
func Uint64s(key string, val []uint64) Field {
	return Field{Key: key, Val: Uint64sValue(val)}
}

// Float32s creates a Field containing a slice of float32 values.
func Float32s(key string, val []float32) Field {
	return Field{Key: key, Val: Float32sValue(val)}
}

// Float64s creates a Field containing a slice of float64 values.
func Float64s(key string, val []float64) Field {
	return Field{Key: key, Val: Float64sValue(val)}
}

// Strings creates a Field containing a slice of string values.
func Strings(key string, val []string) Field {
	return Field{Key: key, Val: StringsValue(val)}
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
		return Int8(key, val)
	case *int8:
		return Int8Ptr(key, val)
	case []int8:
		return Int8s(key, val)

	case int16:
		return Int16(key, val)
	case *int16:
		return Int16Ptr(key, val)
	case []int16:
		return Int16s(key, val)

	case int32:
		return Int32(key, val)
	case *int32:
		return Int32Ptr(key, val)
	case []int32:
		return Int32s(key, val)

	case int64:
		return Int64(key, val)
	case *int64:
		return Int64Ptr(key, val)
	case []int64:
		return Int64s(key, val)

	case uint:
		return Uint(key, val)
	case *uint:
		return UintPtr(key, val)
	case []uint:
		return Uints(key, val)

	case uint8:
		return Uint8(key, val)
	case *uint8:
		return Uint8Ptr(key, val)
	case []uint8:
		return Uint8s(key, val)

	case uint16:
		return Uint16(key, val)
	case *uint16:
		return Uint16Ptr(key, val)
	case []uint16:
		return Uint16s(key, val)

	case uint32:
		return Uint32(key, val)
	case *uint32:
		return Uint32Ptr(key, val)
	case []uint32:
		return Uint32s(key, val)

	case uint64:
		return Uint64(key, val)
	case *uint64:
		return Uint64Ptr(key, val)
	case []uint64:
		return Uint64s(key, val)

	case float32:
		return Float32(key, val)
	case *float32:
		return Float32Ptr(key, val)
	case []float32:
		return Float32s(key, val)

	case float64:
		return Float64(key, val)
	case *float64:
		return Float64Ptr(key, val)
	case []float64:
		return Float64s(key, val)

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
