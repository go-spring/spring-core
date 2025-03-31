package resolving

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/util"
)

type GroupFunc = func(p conf.Properties) ([]*gs.BeanDefinition, error)

// BeanMock represents a mocked bean with an object and a target selector.
type BeanMock struct {
	Object interface{}     // The mock object instance
	Target gs.BeanSelector // The target bean selector
}

// Resolving manages mocks, beans, and group functions.
type Resolving struct {
	mocks []BeanMock
	beans []*gs_bean.BeanDefinition
	funcs []GroupFunc
}

// New creates a new Resolving instance.
func New() *Resolving {
	return &Resolving{}
}

// Mock registers a mock object with a specified bean selector.
func (c *Resolving) Mock(obj interface{}, target gs.BeanSelector) {
	x := BeanMock{Object: obj, Target: target}
	c.mocks = append(c.mocks, x)
}

// Register adds a bean definition to the list of managed beans.
func (c *Resolving) Register(b *gs_bean.BeanDefinition) {
	c.beans = append(c.beans, b)
}

// GroupRegister registers a group function for bean resolution.
func (c *Resolving) GroupRegister(fn GroupFunc) {
	c.funcs = append(c.funcs, fn)
}

// RefreshInit initializes and processes group functions, configuration beans, and mocks.
func (c *Resolving) RefreshInit(p conf.Properties) error {

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

	// processes configuration beans to register beans.
	for _, b := range c.beans {
		if b.Configuration() == nil {
			continue
		}
		var foundMock BeanMock
		for _, x := range c.mocks {
			t, s := x.Target.TypeAndName()
			if t != b.Type() { // type is not same
				continue
			}
			if s != "" && s != b.Name() { // name is not equal
				continue
			}
			foundMock = x
			break
		}
		if foundMock.Target != nil {
			b.SetMock(foundMock.Object)
			continue
		}
		newBeans, err := c.scanConfiguration(b)
		if err != nil {
			return err
		}
		c.beans = append(c.beans, newBeans...)
	}

	for _, x := range c.mocks {
		var found []*gs_bean.BeanDefinition
		t, s := x.Target.TypeAndName()
		vt := reflect.TypeOf(x.Object)
		switch t.Kind() {
		case reflect.Interface:
			for _, b := range c.beans {
				if b.Type().Kind() == reflect.Interface {
					if t != b.Type() { // type is not same
						foundType := false
						for _, et := range b.Exports() {
							if et == t {
								foundType = true
								break
							}
						}
						if foundType {
							return fmt.Errorf("found unimplemented interfaces")
						}
						continue
					}
					for _, et := range b.Exports() {
						if !vt.Implements(et) {
							return fmt.Errorf("found unimplemented interfaces")
						}
					}
				} else {
					foundType := false
					for _, et := range b.Exports() {
						if et == t {
							foundType = true
							break
						}
					}
					if !foundType {
						continue
					}
					if len(b.Exports()) > 1 {
						return fmt.Errorf("found unimplemented interfaces")
					}
				}
				if s != "" && s != b.Name() { // name is not equal
					continue
				}
				found = append(found, b)
			}
		default:
			for _, b := range c.beans {
				if t != b.Type() { // type is not same
					continue
				}
				for _, et := range b.Exports() {
					if !vt.Implements(et) {
						return fmt.Errorf("found unimplemented interfaces")
					}
				}
				if s != "" && s != b.Name() { // name is not equal
					continue
				}
				found = append(found, b)
			}
		}
		if len(found) == 0 {
			continue
		}
		if len(found) > 1 {
			return fmt.Errorf("found duplicate mocked beans")
		}
		found[0].SetMock(x.Object)
	}

	return nil
}

// Refresh validates and resolves all beans in the system.
func (c *Resolving) Refresh(p conf.Properties) ([]*gs_bean.BeanDefinition, error) {

	// resolves all beans on their condition.
	ctx := &CondContext{p: p, c: c}
	for _, b := range c.beans {
		if err := ctx.resolveBean(b); err != nil {
			return nil, err
		}
	}

	type BeanID struct {
		s string
		t reflect.Type
	}

	// caches all beans by id and checks for duplicates.
	beansByID := make(map[BeanID]*gs_bean.BeanDefinition)
	for _, b := range c.beans {
		if b.Status() == gs_bean.StatusDeleted {
			continue
		}
		if b.Status() != gs_bean.StatusResolved {
			return nil, fmt.Errorf("unexpected status %d", b.Status())
		}
		beanID := BeanID{b.Name(), b.Type()}
		if d, ok := beansByID[beanID]; ok {
			return nil, fmt.Errorf("found duplicate beans [%s] [%s]", b, d)
		}
		beansByID[beanID] = b
	}
	return c.beans, nil
}

func (c *Resolving) scanConfiguration(bd *gs_bean.BeanDefinition) ([]*gs_bean.BeanDefinition, error) {
	var (
		includes []*regexp.Regexp
		excludes []*regexp.Regexp
	)
	param := bd.Configuration()
	ss := param.Includes
	if len(ss) == 0 {
		ss = []string{"New*"}
	}
	for _, s := range ss {
		var x *regexp.Regexp
		x, err := regexp.Compile(s)
		if err != nil {
			return nil, err
		}
		includes = append(includes, x)
	}
	ss = param.Excludes
	for _, s := range ss {
		var x *regexp.Regexp
		x, err := regexp.Compile(s)
		if err != nil {
			return nil, err
		}
		excludes = append(excludes, x)
	}
	var newBeans []*gs_bean.BeanDefinition
	n := bd.Type().NumMethod()
	for i := 0; i < n; i++ {
		m := bd.Type().Method(i)
		skip := false
		for _, x := range excludes {
			if x.MatchString(m.Name) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		for _, x := range includes {
			if !x.MatchString(m.Name) {
				continue
			}
			fnType := m.Func.Type()
			out0 := fnType.Out(0)
			file, line, _ := util.FileLine(m.Func.Interface())
			f, err := gs_arg.NewCallable(m.Func.Interface(), []gs.Arg{
				gs_arg.Tag(bd.Name()),
			})
			if err != nil {
				return nil, err
			}
			v := reflect.New(out0)
			if util.IsBeanType(out0) {
				v = v.Elem()
			}
			name := bd.Name() + "_" + m.Name
			b := gs_bean.NewBean(v.Type(), v, f, name)
			b.SetFileLine(file, line)
			b.SetCondition(gs_cond.OnBeanSelector(bd))
			newBeans = append(newBeans, b)
			break
		}
	}
	return newBeans, nil
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
	t, name := s.TypeAndName()
	var result []gs.CondBean
	for _, b := range c.c.beans {
		if b.Status() == gs_bean.StatusResolving || b.Status() == gs_bean.StatusDeleted {
			continue
		}
		if t != nil {
			if b.Type() != t {
				foundType := false
				for _, typ := range b.Exports() {
					if typ == t {
						foundType = true
						break
					}
				}
				if !foundType {
					continue
				}
			}
		}
		if name != "" && name != b.Name() {
			continue
		}
		if err := c.resolveBean(b); err != nil {
			return nil, err
		}
		if b.Status() == gs_bean.StatusDeleted {
			continue
		}
		result = append(result, b)
	}
	return result, nil
}
