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
)

// GSContext is the global context for testing.
var GSContext gs.Context

// TestingContext is the context for testing.
type TestingContext struct {
	gs.ContextAware
}

func init() {
	gs.ForceAutowireIsNullable(true)
	gs.Object(&TestingContext{}).Init(func(tc *TestingContext) {
		GSContext = tc.GSContext
	})
}

// Init initializes the test environment.
func Init() error {
	return gs.Start()
}

// Run executes test cases and ensures shutdown of the app context.
func Run(m *testing.M) (code int) {
	code = m.Run()
	gs.Stop()
	return code
}

// Case calls a function with arguments injected.
func Case(fn interface{}, args ...gs.Arg) {
	_, _ = GSContext.Invoke(fn, args...)
}
