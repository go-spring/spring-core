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

package value_struct

import (
	"math"
	"unsafe"

	"fields/encoder"
)

type (
	StringDataPtr *byte // used in Value.Any when the Value is a string
)

type ValueType int

const (
	ValueTypeBool = ValueType(iota)
	ValueTypeInt64
	ValueTypeUint64
	ValueTypeFloat64
	ValueTypeString
	ValueTypeReflect
	ValueTypeBools
	ValueTypeInts
	ValueTypeInt8s
	ValueTypeInt16s
	ValueTypeInt32s
	ValueTypeInt64s
	ValueTypeUints
	ValueTypeUint8s
	ValueTypeUint16s
	ValueTypeUint32s
	ValueTypeUint64s
	ValueTypeFloat32s
	ValueTypeFloat64s
	ValueTypeStrings
)

type Value struct {
	Type ValueType
	Num  uint64
	Any  any
}

func (v Value) Encode(enc encoder.Encoder) {
	switch v.Type {
	case ValueTypeBool:
		if v.Num == 0 {
			enc.AppendBool(false)
		} else {
			enc.AppendBool(true)
		}
	case ValueTypeInt64:
		enc.AppendInt64(int64(v.Num))
	case ValueTypeUint64:
		enc.AppendUint64(v.Num)
	case ValueTypeFloat64:
		enc.AppendFloat64(math.Float64frombits(v.Num))
	case ValueTypeString:
		enc.AppendString(unsafe.String(v.Any.(StringDataPtr), v.Num))
	case ValueTypeReflect:
		enc.AppendReflect(v.Any)
	case ValueTypeBools:
		enc.AppendArrayBegin()
		for _, val := range v.Any.([]bool) {
			enc.AppendBool(val)
		}
		enc.AppendArrayEnd()
	case ValueTypeInts:
		enc.AppendArrayBegin()
		for _, val := range v.Any.([]int) {
			enc.AppendInt64(int64(val))
		}
		enc.AppendArrayEnd()
	case ValueTypeInt8s:
		enc.AppendArrayBegin()
		for _, val := range v.Any.([]int8) {
			enc.AppendInt64(int64(val))
		}
		enc.AppendArrayEnd()
	case ValueTypeInt16s:
		enc.AppendArrayBegin()
		for _, val := range v.Any.([]int16) {
			enc.AppendInt64(int64(val))
		}
		enc.AppendArrayEnd()
	case ValueTypeInt32s:
		enc.AppendArrayBegin()
		for _, val := range v.Any.([]int32) {
			enc.AppendInt64(int64(val))
		}
		enc.AppendArrayEnd()
	case ValueTypeInt64s:
		enc.AppendArrayBegin()
		for _, val := range v.Any.([]int64) {
			enc.AppendInt64(val)
		}
		enc.AppendArrayEnd()
	case ValueTypeUints:
		enc.AppendArrayBegin()
		for _, val := range v.Any.([]uint) {
			enc.AppendUint64(uint64(val))
		}
		enc.AppendArrayEnd()
	case ValueTypeUint8s:
		enc.AppendArrayBegin()
		for _, val := range v.Any.([]uint8) {
			enc.AppendUint64(uint64(val))
		}
		enc.AppendArrayEnd()
	case ValueTypeUint16s:
		enc.AppendArrayBegin()
		for _, val := range v.Any.([]uint16) {
			enc.AppendUint64(uint64(val))
		}
		enc.AppendArrayEnd()
	case ValueTypeUint32s:
		enc.AppendArrayBegin()
		for _, val := range v.Any.([]uint32) {
			enc.AppendUint64(uint64(val))
		}
		enc.AppendArrayEnd()
	case ValueTypeUint64s:
		enc.AppendArrayBegin()
		for _, val := range v.Any.([]uint64) {
			enc.AppendUint64(val)
		}
		enc.AppendArrayEnd()
	case ValueTypeFloat32s:
		enc.AppendArrayBegin()
		for _, val := range v.Any.([]float32) {
			enc.AppendFloat64(float64(val))
		}
		enc.AppendArrayEnd()
	case ValueTypeFloat64s:
		enc.AppendArrayBegin()
		for _, val := range v.Any.([]float64) {
			enc.AppendFloat64(val)
		}
		enc.AppendArrayEnd()
	case ValueTypeStrings:
		enc.AppendArrayBegin()
		for _, val := range v.Any.([]string) {
			enc.AppendString(val)
		}
		enc.AppendArrayEnd()

	default: // for linter
	}
}

