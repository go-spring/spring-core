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

package gstest

import (
	"testing"

	"github.com/go-spring/spring-core/gs"
	"github.com/go-spring/spring-core/util/assert"
)

// GSContext is the global context for testing.
var GSContext gs.Context

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

// Run executes test cases and ensures shutdown of the app context.
func Run(m *testing.M, opts ...RunOption) {
	arg := &runArg{}
	for _, opt := range opts {
		opt(arg)
	}

	gs.ForceAutowireIsNullable(true)

	var err error
	GSContext, err = gs.Start()
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

	gs.Stop()
}

// Case calls a function with arguments injected.
func Case(t *testing.T, fn interface{}, args ...gs.Arg) {
	_, err := GSContext.Invoke(fn, args...)
	assert.Nil(t, err)
}
