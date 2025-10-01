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

	"github.com/go-spring/spring-base/util"
	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
)

// RefreshState represents the current state of the container.
type RefreshState int

const (
	RefreshDefault = RefreshState(iota)
	RefreshPrepare
	Refreshing
	Refreshed
)

// Module represents a module that can register additional beans
// when certain conditions are met.
type Module struct {
	f func(p conf.Properties) error
	c gs.Condition
}

// Resolving is the core container responsible for holding bean definitions,
// processing modules, applying mocks, scanning configuration beans, and
// resolving beans against conditions.
type Resolving struct {
	state   RefreshState              // current refresh state
	mocks   []gs.BeanMock             // registered mocks to override beans
	beans   []*gs_bean.BeanDefinition // all beans managed by the container
	roots   []*gs_bean.BeanDefinition // root beans to wire at the end
	modules []Module                  // registered modules
}

// New creates an empty Resolving instance.
func New() *Resolving {
	return &Resolving{}
}

// Roots returns all root beans.
func (c *Resolving) Roots() []*gs_bean.BeanDefinition {
	return c.roots
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

// AddMock registers a mock bean which can override an existing bean
// during the refresh phase.
func (c *Resolving) AddMock(mock gs.BeanMock) {
	c.mocks = append(c.mocks, mock)
}

// Object registers a pre-constructed instance as a bean definition.
func (c *Resolving) Object(i any) *gs.RegisteredBean {
	b := gs_bean.NewBean(reflect.ValueOf(i))
	return c.Register(b).Caller(1)
}

// Provide registers a constructor function and optional arguments as a bean.
func (c *Resolving) Provide(ctor any, args ...gs.Arg) *gs.RegisteredBean {
	b := gs_bean.NewBean(ctor, args...)
	return c.Register(b).Caller(1)
}

// Register adds a bean definition to the container.
// It must be called before the container starts refreshing.
func (c *Resolving) Register(b *gs.BeanDefinition) *gs.RegisteredBean {
	if c.state >= Refreshing {
		panic("container is refreshing or already refreshed")
	}
	bd := b.BeanRegistration().(*gs_bean.BeanDefinition)
	c.beans = append(c.beans, bd)
	return gs.NewRegisteredBean(bd)
}

// Module registers a conditional module that will be executed
// to add beans before the container starts refreshing.
func (c *Resolving) Module(conditions []gs_cond.ConditionOnProperty, fn func(p conf.Properties) error) {
	var arr []gs.Condition
	for _, cond := range conditions {
		arr = append(arr, cond)
	}
	c.modules = append(c.modules, Module{
		f: fn,
		c: gs_cond.And(arr...),
	})
}

// Root marks a registered bean as a root bean.
func (c *Resolving) Root(b *gs.RegisteredBean) {
	bd := b.BeanRegistration().(*gs_bean.BeanDefinition)
	c.roots = append(c.roots, bd)
}

// Refresh performs the full lifecycle of container initialization.
// The phases are as follows:
//  1. Apply registered modules to register additional beans.
//  2. Scan configuration beans and register methods as beans.
//  3. Apply mock beans to override specific target beans.
//  4. Resolve conditions for all beans and mark inactive ones as deleted.
//  5. Check for duplicate beans (by type and name).
//  6. Validate that all root beans are resolved and ready to wire.
func (c *Resolving) Refresh(p conf.Properties) error {
	if c.state != RefreshDefault {
		return util.FormatError(nil, "container is already refreshing or refreshed")
	}
	c.state = RefreshPrepare

	if err := c.applyModules(p); err != nil {
		return err
	}

	c.state = Refreshing

	if err := c.scanConfigurations(); err != nil {
		return err
	}

	if err := c.applyMocks(); err != nil {
		return err
	}

	if err := c.resolveBeans(p); err != nil {
		return err
	}

	if err := c.checkDuplicateBeans(); err != nil {
		return err
	}

	for _, b := range c.roots {
		if b.Status() == gs_bean.StatusDeleted {
			continue
		}
		if b.Status() != gs_bean.StatusResolved {
			return util.FormatError(nil, "bean %q status is invalid for wiring", b)
		}
	}

	c.state = Refreshed
	return nil
}

// applyModules executes all registered modules that match their conditions.
func (c *Resolving) applyModules(p conf.Properties) error {
	ctx := &ConditionContext{p: p, c: c}
	for _, m := range c.modules {
		if m.c != nil {
			if ok, err := m.c.Matches(ctx); err != nil {
				return err
			} else if !ok {
				continue
			}
		}
		if err := m.f(p); err != nil {
			return util.FormatError(err, "apply module error")
		}
	}
	return nil
}

// scanConfigurations iterates over all beans that represent configuration
// objects and scans their methods to register additional beans.
func (c *Resolving) scanConfigurations() error {
	for _, b := range c.beans {
		if b.Configuration() == nil {
			continue
		}

		// First, check if a mock is defined for this configuration bean.
		var foundMocks []gs.BeanMock
		for _, mock := range c.mocks {
			t, s := mock.Target.TypeAndName()
			if s != "" && s != b.Name() {
				continue
			}
			if t != b.Type() {
				continue
			}
			foundMocks = append(foundMocks, mock)
		}
		if n := len(foundMocks); n > 1 {
			return util.FormatError(nil, "found duplicate mock bean for '%s'", b.Name())
		} else if n == 1 {
			b.SetMock(foundMocks[0].Object)
			continue
		}

		// If not mocked, scan configuration methods.
		beans, err := c.scanConfiguration(b)
		if err != nil {
			return util.FormatError(err, "scan configuration error")
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

	param := bd.Configuration()
	ss := param.Includes
	if len(ss) == 0 {
		ss = []string{"New.*"}
	}
	for _, s := range ss {
		p, err := regexp.Compile(s)
		if err != nil {
			return nil, util.FormatError(err, "invalid regexp '%s'", s)
		}
		includes = append(includes, p)
	}

	ss = param.Excludes
	for _, s := range ss {
		p, err := regexp.Compile(s)
		if err != nil {
			return nil, util.FormatError(err, "invalid regexp '%s'", s)
		}
		excludes = append(excludes, p)
	}

	var ret []*gs_bean.BeanDefinition
	n := bd.Type().NumMethod()
	for i := range n {
		m := bd.Type().Method(i)

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
			b := gs_bean.NewBean(m.Func.Interface(), gs.NewBeanDefinition(bd)).
				Name(bd.Name() + "_" + m.Name).
				Condition(gs_cond.OnBeanSelector(bd)).
				BeanRegistration().(*gs_bean.BeanDefinition)
			file, line, _ := util.FileLine(m.Func.Interface())
			b.SetFileLine(file, line)
			ret = append(ret, b)
			break
		}
	}
	return ret, nil
}

// isBeanMatched checks whether a bean matches the given type and name selector.
func isBeanMatched(t reflect.Type, s string, b *gs_bean.BeanDefinition) bool {
	if s != "" && s != b.Name() {
		return false
	}
	if t != nil && t != b.Type() {
		if !slices.Contains(b.Exports(), t) {
			return false
		}
	}
	return true
}

// applyMocks iterates over all registered mocks and applies them to matching beans.
func (c *Resolving) applyMocks() error {
	for _, mock := range c.mocks {
		if err := c.applyMock(mock); err != nil {
			return err
		}
	}
	return nil
}

// applyMock applies a mock object to a target bean.
// It ensures the mock implements all exported interfaces of the target bean.
// If more than one target bean is found or the mock is invalid, an error is returned.
func (c *Resolving) applyMock(mock gs.BeanMock) error {
	var foundBeans []*gs_bean.BeanDefinition
	vt := reflect.TypeOf(mock.Object)
	t, s := mock.Target.TypeAndName()

	for _, b := range c.beans {
		if !isBeanMatched(t, s, b) {
			continue
		}
		// Verify mock implements all exported interfaces
		for _, et := range b.Exports() {
			if !vt.Implements(et) {
				return util.FormatError(nil, "mock %T does not implement required interface %v", mock.Object, et)
			}
		}
		foundBeans = append(foundBeans, b)
	}
	if len(foundBeans) == 0 {
		return nil
	}
	if len(foundBeans) > 1 {
		return util.FormatError(nil, "found duplicate mocked beans")
	}
	foundBeans[0].SetMock(mock.Object)
	return nil
}

// resolveBeans iterates over all beans and resolves their conditions,
// marking them as resolved or deleted.
func (c *Resolving) resolveBeans(p conf.Properties) error {
	ctx := &ConditionContext{p: p, c: c}
	for _, b := range c.beans {
		if err := ctx.resolveBean(b); err != nil {
			return util.FormatError(err, "resolve bean error")
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
		for _, t := range append(b.Exports(), b.Type()) {
			beanID := gs.BeanID{Name: b.Name(), Type: t}
			if d, ok := beansByID[beanID]; ok {
				return util.FormatError(nil, "found duplicate beans [%s] [%s]", b, d)
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
	p conf.Properties
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
	return c.p.Has(key)
}

// Prop retrieves a configuration property by key,
// optionally returning a default value if the key is not found.
func (c *ConditionContext) Prop(key string, def ...string) string {
	return c.p.Get(key, def...)
}

// Find returns all beans that match the provided selector
// and are successfully resolved (active).
func (c *ConditionContext) Find(s gs.BeanSelector) ([]gs.ConditionBean, error) {
	var found []gs.ConditionBean
	t, name := s.TypeAndName()
	for _, b := range c.c.beans {
		if b.Status() == gs_bean.StatusResolving || b.Status() == gs_bean.StatusDeleted {
			continue
		}
		if !isBeanMatched(t, name, b) {
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
