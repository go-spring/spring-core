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
// in a flexible, modular, and reusable manner. The core concept revolves
// around the [gs.Condition] interface, which determines whether a given
// context satisfies specific criteria. Various implementations of
// [gs.Condition] allow combining basic and complex logic operations, such
// as `And`, `Or`, and `Not`. These are further enhanced with support for
// property-based conditions, bean existence checks, and custom evaluation logic.
package gs_cond

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/util"
)

// FuncCond is a functional implementation of the [gs.Condition] interface.
// It evaluates a condition using the provided function.
type FuncCond func(ctx gs.CondContext) (bool, error)

func (c FuncCond) Matches(ctx gs.CondContext) (bool, error) {
	return c(ctx)
}

// OK returns a [gs.Condition] that always evaluates to true.
func OK() gs.Condition {
	return FuncCond(func(ctx gs.CondContext) (bool, error) {
		return true, nil
	})
}

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
	return !ok, err
}

// onProperty evaluates a condition based on the existence and value of a property.
// - If the property is missing, the result is determined by `matchIfMissing`.
// - If `havingValue` is provided, the property's value must match it.
// - If `havingValue` starts with "expr:", the value is evaluated as an expression.
type onProperty struct {
	name           string // Name of the property to check.
	havingValue    string // Expected value or expression.
	matchIfMissing bool   // Result if the property is missing.
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
	return EvalExpr(c.havingValue[5:], getValue(val))
}

// onMissingProperty evaluates to true when a specified property is absent in the context.
type onMissingProperty struct {
	name string
}

func (c *onMissingProperty) Matches(ctx gs.CondContext) (bool, error) {
	return !ctx.Has(c.name), nil
}

// onBean checks for the existence of beans matching a selector.
// It evaluates to true if at least one bean matches.
type onBean struct {
	selector gs.BeanSelector
}

func (c *onBean) Matches(ctx gs.CondContext) (bool, error) {
	beans, err := ctx.Find(c.selector)
	return len(beans) > 0, err
}

// onMissingBean evaluates to true if no beans match the selector.
type onMissingBean struct {
	selector gs.BeanSelector
}

func (c *onMissingBean) Matches(ctx gs.CondContext) (bool, error) {
	beans, err := ctx.Find(c.selector)
	return len(beans) == 0, err
}

// onSingleBean checks for exactly one matching bean.
type onSingleBean struct {
	selector gs.BeanSelector
}

func (c *onSingleBean) Matches(ctx gs.CondContext) (bool, error) {
	beans, err := ctx.Find(c.selector)
	return len(beans) == 1, err
}

// onExpression evaluates a custom expression within the context.
type onExpression struct {
	expression string
}

func (c *onExpression) Matches(ctx gs.CondContext) (bool, error) {
	return false, util.UnimplementedMethod
}

/******************************** group **************************************/

// Operator defines logical operations between conditions.
// Supported operators include:
// - opOr: At least one condition must be satisfied.
// - opAnd: All conditions must be satisfied.
// - opNone: No condition must be satisfied.
type Operator int

const (
	opOr   = Operator(iota) // Logical OR operation.
	opAnd                   // Logical AND operation.
	opNone                  // Logical NONE (NOT ANY) operation.
)

// group represents a composite condition formed by combining multiple conditions
// using a specified logical operator.
type group struct {
	op   Operator       // Logical operator (Or, And, None).
	cond []gs.Condition // List of conditions to evaluate.
}

// Or combines conditions with an OR operator. Returns a condition that
// evaluates to true if at least one condition is satisfied.
func Or(cond ...gs.Condition) gs.Condition {
	if n := len(cond); n == 0 {
		return OK()
	} else if n == 1 {
		return cond[0]
	}
	return &group{op: opOr, cond: cond}
}

// And combines conditions with an AND operator. Returns a condition that
// evaluates to true only if all conditions are satisfied.
func And(cond ...gs.Condition) gs.Condition {
	if n := len(cond); n == 0 {
		return OK()
	} else if n == 1 {
		return cond[0]
	}
	return &group{op: opAnd, cond: cond}
}

// None combines conditions with a NONE operator. Returns a condition that
// evaluates to true only if none of the conditions are satisfied.
func None(cond ...gs.Condition) gs.Condition {
	if n := len(cond); n == 0 {
		return OK()
	} else if n == 1 {
		return Not(cond[0])
	}
	return &group{op: opNone, cond: cond}
}

func (g *group) matchesOr(ctx gs.CondContext) (bool, error) {
	for _, c := range g.cond {
		if ok, err := c.Matches(ctx); err != nil {
			return false, err
		} else if ok {
			return true, nil
		}
	}
	return false, nil
}

func (g *group) matchesAnd(ctx gs.CondContext) (bool, error) {
	for _, c := range g.cond {
		if ok, err := c.Matches(ctx); err != nil {
			return false, err
		} else if !ok {
			return false, nil
		}
	}
	return true, nil
}

func (g *group) matchesNone(ctx gs.CondContext) (bool, error) {
	for _, c := range g.cond {
		if ok, err := c.Matches(ctx); err != nil {
			return false, err
		} else if ok {
			return false, nil
		}
	}
	return true, nil
}

