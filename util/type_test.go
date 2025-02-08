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

package util_test

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/go-spring/spring-core/util"
	"github.com/go-spring/spring-core/util/assert"
	"github.com/go-spring/spring-core/util/testdata"
	pkg1 "github.com/go-spring/spring-core/util/testdata/pkg/bar"
	pkg2 "github.com/go-spring/spring-core/util/testdata/pkg/foo"
)

type SamePkg struct{}

func (p *SamePkg) Package() {
	fmt.Println("github.com/go-spring/spring-core/util/util_test.SamePkg")
}

func TestPkgPath(t *testing.T) {
	// the name and package path of built-in type are empty.

	data := []struct {
		typ  reflect.Type
		kind reflect.Kind
		name string
		pkg  string
	}{
		{
			reflect.TypeOf(false),
			reflect.Bool,
			"bool",
			"",
		},
		{
			reflect.TypeOf(new(bool)),
			reflect.Ptr,
			"",
			"",
		},
		{
			reflect.TypeOf(make([]bool, 0)),
			reflect.Slice,
			"",
			"",
		},
		{
			reflect.TypeOf(int(3)),
			reflect.Int,
			"int",
			"",
		},
		{
			reflect.TypeOf(new(int)),
			reflect.Ptr,
			"",
			"",
		},
		{
			reflect.TypeOf(make([]int, 0)),
			reflect.Slice,
			"",
			"",
		},
		{
			reflect.TypeOf(uint(3)),
			reflect.Uint,
			"uint",
			"",
		},
		{
			reflect.TypeOf(new(uint)),
			reflect.Ptr,
			"",
			"",
		},
		{
			reflect.TypeOf(make([]uint, 0)),
			reflect.Slice,
			"",
			"",
		},
		{
			reflect.TypeOf(float32(3)),
			reflect.Float32,
			"float32",
			"",
		},
		{
			reflect.TypeOf(new(float32)),
			reflect.Ptr,
			"",
			"",
		},
		{
			reflect.TypeOf(make([]float32, 0)),
			reflect.Slice,
			"",
			"",
		},
		{
			reflect.TypeOf(complex64(3)),
			reflect.Complex64,
			"complex64",
			"",
		},
		{
			reflect.TypeOf(new(complex64)),
			reflect.Ptr,
			"",
			"",
		},
		{
			reflect.TypeOf(make([]complex64, 0)),
			reflect.Slice,
			"",
			"",
		},
		{
			reflect.TypeOf("3"),
			reflect.String,
			"string",
			"",
		},
		{
			reflect.TypeOf(new(string)),
			reflect.Ptr,
			"",
			"",
		},
		{
			reflect.TypeOf(make([]string, 0)),
			reflect.Slice,
			"",
			"",
		},
		{
			reflect.TypeOf(map[int]int{}),
			reflect.Map,
			"",
			"",
		},
		{
			reflect.TypeOf(new(map[int]int)),
			reflect.Ptr,
			"",
			"",
		},
		{
			reflect.TypeOf(make([]map[int]int, 0)),
			reflect.Slice,
			"",
			"",
		},
		{
			reflect.TypeOf(pkg1.SamePkg{}),
			reflect.Struct,
			"SamePkg",
			"github.com/go-spring/spring-core/util/testdata/pkg/bar",
		},
		{
			reflect.TypeOf(new(pkg1.SamePkg)),
			reflect.Ptr,
			"",
			"",
		},
		{
			reflect.TypeOf(make([]pkg1.SamePkg, 0)),
			reflect.Slice,
			"",
			"",
		},
		{
			reflect.TypeOf(make([]*pkg1.SamePkg, 0)),
			reflect.Slice,
			"",
			"",
		},
		{
			reflect.TypeOf(pkg2.SamePkg{}),
			reflect.Struct,
			"SamePkg",
			"github.com/go-spring/spring-core/util/testdata/pkg/foo",
		},
		{
			reflect.TypeOf(new(pkg2.SamePkg)),
			reflect.Ptr,
			"",
			"",
		},
		{
			reflect.TypeOf(make([]pkg2.SamePkg, 0)),
			reflect.Slice,
			"",
			"",
		},
		{
			reflect.TypeOf((*error)(nil)),
			reflect.Ptr,
			"",
			"",
		},
		{
			reflect.TypeOf((*error)(nil)).Elem(),
			reflect.Interface,
			"error",
			"",
		},
		{
			reflect.TypeOf((*io.Reader)(nil)),
			reflect.Ptr,
			"",
			"",
		},
		{
			reflect.TypeOf((*io.Reader)(nil)).Elem(),
			reflect.Interface,
			"Reader",
			"io",
		},
	}

	for _, d := range data {
		assert.Equal(t, d.typ.Kind(), d.kind)
		assert.Equal(t, d.typ.Name(), d.name)
		assert.Equal(t, d.typ.PkgPath(), d.pkg)
	}
}

