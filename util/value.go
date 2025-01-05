/*
 * Copyright 2012-2024 the original author or authors.
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

package util

import (
	"reflect"
)

// errorType the reflection type of error.
var errorType = reflect.TypeOf((*error)(nil)).Elem()

// IsFuncType returns whether `t` is func type.
func IsFuncType(t reflect.Type) bool {
	return t.Kind() == reflect.Func
}

// IsErrorType returns whether `t` is error type.
func IsErrorType(t reflect.Type) bool {
	return t == errorType || t.Implements(errorType)
}

// IsConverter returns whether `t` is a converter type.
func IsConverter(t reflect.Type) bool {
	return IsFuncType(t) &&
		t.NumIn() == 1 &&
		t.In(0).Kind() == reflect.String &&
		t.NumOut() == 2 &&
		(IsValueType(t.Out(0)) || IsFuncType(t.Out(0))) && IsErrorType(t.Out(1))
}

// IsValueType returns whether the input type is the primitive value type and their
// composite type including array, slice, map and struct, such as []int, [3]string,
// []string, map[int]int, map[string]string, etc.
func IsValueType(t reflect.Type) bool {
	fn := func(t reflect.Type) bool {
		return IsPrimitiveValueType(t) || t.Kind() == reflect.Struct
	}
	switch t.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		return fn(t.Elem())
	default:
		return fn(t)
	}
}

// IsPrimitiveValueType returns whether `t` is the primitive value type which only is
// int, unit, float, bool, string and complex.
func IsPrimitiveValueType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	case reflect.String:
		return true
	case reflect.Bool:
		return true
	default:
		return false
	}
}