func (g *group) Matches(ctx gs.CondContext) (bool, error) {
	if len(g.cond) == 0 {
		return false, errors.New("no conditions in group")
	}
	switch g.op {
	case opOr:
		return g.matchesOr(ctx)
	case opAnd:
		return g.matchesAnd(ctx)
	case opNone:
		return g.matchesNone(ctx)
	default:
		return false, fmt.Errorf("error condition operator %d", g.op)
	}
}

/****************************** Conditional **********************************/

// node represents a single node in a linked structure of conditions. Each node
// contains a condition, a logical operator, and a reference to the next node.
type node struct {
	cond gs.Condition // The condition to evaluate.
	op   Operator     // Logical operator to the next node (Or, And).
	next *node        // Reference to the next node.
}

func (n *node) Matches(ctx gs.CondContext) (bool, error) {

	if n.cond == nil {
		return true, nil
	}

	ok, err := n.cond.Matches(ctx)
	if err != nil {
		return false, err
	}

	if n.next == nil {
		return ok, nil
	} else if n.next.cond == nil {
		return false, errors.New("no condition in last node")
	}

	switch n.op {
	case opOr:
		if ok {
			return ok, nil
		} else {
			return n.next.Matches(ctx)
		}
	case opAnd:
		if ok {
			return n.next.Matches(ctx)
		} else {
			return false, nil
		}
	}

	return false, fmt.Errorf("error condition operator %d", n.op)
}

// Conditional provides a chainable structure for combining conditions
// in a linked list format.
type Conditional struct {
	head *node
	curr *node
}

// New initializes a new Conditional structure.
func New() *Conditional {
	n := &node{}
	return &Conditional{head: n, curr: n}
}

func (c *Conditional) Matches(ctx gs.CondContext) (bool, error) {
	return c.head.Matches(ctx)
}

// Or sets the logical operator to OR for the current node and creates a new node.
func (c *Conditional) Or() *Conditional {
	n := &node{}
	c.curr.op = opOr
	c.curr.next = n
	c.curr = n
	return c
}

// And sets the logical operator to AND for the current node and creates a new node.
func (c *Conditional) And() *Conditional {
	n := &node{}
	c.curr.op = opAnd
	c.curr.next = n
	c.curr = n
	return c
}

// On adds a new condition to the current node. If the current node already
// contains a condition, it implicitly starts a new AND condition.
func On(cond gs.Condition) *Conditional {
	return New().On(cond)
}

func (c *Conditional) On(cond gs.Condition) *Conditional {
	if c.curr.cond != nil {
		c.And()
	}
	c.curr.cond = cond
	return c
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
func OnProperty(name string, options ...PropertyOption) *Conditional {
	return New().OnProperty(name, options...)
}

func (c *Conditional) OnProperty(name string, options ...PropertyOption) *Conditional {
	cond := &onProperty{name: name}
	for _, option := range options {
		option(cond)
	}
	return c.On(cond)
}

// OnMissingProperty creates a Conditional that starts with a condition
// that matches if a property does not exist.
func OnMissingProperty(name string) *Conditional {
	return New().OnMissingProperty(name)
}

func (c *Conditional) OnMissingProperty(name string) *Conditional {
	return c.On(&onMissingProperty{name: name})
}

// OnBean creates a Conditional that starts with a condition to match
// when more than one bean is found matching the provided selector.
func OnBean(selector gs.BeanSelector) *Conditional {
	return New().OnBean(selector)
}

func (c *Conditional) OnBean(selector gs.BeanSelector) *Conditional {
	return c.On(&onBean{selector: selector})
}

// OnMissingBean creates a Conditional that starts with a condition
// to match when no beans are found matching the provided selector.
func OnMissingBean(selector gs.BeanSelector) *Conditional {
	return New().OnMissingBean(selector)
}

func (c *Conditional) OnMissingBean(selector gs.BeanSelector) *Conditional {
	return c.On(&onMissingBean{selector: selector})
}

// OnSingleBean creates a Conditional that starts with a condition
// to match when exactly one bean is found matching the provided selector.
func OnSingleBean(selector gs.BeanSelector) *Conditional {
	return New().OnSingleBean(selector)
}

func (c *Conditional) OnSingleBean(selector gs.BeanSelector) *Conditional {
	return c.On(&onSingleBean{selector: selector})
}

// OnExpression creates a Conditional that starts with a condition
// to match based on the evaluation of a string expression. The expression
// should return true or false.
func OnExpression(expression string) *Conditional {
	return New().OnExpression(expression)
}

func (c *Conditional) OnExpression(expression string) *Conditional {
	return c.On(&onExpression{expression: expression})
}

// OnMatches creates a Conditional that starts with a condition to
// match based on a custom function. The function takes a CondContext
// and returns a boolean value and an optional error.
func OnMatches(fn func(ctx gs.CondContext) (bool, error)) *Conditional {
	return New().OnMatches(fn)
}

func (c *Conditional) OnMatches(fn func(ctx gs.CondContext) (bool, error)) *Conditional {
	return c.On(FuncCond(fn))
}

// OnProfile creates a Conditional that starts with a condition
// to match when the property `spring.profiles.active` equals
// the provided profile value.
func OnProfile(profile string) *Conditional {
	return New().OnProfile(profile)
}

func (c *Conditional) OnProfile(profile string) *Conditional {
	return c.OnProperty("spring.profiles.active", HavingValue(profile))
}
