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

/*
Package gstest provides unit testing utilities for dependency injection in Go-Spring framework.

Key Features:
  - Test environment configuration: jobs and servers are disabled, and the "test" profile is automatically activated
  - Autowire failure tolerance: non-critical autowiring errors are tolerated so that missing beans do not break tests
  - Type-safe mocking: compile-time checked MockFor/With methods for registering mock beans
  - Context lifecycle management: TestMain starts and stops the Go-Spring context automatically
  - Injection helpers: Get[T](t) and Wire(t, obj) simplify bean retrieval and dependency injection

Usage Pattern:

	// Step 1: Register your mock beans before tests run
	// by calling `MockFor[T]().With(obj)` inside an `init()` function.
	func init() {
	    gstest.MockFor[*Dao]().With(&MockDao{})
	}

	// Step 2: Implement TestMain and invoke `gstest.TestMain(m, opts...)`
	// to bootstrap the application context, execute all tests, and then shut it down.
	// You can supply `BeforeRun` and `AfterRun` hooks to run code immediately before or after your test suite.
	func TestMain(m *testing.M) {
		gstest.TestMain(m)
	}

	// Step 3: Write your test cases and use Get[T](t) or Wire(t, obj) to retrieve beans and inject dependencies.
	func TestService(t *testing.T) {
	    // Retrieve autowired test target
	    service := gstest.Get[*Service](t)

	    // Verify business logic
	    result := service.Process()
	    assert.Equal(t, expect, result)
	}
*/
package gstest

import (
	"testing"

	"github.com/go-spring/spring-core/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_app"
)

func init() {
	gs.EnableJobs(false)
	gs.EnableServers(false)
	gs.SetActiveProfiles("test")
	gs.ForceAutowireIsNullable(true)
}

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
	gs_app.GS.C.AddMock(gs.BeanMock{
		Object: obj,
		Target: m.selector,
	})
}

type runArg struct {
	beforeRun func()
	afterRun  func()
}

type RunOption func(arg *runArg)

// BeforeRun specifies a function to be executed before all testcases.
func BeforeRun(fn func()) RunOption {
	return func(arg *runArg) {
		arg.beforeRun = fn
	}
}

// AfterRun specifies a function to be executed after all testcases.
func AfterRun(fn func()) RunOption {
	return func(arg *runArg) {
		arg.afterRun = fn
	}
}

// TestMain executes test cases and ensures shutdown of the app context.
func TestMain(m *testing.M, opts ...RunOption) {
	arg := &runArg{}
	for _, opt := range opts {
		opt(arg)
	}

	err := gs_app.GS.Start()
	if err != nil {
		panic(err)
	}

	if arg.beforeRun != nil {
		arg.beforeRun()
	}

	m.Run()

	if arg.afterRun != nil {
		arg.afterRun()
	}

	gs_app.GS.Stop()
}

// Get gets the bean from the app context.
func Get[T any](t *testing.T) T {
	var s struct {
		Value T `autowire:""`
	}
	return Wire(t, &s).Value
}

// Wire injects dependencies into the object.
func Wire[T any](t *testing.T, obj T) T {
	err := gs_app.GS.C.Wire(obj)
	if err != nil {
		t.Fatal(err)
	}
	return obj
}
