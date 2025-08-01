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

package util_test

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/go-spring/gs-assert/assert"
	"github.com/go-spring/spring-core/util"
)

func TestIsErrorType(t *testing.T) {
	err := fmt.Errorf("error")
	assert.That(t, util.IsErrorType(reflect.TypeOf(err))).True()
	err = os.ErrClosed
	assert.That(t, util.IsErrorType(reflect.TypeOf(err))).True()
	assert.That(t, util.IsErrorType(reflect.TypeFor[int]())).False()
}

func TestReturnNothing(t *testing.T) {
	assert.That(t, util.ReturnNothing(reflect.TypeOf(func() {}))).True()
	assert.That(t, util.ReturnNothing(reflect.TypeOf(func(key string) {}))).True()
	assert.That(t, util.ReturnNothing(reflect.TypeOf(func() string { return "" }))).False()
}

func TestReturnOnlyError(t *testing.T) {
	assert.That(t, util.ReturnOnlyError(reflect.TypeOf(func() error { return nil }))).True()
	assert.That(t, util.ReturnOnlyError(reflect.TypeOf(func(string) error { return nil }))).True()
	assert.That(t, util.ReturnOnlyError(reflect.TypeOf(func() (string, error) { return "", nil }))).False()
}

func TestIsConstructor(t *testing.T) {
	assert.That(t, util.IsConstructor(reflect.TypeFor[int]())).False()
	assert.That(t, util.IsConstructor(reflect.TypeOf(func() {}))).False()
	assert.That(t, util.IsConstructor(reflect.TypeOf(func() string { return "" }))).True()
	assert.That(t, util.IsConstructor(reflect.TypeOf(func() *string { return nil }))).True()
	assert.That(t, util.IsConstructor(reflect.TypeOf(func() *receiver { return nil }))).True()
	assert.That(t, util.IsConstructor(reflect.TypeOf(func() (*receiver, error) { return nil, nil }))).True()
	assert.That(t, util.IsConstructor(reflect.TypeOf(func() (bool, *receiver, error) { return false, nil, nil }))).False()
}

func TestIsPropBindingTarget(t *testing.T) {
	data := []struct {
		i any
		v bool
	}{
		{true, true},                      // Bool
		{int(1), true},                    // Int
		{int8(1), true},                   // Int8
		{int16(1), true},                  // Int16
		{int32(1), true},                  // Int32
		{int64(1), true},                  // Int64
		{uint(1), true},                   // Uint
		{uint8(1), true},                  // Uint8
		{uint16(1), true},                 // Uint16
		{uint32(1), true},                 // Uint32
		{uint64(1), true},                 // Uint64
		{uintptr(0), false},               // Uintptr
		{float32(1), true},                // Float32
		{float64(1), true},                // Float64
		{complex64(1), false},             // Complex64
		{complex128(1), false},            // Complex128
		{[1]int{0}, true},                 // Array
		{make(chan struct{}), false},      // Chan
		{func() {}, false},                // Func
		{reflect.TypeFor[error](), false}, // Interface
		{make(map[int]int), true},         // Map
		{make(map[string]*int), false},    //
		{new(int), false},                 // Ptr
		{new(struct{}), false},            //
		{[]int{0}, true},                  // Slice
		{[]*int{}, false},                 //
		{"this is a string", true},        // String
		{struct{}{}, true},                // Struct
		{unsafe.Pointer(new(int)), false}, // UnsafePointer
	}
	for _, d := range data {
		var typ reflect.Type
		switch i := d.i.(type) {
		case reflect.Type:
			typ = i
		default:
			typ = reflect.TypeOf(i)
		}
		if r := util.IsPropBindingTarget(typ); d.v != r {
			t.Errorf("%v expect %v but %v", typ, d.v, r)
		}
	}
}

func TestIsBeanType(t *testing.T) {
	data := []struct {
		i any
		v bool
	}{
		{true, false},                     // Bool
		{int(1), false},                   // Int
		{int8(1), false},                  // Int8
		{int16(1), false},                 // Int16
		{int32(1), false},                 // Int32
		{int64(1), false},                 // Int64
		{uint(1), false},                  // Uint
		{uint8(1), false},                 // Uint8
		{uint16(1), false},                // Uint16
		{uint32(1), false},                // Uint32
		{uint64(1), false},                // Uint64
		{uintptr(0), false},               // Uintptr
		{float32(1), false},               // Float32
		{float64(1), false},               // Float64
		{complex64(1), false},             // Complex64
		{complex128(1), false},            // Complex128
		{[1]int{0}, false},                // Array
		{make(chan struct{}), true},       // Chan
		{func() {}, true},                 // Func
		{reflect.TypeFor[error](), true},  // Interface
		{make(map[int]int), false},        // Map
		{make(map[string]*int), false},    //
		{new(int), false},                 //
		{new(struct{}), true},             //
		{[]int{0}, false},                 // Slice
		{[]*int{}, false},                 //
		{"this is a string", false},       // String
		{struct{}{}, false},               // Struct
		{unsafe.Pointer(new(int)), false}, // UnsafePointer
	}
	for _, d := range data {
		var typ reflect.Type
		switch i := d.i.(type) {
		case reflect.Type:
			typ = i
		default:
			typ = reflect.TypeOf(i)
		}
		if r := util.IsBeanType(typ); d.v != r {
			t.Errorf("%v expect %v but %v", typ, d.v, r)
		}
	}
}

func TestIsBeanInjectionTarget(t *testing.T) {
	assert.That(t, util.IsBeanInjectionTarget(reflect.TypeOf("abc"))).False()
	assert.That(t, util.IsBeanInjectionTarget(reflect.TypeOf(new(string)))).False()
	assert.That(t, util.IsBeanInjectionTarget(reflect.TypeOf(errors.New("abc")))).True()
	assert.That(t, util.IsBeanInjectionTarget(reflect.TypeOf([]string{}))).False()
	assert.That(t, util.IsBeanInjectionTarget(reflect.TypeOf([]*string{}))).False()
	assert.That(t, util.IsBeanInjectionTarget(reflect.TypeOf([]fmt.Stringer{}))).True()
	assert.That(t, util.IsBeanInjectionTarget(reflect.TypeOf(map[string]string{}))).False()
	assert.That(t, util.IsBeanInjectionTarget(reflect.TypeOf(map[string]*string{}))).False()
	assert.That(t, util.IsBeanInjectionTarget(reflect.TypeOf(map[string]fmt.Stringer{}))).True()
}
