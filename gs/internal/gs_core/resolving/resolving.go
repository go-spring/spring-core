package resolving

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/util"
)

type BeanMock struct {
	Object interface{}
	Target gs.BeanSelector
}

type Resolving struct {
	Mocks []BeanMock
	Beans []*gs_bean.BeanDefinition
	Funcs []gs.GroupFunc
}

func (c *Resolving) RefreshInit(p gs.Properties) error {
	// processes all group functions to register beans.
	for _, fn := range c.Funcs {
		beans, err := fn(p)
		if err != nil {
			return err
		}
		for _, b := range beans {
			d := b.BeanRegistration().(*gs_bean.BeanDefinition)
			c.Beans = append(c.Beans, d)
		}
	}

	// processes configuration beans to register beans.
	for _, b := range c.Beans {
		if !b.ConfigurationBean() {
			continue
		}
		newBeans, err := c.scanConfiguration(b)
		if err != nil {
			return err
		}
		c.Beans = append(c.Beans, newBeans...)
	}
	return nil
}

func (c *Resolving) Refresh(p gs.Properties) error {

	// resolves all beans on their condition.
	ctx := &CondContext{p: p, c: c}
	for _, b := range c.Beans {
		if err := ctx.resolveBean(b); err != nil {
			return err
		}
	}

	type BeanID struct {
		s string
		t reflect.Type
	}

	// caches all beans by id and checks for duplicates.
	beansByID := make(map[BeanID]*gs_bean.BeanDefinition)
	for _, b := range c.Beans {
		if b.Status() == gs_bean.StatusDeleted {
			continue
		}
		if b.Status() != gs_bean.StatusResolved {
			return fmt.Errorf("unexpected status %d", b.Status())
		}
		beanID := BeanID{b.Name(), b.Type()}
		if d, ok := beansByID[beanID]; ok {
			return fmt.Errorf("found duplicate beans [%s] [%s]", b, d)
		}
		beansByID[beanID] = b
	}
	return nil
}

func (c *Resolving) scanConfiguration(bd *gs_bean.BeanDefinition) ([]*gs_bean.BeanDefinition, error) {
	var (
		includes []*regexp.Regexp
		excludes []*regexp.Regexp
	)
	param := bd.ConfigurationParam()
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
			f, err := gs_arg.Bind(m.Func.Interface(), []gs.Arg{
				gs_arg.Tag(bd.Name()),
			})
			if err != nil {
				return nil, err
			}
			f.SetFileLine(file, line)
			v := reflect.New(out0)
			if util.IsBeanType(out0) {
				v = v.Elem()
			}
			name := bd.Name() + "_" + m.Name
			b := gs_bean.NewBean(v.Type(), v, f, name)
			b.SetFileLine(file, line)
			b.SetCondition(gs_cond.OnBean(bd))
			newBeans = append(newBeans, b)
			break
		}
	}
	return newBeans, nil
}

type CondContext struct {
	c *Resolving
	p gs.Properties
}

// resolveBean determines the validity of the bean.
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

func (c *CondContext) Has(key string) bool {
	return c.p.Has(key)
}

func (c *CondContext) Prop(key string, def ...string) string {
	return c.p.Get(key, def...)
}

// Find 查找符合条件的 bean 对象，注意该函数只能保证返回的 bean 是有效的，
// 即未被标记为删除的，而不能保证已经完成属性绑定和依赖注入。
func (c *CondContext) Find(s gs.BeanSelector) ([]gs.CondBean, error) {
	t, name := s.TypeAndName()
	var result []gs.CondBean
	for _, b := range c.c.Beans {
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
