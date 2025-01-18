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

// Package gs_cond provides utilities for evaluating contextual conditions
// in a flexible, modular, and reusable manner.
//
// The core concept revolves around the [gs.Condition] interface, which
// determines whether a given context satisfies specific criteria.
// Various implementations of [gs.Condition] enable the combination of
// basic and complex logical operations such as `And`, `Or`, and `Not`.
//
// These utilities are further enhanced with support for:
// - Property-based conditions
// - Bean existence checks
// - Custom evaluation logic
//
// When implementing custom conditions, note that only terminal conditions
// should return [gs.ConditionError].
package gs_cond

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/util"
)

// ConditionError is a wrapper of Condition and error.
type ConditionError struct {
	err   error
	stack []gs.Condition
}

// NewCondError returns a new ConditionError.
func NewCondError(cond gs.Condition, err error) error {
	var e *ConditionError
	if errors.As(err, &e) {
		e.stack = append(e.stack, cond)
		return e
	}
	return &ConditionError{err: err, stack: []gs.Condition{cond}}
}

// Error returns the error message.
func (e *ConditionError) Error() string {
	var sb strings.Builder
	sb.WriteString("condition error: ")
	for i := len(e.stack) - 1; i >= 0; i-- {
		c := e.stack[i]
		switch c.(type) {
		case *onOr:
			sb.WriteString("Or(...)")
		case *onAnd:
			sb.WriteString("And(...)")
		case *onNone:
			sb.WriteString("None(...)")
		default:
			sb.WriteString(fmt.Sprint(c))
		}
		if i > 0 {
			sb.WriteString(" -> ")
		}
	}
	sb.WriteString(" -> ")
	sb.WriteString(e.err.Error())
	return sb.String()
}

// Unwrap returns the error wrapped by ConditionError.
func (e *ConditionError) Unwrap() error { return e.err }

/********************************* OnFunc ************************************/

// onFunc is an implementation of [gs.Condition] that wraps a function.
type onFunc struct {
	fn gs.CondFunc
}

// OnFunc creates a Conditional that starts with a condition to
// match based on a custom function. The function takes a CondContext
// and returns a boolean value and an optional error.
func OnFunc(fn gs.CondFunc) gs.Condition {
	return &onFunc{fn: fn}
}

func (c *onFunc) Matches(ctx gs.CondContext) (bool, error) {
	ok, err := c.fn(ctx)
	if err != nil {
		return false, NewCondError(c, err)
	}
	return ok, nil
}

func (c *onFunc) String() string {
	_, _, fnName := util.FileLine(c.fn)
	return fmt.Sprintf("OnFunc(fn=%s)", fnName)
}

/******************************* OnProperty **********************************/

// onProperty evaluates a condition based on the existence and value of a property.
// - If the property is missing, the result is determined by `matchIfMissing`.
// - If `havingValue` is provided, the property's value must match it.
// - If `havingValue` starts with "expr:", the value is evaluated as an expression.
type onProperty struct {
	name           string // Name of the property to check.
	havingValue    string // Expected value or expression.
	matchIfMissing bool   // Result if the property is missing.
}

// PropertyOption is a function type that modifies the behavior of an
// `onProperty` condition. It allows customizing how a property is evaluated.
type PropertyOption func(*onProperty)

// MatchIfMissing is a property option that configures the condition
// to match if the property is missing. When applied, the condition
// will return true if the specified property does not exist.
func MatchIfMissing() PropertyOption {
	return func(c *onProperty) {
		c.matchIfMissing = true
	}
}

// HavingValue is a property option that sets a specific value the
// property must match. When applied, the condition will return true
// if the property value equals `havingValue`.
func HavingValue(havingValue string) PropertyOption {
	return func(c *onProperty) {
		c.havingValue = havingValue
	}
}

// OnProperty creates a Conditional that starts with a condition checking
// a specific property and its value. Additional property options can be
// provided to customize the behavior.
func OnProperty(name string, options ...PropertyOption) gs.Condition {
	c := &onProperty{name: name}
	for _, option := range options {
		option(c)
	}
	return c
}

