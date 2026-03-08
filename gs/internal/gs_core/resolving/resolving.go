/*
 * Copyright 2025 The Go-Spring Authors.
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

package resolving

import (
	"reflect"
	"regexp"
	"slices"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/gs/internal/gs_init"
	"github.com/go-spring/stdlib/errutil"
	"github.com/go-spring/stdlib/flatten"
	"github.com/go-spring/stdlib/funcutil"
)

// RefreshState represents the current state of the container.
type RefreshState int

const (
	RefreshDefault = RefreshState(iota)
	RefreshPrepare
	Refreshing
	Refreshed
)

// Resolving is the core container responsible for holding bean definitions,
// processing modules, scanning configuration beans, and
// resolving beans against conditions.
type Resolving struct {
	state RefreshState              // current refresh state
	beans []*gs_bean.BeanDefinition // all beans managed by the container
}

// New creates an empty Resolving instance.
func New() *Resolving {
	return &Resolving{}
}

// Beans returns all active bean definitions, excluding deleted ones.
func (c *Resolving) Beans() []*gs_bean.BeanDefinition {
	var beans []*gs_bean.BeanDefinition
	for _, b := range c.beans {
		if b.Status() == gs_bean.StatusDeleted {
			continue
		}
		beans = append(beans, b)
	}
	return beans
}

// Provide registers a bean definition.
// It accepts either an existing instance or a constructor function.
func (c *Resolving) Provide(objOrCtor any, args ...gs.Arg) *gs_bean.BeanDefinition {
	if c.state >= Refreshing {
		panic("container is already refreshing or refreshed")
	}
	b := gs_bean.NewBean(objOrCtor, args...)
	c.beans = append(c.beans, b)
	return b.Caller(2)
}

// Refresh performs the full lifecycle of container initialization.
// The phases are as follows:
//  1. Apply registered modules to register additional beans.
//  2. Scan configuration beans and register methods as beans.
//  4. Resolve conditions for all beans and mark inactive ones as deleted.
//  5. Check for duplicate beans (by type and name).
//  6. Validate that all root beans are resolved and ready to wire.
func (c *Resolving) Refresh(p flatten.Storage) error {
	if c.state != RefreshDefault {
		return errutil.Explain(nil, "container is already refreshing or refreshed")
	}
	c.state = RefreshPrepare

	c.beans = append(gs_init.Beans(), c.beans...)
	if err := c.applyModules(p); err != nil {
		return err
	}

	c.state = Refreshing

	if err := c.scanConfigurations(); err != nil {
		return err
	}

	if err := c.resolveBeans(p); err != nil {
		return err
	}

	if err := c.checkDuplicateBeans(); err != nil {
		return err
	}

	c.state = Refreshed
	return nil
}

// applyModules executes all registered modules that match their conditions.
func (c *Resolving) applyModules(p flatten.Storage) error {
	ctx := &ConditionContext{p: p, c: c}
	for _, m := range gs_init.Modules() {
		if m.Condition != nil {
			if ok, err := m.Condition.Matches(ctx); err != nil {
				return err
			} else if !ok {
				continue
			}
		}
		if err := m.ModuleFunc(c, p); err != nil {
			return errutil.Explain(err, "apply module error")
		}
	}
	return nil
}

// scanConfigurations iterates over all beans that represent configuration
// objects and scans their methods to register additional beans.
func (c *Resolving) scanConfigurations() error {
	for _, b := range c.beans {
		if b.GetConfiguration() == nil {
			continue
		}
		beans, err := c.scanConfiguration(b)
		if err != nil {
			return errutil.Explain(err, "scan configuration error")
		}
		c.beans = append(c.beans, beans...)
	}
	return nil
}

// scanConfiguration inspects methods of a configuration bean and registers
// methods as beans if they match the inclusion/exclusion patterns.
// By default, include methods named like "NewXxx"
func (c *Resolving) scanConfiguration(bd *gs_bean.BeanDefinition) ([]*gs_bean.BeanDefinition, error) {
	var (
		includes []*regexp.Regexp
		excludes []*regexp.Regexp
	)

	param := bd.GetConfiguration()
	ss := param.Includes
	if len(ss) == 0 {
		ss = []string{"New.*"}
	}
	for _, s := range ss {
		p, err := regexp.Compile(s)
		if err != nil {
			return nil, errutil.Explain(err, "invalid regexp '%s'", s)
		}
		includes = append(includes, p)
	}

	ss = param.Excludes
	for _, s := range ss {
		p, err := regexp.Compile(s)
		if err != nil {
			return nil, errutil.Explain(err, "invalid regexp '%s'", s)
		}
		excludes = append(excludes, p)
	}

	var ret []*gs_bean.BeanDefinition
	n := bd.GetType().NumMethod()
	for i := range n {
		m := bd.GetType().Method(i)

		// Skip methods matching any exclusion pattern.
		skip := false
		for _, p := range excludes {
			if p.MatchString(m.Name) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		// Register method as a bean if it matches inclusion pattern.
		for _, p := range includes {
			if !p.MatchString(m.Name) {
				continue
			}
			b := gs_bean.NewBean(m.Func.Interface(), bd).
				Name(bd.GetName() + "_" + m.Name).
				Condition(gs_cond.OnBeanID(bd.BeanID()))
			file, line, _ := funcutil.FileLine(m.Func.Interface())
			b.SetFileLine(file, line)
			ret = append(ret, b)
			break
		}
	}
	return ret, nil
}

// isBeanMatched checks whether a bean matches the given type and name selector.
func isBeanMatched(t reflect.Type, s string, b *gs_bean.BeanDefinition) bool {
	if s != "" && s != b.GetName() {
		return false
	}
	if t != nil && t != b.GetType() {
		if !slices.Contains(b.Exports(), t) {
			return false
		}
	}
	return true
}

// resolveBeans iterates over all beans and resolves their conditions,
// marking them as resolved or deleted.
func (c *Resolving) resolveBeans(p flatten.Storage) error {
	ctx := &ConditionContext{p: p, c: c}
	for _, b := range c.beans {
		if err := ctx.resolveBean(b); err != nil {
			return errutil.Explain(err, "resolve bean error")
		}
	}
	return nil
}

// checkDuplicateBeans ensures that no two beans share the same type and name.
func (c *Resolving) checkDuplicateBeans() error {
	beansByID := make(map[gs.BeanID]*gs_bean.BeanDefinition)
	for _, b := range c.beans {
		if b.Status() == gs_bean.StatusDeleted {
			continue
		}
		for _, t := range append(b.Exports(), b.GetType()) {
			beanID := gs.BeanID{Name: b.GetName(), Type: t}
			if d, ok := beansByID[beanID]; ok {
				return errutil.Explain(nil, "found duplicate beans [%s] [%s]", b, d)
			}
			beansByID[beanID] = b
		}
	}
	return nil
}

// ConditionContext provides an evaluation context for conditions
// during bean resolution.
type ConditionContext struct {
	c *Resolving
	p flatten.Storage
}

// resolveBean evaluates a bean's conditions and updates its status accordingly.
// If any condition fails, the bean is marked as deleted.
func (c *ConditionContext) resolveBean(b *gs_bean.BeanDefinition) error {
	if b.Status() >= gs_bean.StatusResolving {
		return nil
	}
	b.SetStatus(gs_bean.StatusResolving)
	for _, cond := range b.Conditions() {
		if ok, err := cond.Matches(c); err != nil {
			return err
		} else if !ok {
			b.SetStatus(gs_bean.StatusDeleted)
			return nil
		}
	}
	b.SetStatus(gs_bean.StatusResolved)
	return nil
}

// Has checks if a configuration property exists.
func (c *ConditionContext) Has(key string) bool {
	return c.p.Exists(key)
}

// Prop retrieves a configuration property by key,
// optionally returning a default value if the key is not found.
func (c *ConditionContext) Prop(key string, def ...string) string {
	str, ok := c.p.Value(key)
	if ok {
		return str
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

// Find returns all beans that match the provided selector
// and are successfully resolved (active).
func (c *ConditionContext) Find(beanID gs.BeanID) ([]gs.ConditionBean, error) {
	var found []gs.ConditionBean
	for _, b := range c.c.beans {
		if b.Status() == gs_bean.StatusResolving || b.Status() == gs_bean.StatusDeleted {
			continue
		}
		if !isBeanMatched(beanID.Type, beanID.Name, b) {
			continue
		}
		if err := c.resolveBean(b); err != nil {
			return nil, err
		}
		if b.Status() == gs_bean.StatusDeleted {
			continue
		}
		found = append(found, b)
	}
	return found, nil
}
