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

package gs_cond

import (
	"fmt"

	"github.com/expr-lang/expr"
)

// funcMap holds a map of function names to their implementations.
// These functions can be used within expressions passed to EvalExpr.
var funcMap = map[string]interface{}{}

// RegisterExpressFunc registers an express function with a specified name.
// The function can later be used in expressions evaluated by EvalExpr.
func RegisterExpressFunc(name string, fn interface{}) {
	funcMap[name] = fn
}

// EvalExpr evaluates a given boolean expression (input) using the provided value (val).
// The `input` parameter is a string that represents an expression expected to return a
// boolean value. The `val` parameter is the data object that will be referred to as "$"
// within the expression context.
func EvalExpr(input string, val interface{}) (bool, error) {
	env := map[string]interface{}{"$": val}
	for k, v := range funcMap {
		env[k] = v
	}
	r, err := expr.Eval(input, env)
	if err != nil {
		return false, fmt.Errorf("eval %q returns error, %w", input, err)
	}
	ret, ok := r.(bool)
	if !ok {
		return false, fmt.Errorf("eval %q doesn't return bool value", input)
	}
	return ret, nil
}