func BoolValue(v bool) Value {
	if v {
		return Value{
			Type: ValueTypeBool,
			Num:  1,
		}
	} else {
		return Value{
			Type: ValueTypeBool,
			Num:  0,
		}
	}
}

func Int64Value(v int64) Value {
	return Value{
		Type: ValueTypeInt64,
		Num:  uint64(v),
	}
}

func Uint64Value(v uint64) Value {
	return Value{
		Type: ValueTypeUint64,
		Num:  v,
	}
}

func Float64Value(v float64) Value {
	return Value{
		Type: ValueTypeFloat64,
		Num:  math.Float64bits(v),
	}
}

func StringValue(v string) Value {
	return Value{
		Type: ValueTypeString,
		Num:  uint64(len(v)),
		Any:  StringDataPtr(unsafe.StringData(v)),
	}
}

func ReflectValue(v any) Value {
	return Value{
		Type: ValueTypeReflect,
		Any:  v,
	}
}

func BoolsValue(v []bool) Value {
	return Value{
		Type: ValueTypeBools,
		Any:  v,
	}
}

func IntsValue(v []int) Value {
	return Value{
		Type: ValueTypeInts,
		Any:  v,
	}
}

func Int8sValue(v []int8) Value {
	return Value{
		Type: ValueTypeInt8s,
		Any:  v,
	}
}

func Int16sValue(v []int16) Value {
	return Value{
		Type: ValueTypeInt16s,
		Any:  v,
	}
}

func Int32sValue(v []int32) Value {
	return Value{
		Type: ValueTypeInt32s,
		Any:  v,
	}
}

func Int64sValue(v []int64) Value {
	return Value{
		Type: ValueTypeInt64s,
		Any:  v,
	}
}

func UintsValue(v []uint) Value {
	return Value{
		Type: ValueTypeUints,
		Any:  v,
	}
}

func Uint8sValue(v []uint8) Value {
	return Value{
		Type: ValueTypeUint8s,
		Any:  v,
	}
}

func Uint16sValue(v []uint16) Value {
	return Value{
		Type: ValueTypeUint16s,
		Any:  v,
	}
}

func Uint32sValue(v []uint32) Value {
	return Value{
		Type: ValueTypeUint32s,
		Any:  v,
	}
}

func Uint64sValue(v []uint64) Value {
	return Value{
		Type: ValueTypeUint64s,
		Any:  v,
	}
}

func Float32sValue(v []float32) Value {
	return Value{
		Type: ValueTypeFloat32s,
		Any:  v,
	}
}

func Float64sValue(v []float64) Value {
	return Value{
		Type: ValueTypeFloat64s,
		Any:  v,
	}
}

func StringsValue(v []string) Value {
	return Value{
		Type: ValueTypeStrings,
		Any:  v,
	}
}

type ArrayValue []Value

func (v ArrayValue) Encode(enc encoder.Encoder) {
	enc.AppendArrayBegin()
	for _, val := range v {
		val.Encode(enc)
	}
	enc.AppendArrayEnd()
}

type ObjectValue []Field

func (v ObjectValue) Encode(enc encoder.Encoder) {
	enc.AppendObjectBegin()
	WriteFields(enc, v)
	enc.AppendObjectEnd()
}

func WriteFields(enc encoder.Encoder, fields []Field) {
	for _, f := range fields {
		enc.AppendKey(f.Key)
		f.Val.Encode(enc)
	}
}