func TestTypeName(t *testing.T) {

	data := map[reflect.Type]struct {
		typeName string
		baseName string
	}{
		reflect.TypeOf(3):                 {"int", "int"},
		reflect.TypeOf(new(int)):          {"int", "*int"},
		reflect.TypeOf(make([]int, 0)):    {"int", "[]int"},
		reflect.TypeOf(&[]int{3}):         {"int", "*[]int"},
		reflect.TypeOf(make([]*int, 0)):   {"int", "[]*int"},
		reflect.TypeOf(make([][]int, 0)):  {"int", "[][]int"},
		reflect.TypeOf(make(map[int]int)): {"map[int]int", "map[int]int"},

		reflect.TypeOf(int8(3)):             {"int8", "int8"},
		reflect.TypeOf(new(int8)):           {"int8", "*int8"},
		reflect.TypeOf(make([]int8, 0)):     {"int8", "[]int8"},
		reflect.TypeOf(&[]int8{3}):          {"int8", "*[]int8"},
		reflect.TypeOf(make(map[int8]int8)): {"map[int8]int8", "map[int8]int8"},

		reflect.TypeOf(int16(3)):              {"int16", "int16"},
		reflect.TypeOf(new(int16)):            {"int16", "*int16"},
		reflect.TypeOf(make([]int16, 0)):      {"int16", "[]int16"},
		reflect.TypeOf(&[]int16{3}):           {"int16", "*[]int16"},
		reflect.TypeOf(make(map[int16]int16)): {"map[int16]int16", "map[int16]int16"},

		reflect.TypeOf(int32(3)):              {"int32", "int32"},
		reflect.TypeOf(new(int32)):            {"int32", "*int32"},
		reflect.TypeOf(make([]int32, 0)):      {"int32", "[]int32"},
		reflect.TypeOf(&[]int32{3}):           {"int32", "*[]int32"},
		reflect.TypeOf(make(map[int32]int32)): {"map[int32]int32", "map[int32]int32"},

		reflect.TypeOf(int64(3)):              {"int64", "int64"},
		reflect.TypeOf(new(int64)):            {"int64", "*int64"},
		reflect.TypeOf(make([]int64, 0)):      {"int64", "[]int64"},
		reflect.TypeOf(&[]int64{3}):           {"int64", "*[]int64"},
		reflect.TypeOf(make(map[int64]int64)): {"map[int64]int64", "map[int64]int64"},

		reflect.TypeOf(uint(3)):             {"uint", "uint"},
		reflect.TypeOf(new(uint)):           {"uint", "*uint"},
		reflect.TypeOf(make([]uint, 0)):     {"uint", "[]uint"},
		reflect.TypeOf(&[]uint{3}):          {"uint", "*[]uint"},
		reflect.TypeOf(make(map[uint]uint)): {"map[uint]uint", "map[uint]uint"},

		reflect.TypeOf(uint8(3)):              {"uint8", "uint8"},
		reflect.TypeOf(new(uint8)):            {"uint8", "*uint8"},
		reflect.TypeOf(make([]uint8, 0)):      {"uint8", "[]uint8"},
		reflect.TypeOf(&[]uint8{3}):           {"uint8", "*[]uint8"},
		reflect.TypeOf(make(map[uint8]uint8)): {"map[uint8]uint8", "map[uint8]uint8"},

		reflect.TypeOf(uint16(3)):               {"uint16", "uint16"},
		reflect.TypeOf(new(uint16)):             {"uint16", "*uint16"},
		reflect.TypeOf(make([]uint16, 0)):       {"uint16", "[]uint16"},
		reflect.TypeOf(&[]uint16{3}):            {"uint16", "*[]uint16"},
		reflect.TypeOf(make(map[uint16]uint16)): {"map[uint16]uint16", "map[uint16]uint16"},

		reflect.TypeOf(uint32(3)):               {"uint32", "uint32"},
		reflect.TypeOf(new(uint32)):             {"uint32", "*uint32"},
		reflect.TypeOf(make([]uint32, 0)):       {"uint32", "[]uint32"},
		reflect.TypeOf(&[]uint32{3}):            {"uint32", "*[]uint32"},
		reflect.TypeOf(make(map[uint32]uint32)): {"map[uint32]uint32", "map[uint32]uint32"},

		reflect.TypeOf(uint64(3)):               {"uint64", "uint64"},
		reflect.TypeOf(new(uint64)):             {"uint64", "*uint64"},
		reflect.TypeOf(make([]uint64, 0)):       {"uint64", "[]uint64"},
		reflect.TypeOf(&[]uint64{3}):            {"uint64", "*[]uint64"},
		reflect.TypeOf(make(map[uint64]uint64)): {"map[uint64]uint64", "map[uint64]uint64"},

		reflect.TypeOf(true):                {"bool", "bool"},
		reflect.TypeOf(new(bool)):           {"bool", "*bool"},
		reflect.TypeOf(make([]bool, 0)):     {"bool", "[]bool"},
		reflect.TypeOf(&[]bool{true}):       {"bool", "*[]bool"},
		reflect.TypeOf(make(map[bool]bool)): {"map[bool]bool", "map[bool]bool"},

		reflect.TypeOf(float32(3)):                {"float32", "float32"},
		reflect.TypeOf(new(float32)):              {"float32", "*float32"},
		reflect.TypeOf(make([]float32, 0)):        {"float32", "[]float32"},
		reflect.TypeOf(&[]float32{3}):             {"float32", "*[]float32"},
		reflect.TypeOf(make(map[float32]float32)): {"map[float32]float32", "map[float32]float32"},

		reflect.TypeOf(float64(3)):                {"float64", "float64"},
		reflect.TypeOf(new(float64)):              {"float64", "*float64"},
		reflect.TypeOf(make([]float64, 0)):        {"float64", "[]float64"},
		reflect.TypeOf(&[]float64{3}):             {"float64", "*[]float64"},
		reflect.TypeOf(make(map[float64]float64)): {"map[float64]float64", "map[float64]float64"},

		reflect.TypeOf(complex64(3)):                  {"complex64", "complex64"},
		reflect.TypeOf(new(complex64)):                {"complex64", "*complex64"},
		reflect.TypeOf(make([]complex64, 0)):          {"complex64", "[]complex64"},
		reflect.TypeOf(&[]complex64{3}):               {"complex64", "*[]complex64"},
		reflect.TypeOf(make(map[complex64]complex64)): {"map[complex64]complex64", "map[complex64]complex64"},

		reflect.TypeOf(complex128(3)):                   {"complex128", "complex128"},
		reflect.TypeOf(new(complex128)):                 {"complex128", "*complex128"},
		reflect.TypeOf(make([]complex128, 0)):           {"complex128", "[]complex128"},
		reflect.TypeOf(&[]complex128{3}):                {"complex128", "*[]complex128"},
		reflect.TypeOf(make(map[complex128]complex128)): {"map[complex128]complex128", "map[complex128]complex128"},

		reflect.TypeOf(make(chan int)):      {"chan int", "chan int"},
		reflect.TypeOf(make(chan struct{})): {"chan struct {}", "chan struct {}"},
		reflect.TypeOf(func() {}):           {"func()", "func()"},

		reflect.TypeOf((*error)(nil)).Elem():        {"error", "error"},
		reflect.TypeOf((*fmt.Stringer)(nil)).Elem(): {"fmt/fmt.Stringer", "fmt.Stringer"},

		reflect.TypeOf("string"):                {"string", "string"},
		reflect.TypeOf(new(string)):             {"string", "*string"},
		reflect.TypeOf(make([]string, 0)):       {"string", "[]string"},
		reflect.TypeOf(&[]string{"string"}):     {"string", "*[]string"},
		reflect.TypeOf(make(map[string]string)): {"map[string]string", "map[string]string"},

		reflect.TypeOf(pkg1.SamePkg{}):             {"github.com/go-spring/spring-core/util/testdata/pkg/bar/pkg.SamePkg", "pkg.SamePkg"},
		reflect.TypeOf(new(pkg1.SamePkg)):          {"github.com/go-spring/spring-core/util/testdata/pkg/bar/pkg.SamePkg", "*pkg.SamePkg"},
		reflect.TypeOf(make([]pkg1.SamePkg, 0)):    {"github.com/go-spring/spring-core/util/testdata/pkg/bar/pkg.SamePkg", "[]pkg.SamePkg"},
		reflect.TypeOf(&[]pkg1.SamePkg{}):          {"github.com/go-spring/spring-core/util/testdata/pkg/bar/pkg.SamePkg", "*[]pkg.SamePkg"},
		reflect.TypeOf(make(map[int]pkg1.SamePkg)): {"map[int]pkg.SamePkg", "map[int]pkg.SamePkg"},

		reflect.TypeOf(pkg2.SamePkg{}):             {"github.com/go-spring/spring-core/util/testdata/pkg/foo/pkg.SamePkg", "pkg.SamePkg"},
		reflect.TypeOf(new(pkg2.SamePkg)):          {"github.com/go-spring/spring-core/util/testdata/pkg/foo/pkg.SamePkg", "*pkg.SamePkg"},
		reflect.TypeOf(make([]pkg2.SamePkg, 0)):    {"github.com/go-spring/spring-core/util/testdata/pkg/foo/pkg.SamePkg", "[]pkg.SamePkg"},
		reflect.TypeOf(&[]pkg2.SamePkg{}):          {"github.com/go-spring/spring-core/util/testdata/pkg/foo/pkg.SamePkg", "*[]pkg.SamePkg"},
		reflect.TypeOf(make(map[int]pkg2.SamePkg)): {"map[int]pkg.SamePkg", "map[int]pkg.SamePkg"},

		reflect.TypeOf(SamePkg{}):             {"github.com/go-spring/spring-core/util/util_test.SamePkg", "util_test.SamePkg"},
		reflect.TypeOf(new(SamePkg)):          {"github.com/go-spring/spring-core/util/util_test.SamePkg", "*util_test.SamePkg"},
		reflect.TypeOf(make([]SamePkg, 0)):    {"github.com/go-spring/spring-core/util/util_test.SamePkg", "[]util_test.SamePkg"},
		reflect.TypeOf(&[]SamePkg{}):          {"github.com/go-spring/spring-core/util/util_test.SamePkg", "*[]util_test.SamePkg"},
		reflect.TypeOf(make(map[int]SamePkg)): {"map[int]util_test.SamePkg", "map[int]util_test.SamePkg"},
	}

	for i, v := range data {
		typeName := util.TypeName(i)
		assert.Equal(t, typeName, v.typeName)
		assert.Equal(t, i.String(), v.baseName)
	}
}

