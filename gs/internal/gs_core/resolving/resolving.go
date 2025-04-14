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

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/util"
)

type RefreshState int

const (
	RefreshDefault = RefreshState(iota)
	Refreshing
	Refreshed
)

type BeanGroupFunc = func(p conf.Properties) ([]*gs.BeanDefinition, error)

// BeanMock represents a mocked bean with an object and a target selector.
type BeanMock struct {
	Object interface{}     // The mock object instance
	Target gs.BeanSelector // The target bean selector
}

// Resolving manages mocks, beans, and group functions.
type Resolving struct {
	state RefreshState
	mocks []BeanMock
	beans []*gs_bean.BeanDefinition
	funcs []BeanGroupFunc
}

// New creates a new Resolving instance.
func New() *Resolving {
	return &Resolving{}
}

// Beans returns all managed beans.
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

// Mock registers a mock object with a specified bean selector.
func (c *Resolving) Mock(obj interface{}, target gs.BeanSelector) {
	mock := BeanMock{Object: obj, Target: target}
	c.mocks = append(c.mocks, mock)
}

// Object registers a bean in object form.
func (c *Resolving) Object(i interface{}) *gs.RegisteredBean {
	b := gs_bean.NewBean(reflect.ValueOf(i))
	return c.Register(b).Caller(1)
}

// Provide registers a bean in constructor function form.
func (c *Resolving) Provide(ctor interface{}, args ...gs.Arg) *gs.RegisteredBean {
	b := gs_bean.NewBean(ctor, args...)
	return c.Register(b).Caller(1)
}

// Register adds a bean definition to the list of managed beans.
func (c *Resolving) Register(b *gs.BeanDefinition) *gs.RegisteredBean {
	if c.state >= Refreshing {
		panic("container is refreshing or refreshed")
	}
	bd := b.BeanRegistration().(*gs_bean.BeanDefinition)
	c.beans = append(c.beans, bd)
	return gs.NewRegisteredBean(bd)
}

// GroupRegister registers a group function for bean resolution.
func (c *Resolving) GroupRegister(fn BeanGroupFunc) {
	c.funcs = append(c.funcs, fn)
}

// Refresh validates and resolves all beans in the system.
func (c *Resolving) Refresh(p conf.Properties) error {
	if c.state != RefreshDefault {
		return errors.New("container is refreshing or refreshed")
	}
	c.state = Refreshing

	// patches all group functions to register beans.
	if err := c.patchFuncs(p); err != nil {
		return err
	}

	// scans all configuration beans to register beans.
	if err := c.scanConfigurations(); err != nil {
		return err
	}

	// patches all mocks to the beans.
	if err := c.patchMocks(); err != nil {
		return err
	}

	// resolves all beans in the system.
	if err := c.resolveBeans(p); err != nil {
		return err
	}

	// checks all beans in the system for duplicate definitions.
	if err := c.checkDuplicateBeans(); err != nil {
		return err
	}

	c.state = Refreshed
	return nil
}

func (c *Resolving) patchFuncs(p conf.Properties) error {
	// processes all group functions to register beans.
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

func (c *Resolving) scanConfigurations() error {
	// processes configuration beans to register beans.
	for _, b := range c.beans {
		if b.Configuration() == nil {
			continue
		}
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
		temp, err := c.scanConfiguration(b)
		if err != nil {
			return err
		}
		c.beans = append(c.beans, temp...)
	}
	return nil
}

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

// isBeanMatched checks if the bean is matched with the target type.
func isBeanMatched(t reflect.Type, s string, b *gs_bean.BeanDefinition) bool {
	if s != "" && s != b.Name() {
		return false
	}
	if t != nil && t != b.Type() {
		var found bool
		for _, et := range b.Exports() {
			if et == t {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (c *Resolving) patchMocks() error {
	for _, mock := range c.mocks {
		if err := c.patchMock(mock); err != nil {
			return err
		}
	}
	return nil
}

func (c *Resolving) patchMock(mock BeanMock) error {
	var foundBeans []*gs_bean.BeanDefinition
	vt := reflect.TypeOf(mock.Object)
	t, s := mock.Target.TypeAndName()
	for _, b := range c.beans {
		if !isBeanMatched(t, s, b) {
			continue
		}
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

func (c *Resolving) resolveBeans(p conf.Properties) error {
	// resolves all beans on their condition.
	ctx := &CondContext{p: p, c: c}
	for _, b := range c.beans {
		if err := ctx.resolveBean(b); err != nil {
			return err
		}
	}
	return nil
}

func (c *Resolving) checkDuplicateBeans() error {
	type BeanID struct {
		s string
		t reflect.Type
	}
	beansByID := make(map[BeanID]*gs_bean.BeanDefinition)
	for _, b := range c.beans {
		if b.Status() == gs_bean.StatusDeleted {
			continue
		}
		types := append(b.Exports(), b.Type())
		for _, t := range types {
			beanID := BeanID{b.Name(), t}
			if d, ok := beansByID[beanID]; ok {
				return fmt.Errorf("found duplicate beans [%s] [%s]", b, d)
			}
			beansByID[beanID] = b
		}
	}
	return nil
}

type CondContext struct {
	c *Resolving
	p conf.Properties
}

// resolveBean verifies if a bean meets its conditions.
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

// Has checks if the given key exists in the configuration properties.
func (c *CondContext) Has(key string) bool {
	return c.p.Has(key)
}

// Prop retrieves the value of the given key from the configuration properties.
// If the key is not found, it returns the provided default values (if any).
func (c *CondContext) Prop(key string, def ...string) string {
	return c.p.Get(key, def...)
}

// Find searches for beans that match the specified selector.
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
