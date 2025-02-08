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
	"reflect"
	"testing"

	"github.com/go-spring/spring-core/util"
	"github.com/go-spring/spring-core/util/assert"
	"github.com/go-spring/spring-core/util/testdata"
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

func fnNoArgs() {}

func fnWithArgs(i int) {}

type receiver struct{}

func (r receiver) fnNoArgs() {}

func (r receiver) fnWithArgs(i int) {}

func (r *receiver) ptrFnNoArgs() {}

func (r *receiver) ptrFnWithArgs(i int) {}

func TestFileLine(t *testing.T) {
	offset := 62
	testcases := []struct {
		fn     interface{}
		file   string
		line   int
		fnName string
	}{
		{
			fn:     fnNoArgs,
			file:   "spring-core/util/value_test.go",
			line:   offset - 15,
			fnName: "fnNoArgs",
		},
		{
			fnWithArgs,
			"spring-core/util/value_test.go",
			offset - 13,
			"fnWithArgs",
		},
		{
			receiver{}.fnNoArgs,
			"spring-core/util/value_test.go",
			offset - 9,
			"receiver.fnNoArgs",
		},
		{
			receiver.fnNoArgs,
			"spring-core/util/value_test.go",
			offset - 9,
			"receiver.fnNoArgs",
		},
		{
			receiver{}.fnWithArgs,
			"spring-core/util/value_test.go",
			offset - 7,
			"receiver.fnWithArgs",
		},
		{
			receiver.fnWithArgs,
			"spring-core/util/value_test.go",
			offset - 7,
			"receiver.fnWithArgs",
		},
		{
			(&receiver{}).ptrFnNoArgs,
			"spring-core/util/value_test.go",
			offset - 5,
			"(*receiver).ptrFnNoArgs",
		},
		{
			(*receiver).ptrFnNoArgs,
			"spring-core/util/value_test.go",
			offset - 5,
			"(*receiver).ptrFnNoArgs",
		},
		{
			(&receiver{}).ptrFnWithArgs,
			"spring-core/util/value_test.go",
			offset - 3,
			"(*receiver).ptrFnWithArgs",
		},
		{
			(*receiver).ptrFnWithArgs,
			"spring-core/util/value_test.go",
			offset - 3,
			"(*receiver).ptrFnWithArgs",
		},
		{
			testdata.FnNoArgs,
			"spring-core/util/testdata/pkg.go",
			19,
			"FnNoArgs",
		},
		{
			testdata.FnWithArgs,
			"spring-core/util/testdata/pkg.go",
			21,
			"FnWithArgs",
		},
		{
			testdata.Receiver{}.FnNoArgs,
			"spring-core/util/testdata/pkg.go",
			25,
			"Receiver.FnNoArgs",
		},
		{
			testdata.Receiver{}.FnWithArgs,
			"spring-core/util/testdata/pkg.go",
			27,
			"Receiver.FnWithArgs",
		},
		{
			(&testdata.Receiver{}).PtrFnNoArgs,
			"spring-core/util/testdata/pkg.go",
			29,
			"(*Receiver).PtrFnNoArgs",
		},
		{
			(&testdata.Receiver{}).PtrFnWithArgs,
			"spring-core/util/testdata/pkg.go",
			31,
			"(*Receiver).PtrFnWithArgs",
		},
		{
			testdata.Receiver.FnNoArgs,
			"spring-core/util/testdata/pkg.go",
			25,
			"Receiver.FnNoArgs",
		},
		{
			testdata.Receiver.FnWithArgs,
			"spring-core/util/testdata/pkg.go",
			27,
			"Receiver.FnWithArgs",
		},
		{
			(*testdata.Receiver).PtrFnNoArgs,
			"spring-core/util/testdata/pkg.go",
			29,
			"(*Receiver).PtrFnNoArgs",
		},
		{
			(*testdata.Receiver).PtrFnWithArgs,
			"spring-core/util/testdata/pkg.go",
			31,
			"(*Receiver).PtrFnWithArgs",
		},
	}
	for _, c := range testcases {
		file, line, fnName := util.FileLine(c.fn)
		assert.String(t, file).HasSuffix(c.file)
		assert.Equal(t, line, c.line)
		assert.Equal(t, fnName, c.fnName)
	}
}
