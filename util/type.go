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
	"strings"
)

// errorType the reflection type of error.
var errorType = reflect.TypeOf((*error)(nil)).Elem()

// TypeName returns a fully qualified name consisting of package path and type name.
func TypeName(t reflect.Type) string {

	for {
		if k := t.Kind(); k == reflect.Ptr || k == reflect.Slice {
			t = t.Elem()
		} else {
			break
		}
	}

	if pkgPath := t.PkgPath(); pkgPath != "" {
		pkgPath = strings.TrimSuffix(pkgPath, "_test")
		return pkgPath + "/" + t.String()
	}
	return t.String() // the path of built-in type is empty
}

// IsFuncType returns whether `t` is func type.
func IsFuncType(t reflect.Type) bool {
	return t.Kind() == reflect.Func
}

// IsErrorType returns whether `t` is error type.
func IsErrorType(t reflect.Type) bool {
	return t == errorType || t.Implements(errorType)
}

// ReturnNothing returns whether the function has no return value.
func ReturnNothing(t reflect.Type) bool {
	return t.NumOut() == 0
}

// ReturnOnlyError returns whether the function returns only error value.
func ReturnOnlyError(t reflect.Type) bool {
	return t.NumOut() == 1 && IsErrorType(t.Out(0))
}

// IsStructPtr returns whether it is the pointer type of structure.
func IsStructPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

// IsConstructor returns whether `t` is a constructor type. What is a constructor?
// It should be a function first, has any number of inputs and supports the option
// pattern input, has one or two outputs and the second output should be an error.
func IsConstructor(t reflect.Type) bool {
	returnError := t.NumOut() == 2 && IsErrorType(t.Out(1))
	return IsFuncType(t) && (t.NumOut() == 1 || returnError)
}

// HasReceiver returns whether the function has a receiver.
func HasReceiver(t reflect.Type, receiver reflect.Value) bool {
	if t.NumIn() < 1 {
		return false
	}
	t0 := t.In(0)
	if t0.Kind() != reflect.Interface {
		return t0 == receiver.Type()
	}
	return receiver.Type().Implements(t0)
}

// IsPrimitiveValueType returns whether `t` is the primitive value type which only is
// int, unit, float, bool, string and complex.
func IsPrimitiveValueType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Complex64, reflect.Complex128:
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

// IsValueType returns whether the input type is the primitive value type and their
// composite type including array, slice, map and struct, such as []int, [3]string,
// []string, map[int]int, map[string]string, etc.
func IsValueType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		t = t.Elem()
	default:
		// do nothing
	}
	return IsPrimitiveValueType(t) || t.Kind() == reflect.Struct
}

// IsBeanType returns whether `t` is a bean type.
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

// IsBeanInjectionTarget returns whether `t` is a bean injection target.
func IsBeanInjectionTarget(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		t = t.Elem()
	default:
		// do nothing
	}
	return IsBeanType(t)
}
