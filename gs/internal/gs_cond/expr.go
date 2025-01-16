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

var exprFuncs = map[string]interface{}{}

func RegisterExprFunc(name string, fn interface{}) {
	exprFuncs[name] = fn
}

// EvalExpr returns the value for the expression expr.
func EvalExpr(input string, val interface{}) (bool, error) {
	env := map[string]interface{}{"$": val}
	for k, v := range exprFuncs {
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
