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

	"github.com/go-spring/spring-base/util"
	"github.com/go-spring/spring-core/gs/internal/gs"
)

// BeanMock represents a mock bean for testing.
type BeanMock[T any] struct {
	selector gs.BeanSelector
}

// MockFor creates a BeanMock for the given type and optional bean name.
// It allows you to specify which bean in the IoC container should be mocked.
func MockFor[T any](name ...string) BeanMock[T] {
	return BeanMock[T]{
		selector: gs.BeanSelectorFor[T](name...),
	}
}

// With registers a mock instance into the IoC container,
// replacing the original bean defined by the selector.
// This allows tests to use mocked dependencies.
func (m BeanMock[T]) With(obj T) {
	app.C.AddMock(gs.BeanMock{
		Object: obj,
		Target: m.selector,
	})
}

// testers stores all registered tester instances.
// Each tester can contain multiple test methods.
var testers []any

// AddTester registers a tester instance into the test suite.
// The tester will be scanned for methods prefixed with "Test",
// which will be automatically added to the Go test framework.
func AddTester(t any) {
	testers = append(testers, t)
	app.C.Root(app.C.Object(t))
}

// TestMain is the custom entry point for the Go test framework.
// It injects test methods defined in registered testers into the
// internal 'tests' slice of testing.M, then starts the app and tests.
func TestMain(m *testing.M) {

	// Patch m.tests using reflection (a non-standard hack).
	// This allows dynamically adding test cases at runtime.
	mValue := util.PatchValue(reflect.ValueOf(m))
	fValue := util.PatchValue(mValue.Elem().FieldByName("tests"))
	tests := fValue.Interface().([]testing.InternalTest)

	// Scan all registered testers for methods starting with "Test".
	for _, tester := range testers {
		tt := reflect.TypeOf(tester)
		typeName := tt.Elem().String()
		for i := 0; i < tt.NumMethod(); i++ {
			methodType := tt.Method(i)
			// Only consider methods whose names start with "Test"
			if strings.HasPrefix(methodType.Name, "Test") {
				tests = append(tests, testing.InternalTest{
					Name: typeName + "." + methodType.Name, // Full test name
					F: func(t *testing.T) { // Test function to execute
						testMethod := reflect.ValueOf(tester).Method(i)
						testMethod.Call([]reflect.Value{reflect.ValueOf(t)})
					},
				})
			}
		}
	}
	fValue.Set(reflect.ValueOf(tests))

	// Run the application asynchronously.
	stop, err := RunAsync()
	if err != nil {
		panic(err)
	}

	// Run all collected tests.
	m.Run()

	// Stop the application gracefully after all tests complete.
	stop()
}
