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

package gs

import (
	"reflect"
	"strings"
	"testing"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/util"
)

// BeanMock is a mock for bean.
type BeanMock[T any] struct {
	selector gs.BeanSelector
}

// MockFor creates a mock for bean.
func MockFor[T any](name ...string) BeanMock[T] {
	return BeanMock[T]{
		selector: gs.BeanSelectorFor[T](name...),
	}
}

// With registers a mock bean.
func (m BeanMock[T]) With(obj T) {
	app.C.AddMock(gs.BeanMock{
		Object: obj,
		Target: m.selector,
	})
}

var testers []any

// AddTester adds a tester to the test suite.
func AddTester(t any) {
	testers = append(testers, t)
	app.C.RootBean(app.C.Object(t))
}

// TestMain is the entry point for testing.
func TestMain(m *testing.M) {

	// patch m.tests
	mValue := util.PatchValue(reflect.ValueOf(m))
	fValue := util.PatchValue(mValue.Elem().FieldByName("tests"))
	tests := fValue.Interface().([]testing.InternalTest)
	for _, tester := range testers {
		tt := reflect.TypeOf(tester)
		typeName := tt.Elem().String()
		for i := range tt.NumMethod() {
			methodType := tt.Method(i)
			if strings.HasPrefix(methodType.Name, "Test") {
				tests = append(tests, testing.InternalTest{
					Name: typeName + "." + methodType.Name,
					F: func(t *testing.T) {
						testMethod := reflect.ValueOf(tester).Method(i)
						testMethod.Call([]reflect.Value{reflect.ValueOf(t)})
					},
				})
			}
		}
	}
	fValue.Set(reflect.ValueOf(tests))

	// run app
	stop, err := RunAsync()
	if err != nil {
		panic(err)
	}

	// run test
	m.Run()

	// stop app
	stop()
}