func (c *onProperty) Matches(ctx gs.CondContext) (bool, error) {

	if !ctx.Has(c.name) {
		return c.matchIfMissing, nil
	}

	if c.havingValue == "" {
		return true, nil
	}

	val := ctx.Prop(c.name)
	if !strings.HasPrefix(c.havingValue, "expr:") {
		return val == c.havingValue, nil
	}

	getValue := func(val string) interface{} {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
		if u, err := strconv.ParseUint(val, 10, 64); err == nil {
			return u
		}
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
		return val
	}
	ok, err := EvalExpr(c.havingValue[5:], getValue(val))
	if err != nil {
		return false, NewCondError(c, err)
	}
	return ok, nil
}

func (c *onProperty) String() string {
	var sb strings.Builder
	sb.WriteString("OnProperty(name=")
	sb.WriteString(c.name)
	if c.havingValue != "" {
		sb.WriteString(", havingValue=")
		sb.WriteString(c.havingValue)
	}
	if c.matchIfMissing {
		sb.WriteString(", matchIfMissing=true")
	}
	sb.WriteString(")")
	return sb.String()
}

/*************************** OnMissingProperty *******************************/

// onMissingProperty evaluates to true when a specified property is absent in the context.
type onMissingProperty struct {
	name string
}

// OnMissingProperty creates a Conditional that starts with a condition
// that matches if a property does not exist.
func OnMissingProperty(name string) gs.Condition {
	return &onMissingProperty{name: name}
}

func (c *onMissingProperty) Matches(ctx gs.CondContext) (bool, error) {
	return !ctx.Has(c.name), nil
}

func (c *onMissingProperty) String() string {
	return fmt.Sprintf("OnMissingProperty(name=%s)", c.name)
}

/********************************* OnBean ************************************/

// onBean checks for the existence of beans matching a selector.
// It evaluates to true if at least one bean matches.
type onBean struct {
	selector gs.BeanSelector
}

// OnBean creates a Conditional that starts with a condition to match
// when more than one bean is found matching the provided selector.
func OnBean(selector gs.BeanSelector) gs.Condition {
	return &onBean{selector: selector}
}

func (c *onBean) Matches(ctx gs.CondContext) (bool, error) {
	beans, err := ctx.Find(c.selector)
	if err != nil {
		return false, NewCondError(c, err)
	}
	return len(beans) > 0, nil
}

func (c *onBean) String() string {
	return fmt.Sprintf("OnBean(selector=%s)", gs.BeanSelectorToString(c.selector))
}

/***************************** OnMissingBean *********************************/

// onMissingBean evaluates to true if no beans match the selector.
type onMissingBean struct {
	selector gs.BeanSelector
}

// OnMissingBean creates a Conditional that starts with a condition
// to match when no beans are found matching the provided selector.
func OnMissingBean(selector gs.BeanSelector) gs.Condition {
	return &onMissingBean{selector: selector}
}

func (c *onMissingBean) Matches(ctx gs.CondContext) (bool, error) {
	beans, err := ctx.Find(c.selector)
	if err != nil {
		return false, NewCondError(c, err)
	}
	return len(beans) == 0, nil
}

func (c *onMissingBean) String() string {
	return fmt.Sprintf("OnMissingBean(selector=%s)", gs.BeanSelectorToString(c.selector))
}

/***************************** OnSingleBean **********************************/

// onSingleBean checks for exactly one matching bean.
type onSingleBean struct {
	selector gs.BeanSelector
}

// OnSingleBean creates a Conditional that starts with a condition
// to match when exactly one bean is found matching the provided selector.
func OnSingleBean(selector gs.BeanSelector) gs.Condition {
	return &onSingleBean{selector: selector}
}

func (c *onSingleBean) Matches(ctx gs.CondContext) (bool, error) {
	beans, err := ctx.Find(c.selector)
	if err != nil {
		return false, NewCondError(c, err)
	}
	return len(beans) == 1, nil
}

