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

package conf

import (
	"fmt"
	"maps"

	"github.com/expr-lang/expr"
)

// ValidateFunc defines a type for validation functions, which accept
// a value of type T and return a boolean result.
type ValidateFunc[T any] func(T) bool

// validateFuncs holds a map of registered validation functions.
var validateFuncs = map[string]any{}

// RegisterValidateFunc registers a validation function with a specific name.
// The function can then be used in validation expressions.
func RegisterValidateFunc[T any](name string, fn ValidateFunc[T]) {
	validateFuncs[name] = fn
}

// validateField validates a field using a validation expression (tag) and the field value (i).
// It evaluates the expression and checks if the result is true (i.e., the validation passes).
// If any error occurs during evaluation or if the validation fails, an error is returned.
func validateField(tag string, i any) error {
	env := map[string]any{"$": i}
	maps.Copy(env, validateFuncs)
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
