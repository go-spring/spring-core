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

package gs_cond

import (
	"fmt"

	"github.com/expr-lang/expr"
)

// funcMap is used to store custom functions that can be referenced in expressions.
var funcMap = map[string]interface{}{}

// CustomFunction registers a custom function with a specified name.
// The function can later be used in expressions evaluated by EvalExpr.
func CustomFunction(name string, fn interface{}) {
	funcMap[name] = fn
}

// EvalExpr evaluates a boolean expression (input) using the provided value (val)
// and any registered custom functions. The string expression to evaluate (should
// return a boolean value). The value used in the evaluation (referred to as `$`
// in the expression context).
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
