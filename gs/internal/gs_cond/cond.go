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
	"strings"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/util"
	"github.com/go-spring/spring-core/util/errutil"
)

/********************************* OnFunc ************************************/

// onFunc is an implementation of [gs.Condition] that wraps a function.
// It allows a condition to be evaluated based on the result of a function.
type onFunc struct {
	fn func(ctx gs.CondContext) (bool, error)
}

// OnFunc creates a Conditional that evaluates using a custom function.
func OnFunc(fn func(ctx gs.CondContext) (bool, error)) gs.Condition {
	return &onFunc{fn: fn}
}

// Matches checks if the condition is met according to the provided context.
func (c *onFunc) Matches(ctx gs.CondContext) (bool, error) {
	ok, err := c.fn(ctx)
	if err != nil {
		return false, errutil.WrapError(err, "condition matches error: %s", c)
	}
	return ok, nil
}

func (c *onFunc) String() string {
	_, _, fnName := util.FileLine(c.fn)
	return fmt.Sprintf("OnFunc(fn=%s)", fnName)
}

/******************************* OnProperty **********************************/

// OnPropertyInterface defines the methods for evaluating a condition based on a property.
// This interface provides flexibility for matching missing properties and checking their values.
type OnPropertyInterface interface {
	gs.Condition
	MatchIfMissing() OnPropertyInterface
	HavingValue(s string) OnPropertyInterface
}

// onProperty evaluates a condition based on the existence and value of a property
// in the context. It allows for complex matching behaviors such as matching missing
// properties or evaluating expressions.
type onProperty struct {
	name           string // The name of the property to check.
	matchIfMissing bool   // Whether to match if the property is missing.
	havingValue    any    // The expected value or expression to match.
}

// OnProperty creates a condition based on the presence and value of a specified property.
func OnProperty(name string) OnPropertyInterface {
	return &onProperty{name: name}
}

// MatchIfMissing sets the condition to match if the property is missing.
func (c *onProperty) MatchIfMissing() OnPropertyInterface {
	c.matchIfMissing = true
	return c
}

// HavingValue sets the expected value or expression to match.
func (c *onProperty) HavingValue(s string) OnPropertyInterface {
	c.havingValue = s
	return c
}

// Matches checks if the condition is met according to the provided context.
func (c *onProperty) Matches(ctx gs.CondContext) (bool, error) {

	// If the context doesn't have the property, handle accordingly.
	if !ctx.Has(c.name) {
		return c.matchIfMissing, nil
	}

	// If the expected value is nil, the condition is always true.
	if c.havingValue == nil {
		return true, nil
	}

	havingValue := c.havingValue.(string)

	// Retrieve the property's value and compare it with the expected value.
	val := ctx.Prop(c.name)
	if !strings.HasPrefix(havingValue, "expr:") {
		return val == havingValue, nil
	}

	// Evaluate the expression and return the result.
	ok, err := EvalExpr(havingValue[5:], val)
	if err != nil {
		return false, errutil.WrapError(err, "condition matches error: %s", c)
	}
	return ok, nil
}

func (c *onProperty) String() string {
	var sb strings.Builder
	sb.WriteString("OnProperty(name=")
	sb.WriteString(c.name)
	if c.havingValue != nil {
		sb.WriteString(", havingValue=")
		sb.WriteString(c.havingValue.(string))
	}
	if c.matchIfMissing {
		sb.WriteString(", matchIfMissing")
	}
	sb.WriteString(")")
	return sb.String()
}

/*************************** OnMissingProperty *******************************/

// onMissingProperty is a condition that matches when a specified property is
// absent from the context.
type onMissingProperty struct {
	name string // The name of the property to check for absence.
}

// OnMissingProperty creates a condition that matches if the specified property is missing.
func OnMissingProperty(name string) gs.Condition {
	return &onMissingProperty{name: name}
}

// Matches checks if the condition is met according to the provided context.
func (c *onMissingProperty) Matches(ctx gs.CondContext) (bool, error) {
	return !ctx.Has(c.name), nil
}

func (c *onMissingProperty) String() string {
	return fmt.Sprintf("OnMissingProperty(name=%s)", c.name)
}

/********************************* OnBean ************************************/