func TestIsPropBindingTarget(t *testing.T) {

	data := []struct {
		i interface{}
		v bool
	}{
		{true, true},                 // Bool
		{int(1), true},               // Int
		{int8(1), true},              // Int8
		{int16(1), true},             // Int16
		{int32(1), true},             // Int32
		{int64(1), true},             // Int64
		{uint(1), true},              // Uint
		{uint8(1), true},             // Uint8
		{uint16(1), true},            // Uint16
		{uint32(1), true},            // Uint32
		{uint64(1), true},            // Uint64
		{uintptr(0), false},          // Uintptr
		{float32(1), true},           // Float32
		{float64(1), true},           // Float64
		{complex64(1), false},        // Complex64
		{complex128(1), false},       // Complex128
		{[1]int{0}, true},            // Array
		{make(chan struct{}), false}, // Chan
		{func() {}, false},           // Func
		{reflect.TypeOf((*error)(nil)).Elem(), false}, // Interface
		{make(map[int]int), true},                     // Map
		{make(map[string]*int), false},                //
		{new(int), false},                             // Ptr
		{new(struct{}), false},                        //
		{[]int{0}, true},                              // Slice
		{[]*int{}, false},                             //
		{"this is a string", true},                    // String
		{struct{}{}, true},                            // Struct
		{unsafe.Pointer(new(int)), false},             // UnsafePointer
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
		i interface{}
		v bool
	}{
		{true, false},                                // Bool
		{int(1), false},                              // Int
		{int8(1), false},                             // Int8
		{int16(1), false},                            // Int16
		{int32(1), false},                            // Int32
		{int64(1), false},                            // Int64
		{uint(1), false},                             // Uint
		{uint8(1), false},                            // Uint8
		{uint16(1), false},                           // Uint16
		{uint32(1), false},                           // Uint32
		{uint64(1), false},                           // Uint64
		{uintptr(0), false},                          // Uintptr
		{float32(1), false},                          // Float32
		{float64(1), false},                          // Float64
		{complex64(1), false},                        // Complex64
		{complex128(1), false},                       // Complex128
		{[1]int{0}, false},                           // Array
		{make(chan struct{}), true},                  // Chan
		{func() {}, true},                            // Func
		{reflect.TypeOf((*error)(nil)).Elem(), true}, // Interface
		{make(map[int]int), false},                   // Map
		{make(map[string]*int), false},               //
		{new(int), false},                            //
		{new(struct{}), true},                        //
		{[]int{0}, false},                            // Slice
		{[]*int{}, false},                            //
		{"this is a string", false},                  // String
		{struct{}{}, false},                          // Struct
		{unsafe.Pointer(new(int)), false},            // UnsafePointer
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

func TestIsErrorType(t *testing.T) {
	err := fmt.Errorf("error")
	assert.True(t, util.IsErrorType(reflect.TypeOf(err)))
	err = os.ErrClosed
	assert.True(t, util.IsErrorType(reflect.TypeOf(err)))
}

func TestReturnNothing(t *testing.T) {
	assert.True(t, util.ReturnNothing(reflect.TypeOf(func() {})))
	assert.True(t, util.ReturnNothing(reflect.TypeOf(func(key string) {})))
	assert.False(t, util.ReturnNothing(reflect.TypeOf(func() string { return "" })))
}

func TestReturnOnlyError(t *testing.T) {
	assert.True(t, util.ReturnOnlyError(reflect.TypeOf(func() error { return nil })))
	assert.True(t, util.ReturnOnlyError(reflect.TypeOf(func(string) error { return nil })))
	assert.False(t, util.ReturnOnlyError(reflect.TypeOf(func() (string, error) { return "", nil })))
}

func TestIsStructPtr(t *testing.T) {
	assert.False(t, util.IsStructPtr(reflect.TypeOf(3)))
	assert.False(t, util.IsStructPtr(reflect.TypeOf(func() {})))
	assert.False(t, util.IsStructPtr(reflect.TypeOf(struct{}{})))
	assert.False(t, util.IsStructPtr(reflect.TypeOf(struct{ a string }{})))
	assert.True(t, util.IsStructPtr(reflect.TypeOf(&struct{ a string }{})))
}

func TestIsConstructor(t *testing.T) {
	assert.False(t, util.IsConstructor(reflect.TypeOf(func() {})))
	assert.True(t, util.IsConstructor(reflect.TypeOf(func() string { return "" })))
	assert.True(t, util.IsConstructor(reflect.TypeOf(func() *string { return nil })))
	assert.True(t, util.IsConstructor(reflect.TypeOf(func() *testdata.Receiver { return nil })))
	assert.True(t, util.IsConstructor(reflect.TypeOf(func() (*testdata.Receiver, error) { return nil, nil })))
	assert.False(t, util.IsConstructor(reflect.TypeOf(func() (bool, *testdata.Receiver, error) { return false, nil, nil })))
}

func TestHasReceiver(t *testing.T) {
	assert.False(t, util.HasReceiver(reflect.TypeOf(func() {}), reflect.ValueOf(new(testdata.Receiver))))
	assert.True(t, util.HasReceiver(reflect.TypeOf(func(*testdata.Receiver) {}), reflect.ValueOf(new(testdata.Receiver))))
	assert.True(t, util.HasReceiver(reflect.TypeOf(func(*testdata.Receiver, int) {}), reflect.ValueOf(new(testdata.Receiver))))
	assert.True(t, util.HasReceiver(reflect.TypeOf(func(fmt.Stringer, int) {}), reflect.ValueOf(new(testdata.Receiver))))
	assert.False(t, util.HasReceiver(reflect.TypeOf(func(error, int) {}), reflect.ValueOf(new(testdata.Receiver))))
}

func TestIsBeanInjectionTarget(t *testing.T) {
	assert.False(t, util.IsBeanInjectionTarget(reflect.TypeOf("abc")))
	assert.False(t, util.IsBeanInjectionTarget(reflect.TypeOf(new(string))))
	assert.True(t, util.IsBeanInjectionTarget(reflect.TypeOf(errors.New("abc"))))
	assert.False(t, util.IsBeanInjectionTarget(reflect.TypeOf([]string{})))
	assert.False(t, util.IsBeanInjectionTarget(reflect.TypeOf([]*string{})))
	assert.True(t, util.IsBeanInjectionTarget(reflect.TypeOf([]fmt.Stringer{})))
	assert.False(t, util.IsBeanInjectionTarget(reflect.TypeOf(map[string]string{})))
	assert.False(t, util.IsBeanInjectionTarget(reflect.TypeOf(map[string]*string{})))
	assert.True(t, util.IsBeanInjectionTarget(reflect.TypeOf(map[string]fmt.Stringer{})))
}
