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

//go:generate mockgen -build_flags="-mod=mod" -package=cond -source=cond.go -destination=cond_mock.go

// Package gs_cond provides many conditions used when registering bean.
package gs_cond

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/util"
)

type FuncCond func(ctx gs.CondContext) (bool, error)

func (c FuncCond) Matches(ctx gs.CondContext) (bool, error) {
	return c(ctx)
}

// OK returns a Condition that always returns true.
func OK() gs.Condition {
	return FuncCond(func(ctx gs.CondContext) (bool, error) {
		return true, nil
	})
}

// not is a Condition that negating to another.
type not struct {
	c gs.Condition
}

// Not returns a Condition that negating to another.
func Not(c gs.Condition) gs.Condition {
	return &not{c: c}
}

func (c *not) Matches(ctx gs.CondContext) (bool, error) {
	ok, err := c.c.Matches(ctx)
	return !ok, err
}

// onProperty is a Condition that checks a property and its value.
type onProperty struct {
	name           string
	havingValue    string
	matchIfMissing bool
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
	return evalExpr(c.havingValue[5:], getValue(val))
}

// onMissingProperty is a Condition that returns true when a property doesn't exist.
type onMissingProperty struct {
	name string
}

func (c *onMissingProperty) Matches(ctx gs.CondContext) (bool, error) {
	return !ctx.Has(c.name), nil
}

// onBean is a Condition that returns true when finding more than one beans.
type onBean struct {
	selector gs.BeanSelector
}

func (c *onBean) Matches(ctx gs.CondContext) (bool, error) {
	beans, err := ctx.Find(c.selector)
	return len(beans) > 0, err
}

// onMissingBean is a Condition that returns true when finding no beans.
type onMissingBean struct {
	selector gs.BeanSelector
}

func (c *onMissingBean) Matches(ctx gs.CondContext) (bool, error) {
	beans, err := ctx.Find(c.selector)
	return len(beans) == 0, err
}

// onSingleBean is a Condition that returns true when finding only one bean.
type onSingleBean struct {
	selector gs.BeanSelector
}

func (c *onSingleBean) Matches(ctx gs.CondContext) (bool, error) {
	beans, err := ctx.Find(c.selector)
	return len(beans) == 1, err
}

// onExpression is a Condition that returns true when an expression returns true.
type onExpression struct {
	expression string
}

func (c *onExpression) Matches(ctx gs.CondContext) (bool, error) {
	return false, util.UnimplementedMethod
}

// Operator defines operation between conditions, including Or、And、None.
type Operator int

const (
	Or   = Operator(1) // at least one of the conditions must be met.
	And  = Operator(2) // all conditions must be met.
	None = Operator(3) // all conditions must be not met.
)

// group is a Condition implemented by operation of Condition(s).
type group struct {
	op   Operator
	cond []gs.Condition
}

// Group returns a Condition implemented by operation of Condition(s).
func Group(op Operator, cond ...gs.Condition) gs.Condition {
	return &group{op: op, cond: cond}
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

// Matches evaluates the group of conditions based on the specified operator.
// - If the operator is Or, it returns true if at least one condition is satisfied.
// - If the operator is And, it returns true if all conditions are satisfied.
// - If the operator is None, it returns true if none of the conditions are satisfied.
func (g *group) Matches(ctx gs.CondContext) (bool, error) {
	if len(g.cond) == 0 {
		return false, errors.New("no conditions in group")
	}
	switch g.op {
	case Or:
		return g.matchesOr(ctx)
	case And:
		return g.matchesAnd(ctx)
	case None:
		return g.matchesNone(ctx)
	default:
		return false, fmt.Errorf("error condition operator %d", g.op)
	}
}

// node is a Condition implemented by link of Condition(s).
type node struct {
	cond gs.Condition
	op   Operator
	next *node
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
	case Or:
		if ok {
			return ok, nil
		} else {
			return n.next.Matches(ctx)
		}
	case And:
		if ok {
			return n.next.Matches(ctx)
		} else {
			return false, nil
		}
	}

	return false, fmt.Errorf("error condition operator %d", n.op)
}

// Conditional is a Condition implemented by link of Condition(s).
type Conditional struct {
	head *node
	curr *node
}

// New returns a Condition implemented by link of Condition(s).
func New() *Conditional {
	n := &node{}
	return &Conditional{head: n, curr: n}
}

func (c *Conditional) Matches(ctx gs.CondContext) (bool, error) {
	return c.head.Matches(ctx)
}