func (c *onSingleBean) String() string {
	return fmt.Sprintf("OnSingleBean(selector=%s)", gs.BeanSelectorToString(c.selector))
}

/***************************** OnExpression **********************************/

// onExpression evaluates a custom expression within the context.
type onExpression struct {
	expression string
}

// OnExpression creates a Conditional that starts with a condition
// to match based on the evaluation of a string expression. The expression
// should return true or false.
func OnExpression(expression string) gs.Condition {
	return &onExpression{expression: expression}
}

func (c *onExpression) Matches(ctx gs.CondContext) (bool, error) {
	return false, NewCondError(c, util.UnimplementedMethod)
}

func (c *onExpression) String() string {
	return fmt.Sprintf("OnExpression(expression=%s)", c.expression)
}

/********************************** Not **************************************/

// not is an implementation of [gs.Condition] that negates another condition.
type not struct {
	c gs.Condition
}

// Not creates a [gs.Condition] that inverts the result of another condition.
func Not(c gs.Condition) gs.Condition {
	return &not{c: c}
}

func (c *not) Matches(ctx gs.CondContext) (bool, error) {
	ok, err := c.c.Matches(ctx)
	if err != nil {
		return false, NewCondError(c, err)
	}
	return !ok, nil
}

func (c *not) String() string {
	return fmt.Sprintf("Not(%s)", c.c)
}

/********************************** Or ***************************************/

type onOr struct {
	cond []gs.Condition
}

// Or combines conditions with an OR operator. Returns a condition that
// evaluates to true if at least one condition is satisfied.
func Or(cond ...gs.Condition) gs.Condition {
	if n := len(cond); n == 0 {
		return nil
	} else if n == 1 {
		return cond[0]
	}
	return &onOr{cond: cond}
}

func (g *onOr) Matches(ctx gs.CondContext) (bool, error) {
	for _, c := range g.cond {
		if ok, err := c.Matches(ctx); err != nil {
			return false, NewCondError(g, err)
		} else if ok {
			return true, nil
		}
	}
	return false, nil
}

func (g *onOr) String() string {
	return FormatGroup("Or", g.cond)
}

/********************************* And ***************************************/

type onAnd struct {
	cond []gs.Condition
}

// And combines conditions with an AND operator. Returns a condition that
// evaluates to true only if all conditions are satisfied.
func And(cond ...gs.Condition) gs.Condition {
	if n := len(cond); n == 0 {
		return nil
	} else if n == 1 {
		return cond[0]
	}
	return &onAnd{cond: cond}
}

func (g *onAnd) String() string {
	return FormatGroup("And", g.cond)
}

func (g *onAnd) Matches(ctx gs.CondContext) (bool, error) {
	for _, c := range g.cond {
		if ok, err := c.Matches(ctx); err != nil {
			return false, NewCondError(g, err)
		} else if !ok {
			return false, nil
		}
	}
	return true, nil
}

/********************************** None *************************************/

type onNone struct {
	cond []gs.Condition
}

// None combines conditions with a NONE operator. Returns a condition that
// evaluates to true only if none of the conditions are satisfied.
func None(cond ...gs.Condition) gs.Condition {
	if n := len(cond); n == 0 {
		return nil
	} else if n == 1 {
		return Not(cond[0])
	}
	return &onNone{cond: cond}
}

func (g *onNone) Matches(ctx gs.CondContext) (bool, error) {
	for _, c := range g.cond {
		if ok, err := c.Matches(ctx); err != nil {
			return false, NewCondError(g, err)
		} else if ok {
			return false, nil
		}
	}
	return true, nil
}

func (g *onNone) String() string {
	return FormatGroup("None", g.cond)
}

/******************************* utilities ***********************************/

func FormatGroup(op string, cond []gs.Condition) string {
	var sb strings.Builder
	sb.WriteString(op)
	sb.WriteString("(")
	for i, c := range cond {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprint(c))
	}
	sb.WriteString(")")
	return sb.String()
}
