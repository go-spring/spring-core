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

	"github.com/go-spring/spring-core/util"
	"github.com/go-spring/spring-core/util/assert"
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

func (r *receiver) ptrFnNoArgs() {}

func (r *receiver) ptrFnWithArgs(i int) {}

func TestFileLine(t *testing.T) {
	testcases := []struct {
		fn     interface{}
		file   string
		line   int
		fnName string
	}{
		{
			fnNoArgs,
			"spring-core/util/value_test.go",
			39,
			"util_test.fnNoArgs",
		},
		{
			fnWithArgs,
			"spring-core/util/value_test.go",
			41,
			"util_test.fnWithArgs",
		},
		{
			(*receiver).ptrFnNoArgs,
			"spring-core/util/value_test.go",
			45,
			"util_test.(*receiver).ptrFnNoArgs",
		},
		{
			(*receiver).ptrFnWithArgs,
			"spring-core/util/value_test.go",
			47,
			"util_test.(*receiver).ptrFnWithArgs",
		},
	}
	for i, c := range testcases {
		file, line, fnName := util.FileLine(c.fn)
		assert.Equal(t, line, c.line, fmt.Sprint(i))
		assert.Equal(t, fnName, c.fnName, fmt.Sprint(i))
		assert.String(t, file).HasSuffix(c.file, fmt.Sprint(i))
	}
}