// Or sets a Or operator.
func (c *Conditional) Or() *Conditional {
	n := &node{}
	c.curr.op = Or
	c.curr.next = n
	c.curr = n
	return c
}

// And sets a And operator.
func (c *Conditional) And() *Conditional {
	n := &node{}
	c.curr.op = And
	c.curr.next = n
	c.curr = n
	return c
}

// On returns a Conditional that starts with one Condition.
func On(cond gs.Condition) *Conditional {
	return New().On(cond)
}

// On adds one Condition.
func (c *Conditional) On(cond gs.Condition) *Conditional {
	if c.curr.cond != nil {
		c.And()
	}
	c.curr.cond = cond
	return c
}

type PropertyOption func(*onProperty)

// MatchIfMissing sets a Condition to return true when property doesn't exist.
func MatchIfMissing() PropertyOption {
	return func(c *onProperty) {
		c.matchIfMissing = true
	}
}

// HavingValue sets a Condition to return true when property value equals to havingValue.
func HavingValue(havingValue string) PropertyOption {
	return func(c *onProperty) {
		c.havingValue = havingValue
	}
}

// OnProperty returns a Conditional that starts with a Condition that checks a property
// and its value.
func OnProperty(name string, options ...PropertyOption) *Conditional {
	return New().OnProperty(name, options...)
}

// OnProperty adds a Condition that checks a property and its value.
func (c *Conditional) OnProperty(name string, options ...PropertyOption) *Conditional {
	cond := &onProperty{name: name}
	for _, option := range options {
		option(cond)
	}
	return c.On(cond)
}

// OnMissingProperty returns a Conditional that starts with a Condition that returns
// true when property doesn't exist.
func OnMissingProperty(name string) *Conditional {
	return New().OnMissingProperty(name)
}

// OnMissingProperty adds a Condition that returns true when property doesn't exist.
func (c *Conditional) OnMissingProperty(name string) *Conditional {
	return c.On(&onMissingProperty{name: name})
}

// OnBean returns a Conditional that starts with a Condition that returns true when
// finding more than one beans.
func OnBean(selector gs.BeanSelector) *Conditional {
	return New().OnBean(selector)
}

// OnBean adds a Condition that returns true when finding more than one beans.
func (c *Conditional) OnBean(selector gs.BeanSelector) *Conditional {
	return c.On(&onBean{selector: selector})
}

// OnMissingBean returns a Conditional that starts with a Condition that returns
// true when finding no beans.
func OnMissingBean(selector gs.BeanSelector) *Conditional {
	return New().OnMissingBean(selector)
}

// OnMissingBean adds a Condition that returns true when finding no beans.
func (c *Conditional) OnMissingBean(selector gs.BeanSelector) *Conditional {
	return c.On(&onMissingBean{selector: selector})
}

// OnSingleBean returns a Conditional that starts with a Condition that returns
// true when finding only one bean.
func OnSingleBean(selector gs.BeanSelector) *Conditional {
	return New().OnSingleBean(selector)
}

// OnSingleBean adds a Condition that returns true when finding only one bean.
func (c *Conditional) OnSingleBean(selector gs.BeanSelector) *Conditional {
	return c.On(&onSingleBean{selector: selector})
}

// OnExpression returns a Conditional that starts with a Condition that returns
// true when an expression returns true.
func OnExpression(expression string) *Conditional {
	return New().OnExpression(expression)
}

// OnExpression adds a Condition that returns true when an expression returns true.
func (c *Conditional) OnExpression(expression string) *Conditional {
	return c.On(&onExpression{expression: expression})
}

// OnMatches returns a Conditional that starts with a Condition that returns true
// when function returns true.
func OnMatches(fn func(ctx gs.CondContext) (bool, error)) *Conditional {
	return New().OnMatches(fn)
}

// OnMatches adds a Condition that returns true when function returns true.
func (c *Conditional) OnMatches(fn func(ctx gs.CondContext) (bool, error)) *Conditional {
	return c.On(FuncCond(fn))
}

// OnProfile returns a Conditional that starts with a Condition that returns true
// when property value equals to profile.
func OnProfile(profile string) *Conditional {
	return New().OnProfile(profile)
}

// OnProfile adds a Condition that returns true when property value equals to profile.
func (c *Conditional) OnProfile(profile string) *Conditional {
	return c.OnProperty("spring.profiles.active", HavingValue(profile))
}
