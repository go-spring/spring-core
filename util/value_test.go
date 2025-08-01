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
	"fmt"
	"reflect"
	"testing"

	"github.com/go-spring/gs-assert/assert"
	"github.com/go-spring/spring-core/util"
)

func TestPatchValue(t *testing.T) {
	var r struct{ v int }
	v := reflect.ValueOf(&r)
	v = v.Elem().Field(0)
	assert.Panic(t, func() {
		v.SetInt(4)
	}, "using value obtained using unexported field")
	v = util.PatchValue(v)
	v.SetInt(4)
}

func TestFuncName(t *testing.T) {
	assert.That(t, util.FuncName(func() {})).Equal("util_test.TestFuncName.func1")
	assert.That(t, util.FuncName(func(i int) {})).Equal("util_test.TestFuncName.func2")
	assert.That(t, util.FuncName(fnNoArgs)).Equal("util_test.fnNoArgs")
	assert.That(t, util.FuncName(fnWithArgs)).Equal("util_test.fnWithArgs")
	assert.That(t, util.FuncName((*receiver).ptrFnNoArgs)).Equal("util_test.(*receiver).ptrFnNoArgs")
	assert.That(t, util.FuncName((*receiver).ptrFnWithArgs)).Equal("util_test.(*receiver).ptrFnWithArgs")
}

func fnNoArgs() {}

func fnWithArgs(i int) {}

type receiver struct{}

func (r *receiver) ptrFnNoArgs() {}

func (r *receiver) ptrFnWithArgs(i int) {}

func TestFileLine(t *testing.T) {
	testcases := []struct {
		fn     any
		file   string
		line   int
		fnName string
	}{
		{
			fnNoArgs,
			"spring-core/util/value_test.go",
			48,
			"util_test.fnNoArgs",
		},
		{
			fnWithArgs,
			"spring-core/util/value_test.go",
			50,
			"util_test.fnWithArgs",
		},
		{
			(*receiver).ptrFnNoArgs,
			"spring-core/util/value_test.go",
			54,
			"util_test.(*receiver).ptrFnNoArgs",
		},
		{
			(*receiver).ptrFnWithArgs,
			"spring-core/util/value_test.go",
			56,
			"util_test.(*receiver).ptrFnWithArgs",
		},
	}
	for i, c := range testcases {
		file, line, fnName := util.FileLine(c.fn)
		assert.That(t, line).Equal(c.line, fmt.Sprint(i))
		assert.That(t, fnName).Equal(c.fnName, fmt.Sprint(i))
		assert.ThatString(t, file).HasSuffix(c.file, fmt.Sprint(i))
	}
}
