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

// errorType is the [reflect.Type] of the error interface.
var errorType = reflect.TypeOf((*error)(nil)).Elem()

// IsFuncType returns true if the provided type t is a function type.
func IsFuncType(t reflect.Type) bool {
	return t.Kind() == reflect.Func
}

// IsErrorType returns true if the provided type t is an error type,
// either directly (error) or via an implementation (i.e., implements the error interface).
func IsErrorType(t reflect.Type) bool {
	return t == errorType || t.Implements(errorType)
}

// ReturnNothing returns true if the provided function type t has no return values.
func ReturnNothing(t reflect.Type) bool {
	return t.NumOut() == 0
}

// ReturnOnlyError returns true if the provided function type t returns only one value,
// and that value is an error.
func ReturnOnlyError(t reflect.Type) bool {
	return t.NumOut() == 1 && IsErrorType(t.Out(0))
}

// IsConstructor returns true if the provided function type t is a constructor.
// A constructor is defined as a function that returns one or two values.
// If it returns two values, the second value must be an error.
func IsConstructor(t reflect.Type) bool {
	if !IsFuncType(t) {
		return false
	}
	switch t.NumOut() {
	case 1:
		return true
	case 2:
		return IsErrorType(t.Out(1))
	default:
		return false
	}
}

// IsPrimitiveValueType returns true if the provided type t is a primitive value type,
// such as int, uint, float, bool, or string.
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

// IsPropBindingTarget returns true if the provided type t is a valid target for property binding.
// This includes primitive value types or composite types (such as array, slice, map, or struct)
// where the elements are primitive value types.
func IsPropBindingTarget(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		t = t.Elem() // for collection types, check the element type
	default:
		// do nothing
	}
	return IsPrimitiveValueType(t) || t.Kind() == reflect.Struct
}

// IsBeanType returns true if the provided type t is considered a "bean" type.
// A "bean" type includes a channel, function, interface, or a pointer to a struct.
func IsBeanType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface:
		return true
	case reflect.Ptr:
		return t.Elem().Kind() == reflect.Struct
	default:
		return false
	}
}

// IsBeanInjectionTarget returns true if the provided type t is a valid target for bean injection.
// This includes maps, slices, arrays, or any other bean type (including pointers to structs).
func IsBeanInjectionTarget(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		t = t.Elem() // for collection types, check the element type
	default:
		// do nothing
	}
	return IsBeanType(t)
}
