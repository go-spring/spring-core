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

package conf

import (
	"fmt"

	"github.com/expr-lang/expr"
)

type ValidateFunc interface{}

var validateFuncs = map[string]interface{}{}

func RegisterValidateFunc(name string, fn ValidateFunc) {
	validateFuncs[name] = fn
}

// validateField validates the field with the given tag and value.
func validateField(tag string, i interface{}) error {
	env := map[string]interface{}{"$": i}
	for k, v := range validateFuncs {
		env[k] = v
	}
	r, err := expr.Eval(tag, env)
	if err != nil {
		return fmt.Errorf("eval %q returns error, %w", tag, err)
	}
	ret, ok := r.(bool)
	if !ok {
		return fmt.Errorf("eval %q doesn't return bool value", tag)
	}
	if !ret {
		return fmt.Errorf("validate failed on %q for value %v", tag, i)
	}
	return nil
}