// onBean checks for the existence of beans that match a selector.
// It returns true if at least one bean matches the selector, and false otherwise.
type onBean struct {
	s gs.BeanSelector // The selector used to match beans in the context.
}

// OnBean creates a condition that evaluates to true if at least one bean
// matches the specified type and name.
func OnBean[T any](name ...string) gs.Condition {
	return &onBean{s: gs.BeanSelectorFor[T](name...)}
}

// OnBeanSelector creates a condition that evaluates to true if at least one
// bean matches the provided selector.
func OnBeanSelector(s gs.BeanSelector) gs.Condition {
	return &onBean{s: s}
}

// Matches checks if the condition is met according to the provided context.
func (c *onBean) Matches(ctx gs.CondContext) (bool, error) {
	beans, err := ctx.Find(c.s)
	if err != nil {
		return false, errutil.WrapError(err, "condition matches error: %s", c)
	}
	return len(beans) > 0, nil
}

func (c *onBean) String() string {
	return fmt.Sprintf("OnBean(selector=%s)", c.s)
}

/***************************** OnMissingBean *********************************/

// onMissingBean checks for the absence of beans matching a selector.
// It returns true if no beans match the selector, and false otherwise.
type onMissingBean struct {
	s gs.BeanSelector // The selector used to find beans.
}

// OnMissingBean creates a condition that evaluates to true if no beans match
// the specified type and name.
func OnMissingBean[T any](name ...string) gs.Condition {
	return &onMissingBean{s: gs.BeanSelectorFor[T](name...)}
}

// OnMissingBeanSelector creates a condition that evaluates to true if no beans
// match the provided selector.
func OnMissingBeanSelector(s gs.BeanSelector) gs.Condition {
	return &onMissingBean{s: s}
}

// Matches checks if the condition is met according to the provided context.
func (c *onMissingBean) Matches(ctx gs.CondContext) (bool, error) {
	beans, err := ctx.Find(c.s)
	if err != nil {
		return false, errutil.WrapError(err, "condition matches error: %s", c)
	}
	return len(beans) == 0, nil
}

func (c *onMissingBean) String() string {
	return fmt.Sprintf("OnMissingBean(selector=%s)", c.s)
}

/***************************** OnSingleBean **********************************/

// onSingleBean checks if exactly one matching bean exists in the context.
// It returns true if exactly one bean matches the selector, and false otherwise.
type onSingleBean struct {
	s gs.BeanSelector // The selector used to find beans.
}

// OnSingleBean creates a condition that evaluates to true if exactly one bean
// matches the specified type and name.
func OnSingleBean[T any](name ...string) gs.Condition {
	return &onSingleBean{s: gs.BeanSelectorFor[T](name...)}
}

// OnSingleBeanSelector creates a condition that evaluates to true if exactly
// one bean matches the provided selector.
func OnSingleBeanSelector(s gs.BeanSelector) gs.Condition {
	return &onSingleBean{s: s}
}

// Matches checks if the condition is met according to the provided context.
func (c *onSingleBean) Matches(ctx gs.CondContext) (bool, error) {
	beans, err := ctx.Find(c.s)
	if err != nil {
		return false, errutil.WrapError(err, "condition matches error: %s", c)
	}
	return len(beans) == 1, nil
}

func (c *onSingleBean) String() string {
	return fmt.Sprintf("OnSingleBean(selector=%s)", c.s)
}

/***************************** OnExpression **********************************/

// onExpression evaluates a custom expression within the context. The expression should
// return true or false, and the evaluation is expected to happen within the context.
type onExpression struct {
	expression string // The string expression to evaluate.
}

// OnExpression creates a condition that evaluates based on a custom string expression.
// The expression is expected to return true or false.
func OnExpression(expression string) gs.Condition {
	return &onExpression{expression: expression}
}

// Matches checks if the condition is met according to the provided context.
func (c *onExpression) Matches(ctx gs.CondContext) (bool, error) {
	err := util.UnimplementedMethod
	return false, errutil.WrapError(err, "condition matches error: %s", c)
}

func (c *onExpression) String() string {
	return fmt.Sprintf("OnExpression(expression=%s)", c.expression)
}

/********************************** Not ***************************************/

