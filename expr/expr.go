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

package expr

import (
	"fmt"

	"github.com/expr-lang/expr"
)

type Validator struct{}

// Name returns the name of the validator.
func (d *Validator) Name() string {
	return "expr"
}

// Field validates the field with the given tag and value.
func (d *Validator) Field(tag string, i interface{}) error {
	if ret, err := Eval(tag, i); err != nil {
		return err
	} else if !ret {
		return fmt.Errorf("validate failed on %q for value %v", tag, i)
	}
	return nil
}

// Eval returns the value for the expression expr.
func Eval(input string, val interface{}) (bool, error) {
	r, err := expr.Eval(input, map[string]interface{}{"$": val})
	if err != nil {
		return false, fmt.Errorf("eval %q returns error, %w", input, err)
	}
	ret, ok := r.(bool)
	if !ok {
		return false, fmt.Errorf("eval %q doesn't return bool value", input)
	}
	return ret, nil
}
