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
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"slices"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/util"
)

// RefreshState represents the current state of the container.
type RefreshState int

const (
	RefreshDefault = RefreshState(iota)
	Refreshing
	Refreshed
)

// BeanGroupFunc defines a function that dynamically registers beans
// based on configuration properties.
type BeanGroupFunc func(p conf.Properties) ([]*gs.BeanDefinition, error)

// BeanMock defines a mock object and its target bean selector for overriding.
type BeanMock struct {
	Object interface{}     // Mock instance to replace the target bean
	Target gs.BeanSelector // Selector to identify the target bean
}

// Resolving manages bean definitions, mocks, and dynamic bean registration functions.
type Resolving struct {
	state RefreshState              // Current refresh state
	mocks []BeanMock                // Registered mock beans
	beans []*gs_bean.BeanDefinition // Managed bean definitions
	funcs []BeanGroupFunc           // Dynamic bean registration functions
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

// Mock registers a mock object to override a bean matching the selector.
func (c *Resolving) Mock(obj interface{}, target gs.BeanSelector) {
	mock := BeanMock{Object: obj, Target: target}
	c.mocks = append(c.mocks, mock)
}

// Object registers a pre-constructed instance as a bean.
func (c *Resolving) Object(i interface{}) *gs.RegisteredBean {
	b := gs_bean.NewBean(reflect.ValueOf(i))
	return c.Register(b).Caller(1)
}

// Provide registers a constructor function to create a bean.
func (c *Resolving) Provide(ctor interface{}, args ...gs.Arg) *gs.RegisteredBean {
	b := gs_bean.NewBean(ctor, args...)
	return c.Register(b).Caller(1)
}

// Register adds a bean definition to the container.
func (c *Resolving) Register(b *gs.BeanDefinition) *gs.RegisteredBean {
	if c.state >= Refreshing {
		panic("container is refreshing or already refreshed")
	}
	bd := b.BeanRegistration().(*gs_bean.BeanDefinition)
	c.beans = append(c.beans, bd)
	return gs.NewRegisteredBean(bd)
}

// GroupRegister adds a function to dynamically register beans.
func (c *Resolving) GroupRegister(fn BeanGroupFunc) {
	c.funcs = append(c.funcs, fn)
}

// Refresh performs the full initialization process of the container.
// It transitions through several phases:
// - Executes group functions to register additional beans.
// - Scans configuration beans and registers their methods as beans.
// - Applies mock beans to override specific targets.
// - Resolves all beans based on their conditions.
// - Validates that no duplicate beans exist.
func (c *Resolving) Refresh(p conf.Properties) error {
	if c.state != RefreshDefault {
		return errors.New("container is already refreshing or refreshed")
	}
	c.state = Refreshing

	if err := c.applyGroupFuncs(p); err != nil {
		return err
	}

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

	c.state = Refreshed
	return nil
}

// applyGroupFuncs executes registered group functions to add dynamic beans.
func (c *Resolving) applyGroupFuncs(p conf.Properties) error {
	for _, fn := range c.funcs {
		beans, err := fn(p)
		if err != nil {
			return err
		}
		for _, b := range beans {
			d := b.BeanRegistration().(*gs_bean.BeanDefinition)
			c.beans = append(c.beans, d)
		}
	}
	return nil
}

// scanConfigurations processes configuration beans to register their methods as beans.
func (c *Resolving) scanConfigurations() error {
	for _, b := range c.beans {
		if b.Configuration() == nil {
			continue
		}
		// Check if the configuration bean has a mock override
		var foundMocks []BeanMock
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
			return fmt.Errorf("found duplicate mock bean for '%s'", b.Name())
		} else if n == 1 {
			b.SetMock(foundMocks[0].Object)
			continue
		}
		// Scan methods if no mock is applied
		beans, err := c.scanConfiguration(b)
		if err != nil {
			return err
		}
		c.beans = append(c.beans, beans...)
	}
	return nil
}

// scanConfiguration inspects the methods of a configuration bean, and for each
// method that matches the include patterns and not the exclude patterns,
// registers it as a bean. This enables dynamic bean registration based on method
// naming conventions or regex.
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
			return nil, err
		}
		includes = append(includes, p)
	}

	ss = param.Excludes
	for _, s := range ss {
		p, err := regexp.Compile(s)
		if err != nil {
			return nil, err
		}
		excludes = append(excludes, p)
	}

	var ret []*gs_bean.BeanDefinition
	n := bd.Type().NumMethod()
	for i := 0; i < n; i++ {
		m := bd.Type().Method(i)
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

// isBeanMatched checks if a bean matches the target type and name selector.
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

// applyMocks overrides target beans with registered mock objects.
func (c *Resolving) applyMocks() error {
	for _, mock := range c.mocks {
		if err := c.applyMock(mock); err != nil {
			return err
		}
	}
	return nil
}

// applyMock applies a mock object to its target bean. It ensures that the mock
// implements all the interfaces that the original bean exported. If multiple
// matching beans are found, or if the mock doesn't implement required interfaces,
// an error is returned.
func (c *Resolving) applyMock(mock BeanMock) error {
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
				return fmt.Errorf("found unimplemented interface")
			}
		}
		foundBeans = append(foundBeans, b)
	}
	if len(foundBeans) == 0 {
		return nil
	}
	if len(foundBeans) > 1 {
		return fmt.Errorf("found duplicate mocked beans")
	}
	foundBeans[0].SetMock(mock.Object)
	return nil
}

// resolveBeans evaluates conditions for all beans and marks inactive ones.
func (c *Resolving) resolveBeans(p conf.Properties) error {
	ctx := &CondContext{p: p, c: c}
	for _, b := range c.beans {
		if err := ctx.resolveBean(b); err != nil {
			return err
		}
	}
	return nil
}

// checkDuplicateBeans ensures no duplicate type/name combinations exist.
func (c *Resolving) checkDuplicateBeans() error {
	beansByID := make(map[gs.BeanID]*gs_bean.BeanDefinition)
	for _, b := range c.beans {
		if b.Status() == gs_bean.StatusDeleted {
			continue
		}
		for _, t := range append(b.Exports(), b.Type()) {
			beanID := gs.BeanID{Name: b.Name(), Type: t}
			if d, ok := beansByID[beanID]; ok {
				return fmt.Errorf("found duplicate beans [%s] [%s]", b, d)
			}
			beansByID[beanID] = b
		}
	}
	return nil
}

// CondContext provides condition evaluation context during resolution.
type CondContext struct {
	c *Resolving
	p conf.Properties
}

// resolveBean evaluates a bean's conditions, updating its status accordingly.
// If any condition fails, the bean is marked as deleted.
func (c *CondContext) resolveBean(b *gs_bean.BeanDefinition) error {
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
func (c *CondContext) Has(key string) bool {
	return c.p.Has(key)
}

// Prop retrieves a configuration property with optional default value.
func (c *CondContext) Prop(key string, def ...string) string {
	return c.p.Get(key, def...)
}

// Find returns beans matching the selector after resolving their conditions.
func (c *CondContext) Find(s gs.BeanSelector) ([]gs.CondBean, error) {
	var found []gs.CondBean
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