// onNot is a condition that negates another condition. It returns true if the wrapped
// condition evaluates to false, and false if the wrapped condition evaluates to true.
type onNot struct {
	c gs.Condition // The condition to negate.
}

// Not creates a condition that inverts the result of the provided condition.
func Not(c gs.Condition) gs.Condition {
	return &onNot{c: c}
}

// Matches checks if the condition is met according to the provided context.
func (c *onNot) Matches(ctx gs.CondContext) (bool, error) {
	ok, err := c.c.Matches(ctx)
	if err != nil {
		return false, errutil.WrapError(err, "condition matches error: %s", c)
	}
	return !ok, nil
}

func (c *onNot) String() string {
	return fmt.Sprintf("Not(%s)", c.c)
}

/********************************** Or ***************************************/

// onOr is a condition that combines multiple conditions with an OR operator.
// It evaluates to true if at least one condition is satisfied.
type onOr struct {
	conditions []gs.Condition // The list of conditions to evaluate with OR.
}

// Or combines multiple conditions with an OR operator, returning true if at
// least one condition is satisfied.
func Or(conditions ...gs.Condition) gs.Condition {
	if n := len(conditions); n == 0 {
		return nil
	} else if n == 1 {
		return conditions[0]
	}
	return &onOr{conditions: conditions}
}

// Matches checks if the condition is met according to the provided context.
func (g *onOr) Matches(ctx gs.CondContext) (bool, error) {
	for _, c := range g.conditions {
		if ok, err := c.Matches(ctx); err != nil {
			return false, errutil.WrapError(err, "condition matches error: %s", g)
		} else if ok {
			return true, nil
		}
	}
	return false, nil
}

func (g *onOr) String() string {
	return FormatGroup("Or", g.conditions)
}

/********************************* And ***************************************/

// onAnd is a condition that combines multiple conditions with an AND operator.
// It evaluates to true only if all conditions are satisfied.
type onAnd struct {
	conditions []gs.Condition // The list of conditions to evaluate with AND.
}

// And combines multiple conditions with an AND operator, returning true if
// all conditions are satisfied.
func And(conditions ...gs.Condition) gs.Condition {
	if n := len(conditions); n == 0 {
		return nil
	} else if n == 1 {
		return conditions[0]
	}
	return &onAnd{conditions: conditions}
}

// Matches checks if the condition is met according to the provided context.
func (g *onAnd) Matches(ctx gs.CondContext) (bool, error) {
	for _, c := range g.conditions {
		ok, err := c.Matches(ctx)
		if err != nil {
			return false, errutil.WrapError(err, "condition matches error: %s", g)
		} else if !ok {
			return false, nil
		}
	}
	return true, nil
}

func (g *onAnd) String() string {
	return FormatGroup("And", g.conditions)
}

/********************************** None *************************************/

// onNone is a condition that combines multiple conditions with a NONE operator.
// It evaluates to true only if none of the conditions are satisfied.
type onNone struct {
	conditions []gs.Condition // The list of conditions to evaluate with NONE.
}

// None combines multiple conditions with a NONE operator, returning true if
// none of the conditions are satisfied.
func None(conditions ...gs.Condition) gs.Condition {
	if n := len(conditions); n == 0 {
		return nil
	} else if n == 1 {
		return Not(conditions[0])
	}
	return &onNone{conditions: conditions}
}

// Matches checks if the condition is met according to the provided context.
func (g *onNone) Matches(ctx gs.CondContext) (bool, error) {
	for _, c := range g.conditions {
		if ok, err := c.Matches(ctx); err != nil {
			return false, errutil.WrapError(err, "condition matches error: %s", g)
		} else if ok {
			return false, nil
		}
	}
	return true, nil
}

func (g *onNone) String() string {
	return FormatGroup("None", g.conditions)
}

/******************************* utilities ***********************************/

// FormatGroup generates a formatted string for a group of conditions (AND, OR, NONE).
func FormatGroup(op string, conditions []gs.Condition) string {
	var sb strings.Builder
	sb.WriteString(op)
	sb.WriteString("(")
	for i, c := range conditions {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprint(c))
	}
	sb.WriteString(")")
	return sb.String()
}
