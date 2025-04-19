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

package injecting

import (
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_dync"
	"github.com/go-spring/spring-core/gs/internal/gs_util"
	"github.com/go-spring/spring-core/util"
	"github.com/go-spring/spring-core/util/syslog"
	"github.com/spf13/cast"
)

// BeanRuntime defines an interface for runtime bean information.
type BeanRuntime interface {
	Name() string
	Type() reflect.Type
	Value() reflect.Value
	Interface() interface{}
	Callable() *gs_arg.Callable
	Status() gs_bean.BeanStatus
	String() string
}

// refreshState represents the state of a refresh operation.
type refreshState int

const (
	RefreshDefault = refreshState(iota) // Not refreshed yet
	Refreshing                          // Currently refreshing
	Refreshed                           // Successfully refreshed
)

// Injecting defines a bean injection container.
type Injecting struct {
	p *gs_dync.Properties

	beansByName map[string][]BeanRuntime // 用于查找未导出接口
	beansByType map[reflect.Type][]BeanRuntime

	destroyers []func()

	allowCircularReferences bool
	forceAutowireIsNullable bool
}

// New creates a new Injecting instance.
func New(p conf.Properties) *Injecting {
	return &Injecting{
		p: gs_dync.New(p),
	}
}

// RefreshProperties refreshes the properties of the container.
func (c *Injecting) RefreshProperties(p conf.Properties) error {
	return c.p.Refresh(p)
}

// Refresh refreshes the container with the given beans.
func (c *Injecting) Refresh(beans []*gs_bean.BeanDefinition) (err error) {
	c.allowCircularReferences = cast.ToBool(c.p.Data().Get("spring.allow-circular-references"))
	c.forceAutowireIsNullable = cast.ToBool(c.p.Data().Get("spring.force-autowire-is-nullable"))

	// registers all beans
	c.beansByName = make(map[string][]BeanRuntime)
	c.beansByType = make(map[reflect.Type][]BeanRuntime)
	for _, b := range beans {
		c.beansByName[b.Name()] = append(c.beansByName[b.Name()], b)
		c.beansByType[b.Type()] = append(c.beansByType[b.Type()], b)
		for _, t := range b.Exports() {
			c.beansByType[t] = append(c.beansByType[t], b)
		}
	}

	stack := NewStack()
	defer func() {
		if err != nil || len(stack.beans) > 0 {
			err = fmt.Errorf("%s ↩\n%s", err, stack.Path())
			syslog.Errorf("%s", err.Error())
		}
	}()

	r := &Injector{
		state:                   RefreshDefault,
		p:                       c.p,
		beansByName:             c.beansByName,
		beansByType:             c.beansByType,
		forceAutowireIsNullable: c.forceAutowireIsNullable,
	}

	// injects all beans
	r.state = Refreshing
	for _, b := range beans {
		if err = r.wireBean(b, stack); err != nil {
			return err
		}
	}
	r.state = Refreshed

	if c.allowCircularReferences {
		// processes the bean fields that are marked for lazy injection.
		for _, f := range stack.lazyFields {
			tag := strings.TrimSuffix(f.tag, ",lazy")
			if err = r.autowire(f.v, tag, stack); err != nil {
				return fmt.Errorf("%q wired error: %s", f.path, err.Error())
			}
		}
	} else if len(stack.lazyFields) > 0 {
		return errors.New("found circular autowire")
	}

	c.destroyers = stack.getSortedDestroyers()

	if !testing.Testing() {
		if c.p.ObjectsCount() == 0 {
			c.p = nil
		}
		c.beansByName = nil
		c.beansByType = nil
		return nil
	}

	c.beansByName = make(map[string][]BeanRuntime)
	c.beansByType = make(map[reflect.Type][]BeanRuntime)
	for _, b := range beans {
		c.beansByName[b.Name()] = append(c.beansByName[b.Name()], b.BeanRuntime)
		c.beansByType[b.Type()] = append(c.beansByType[b.Type()], b.BeanRuntime)
		for _, t := range b.Exports() {
			c.beansByType[t] = append(c.beansByType[t], b.BeanRuntime)
		}
	}
	return nil
}

// Wire injects dependencies into the given object.
func (c *Injecting) Wire(obj interface{}) error {
	if !testing.Testing() {
		return errors.New("not allowed to call Wire method in non-test mode")
	}
	r := &Injector{
		state:                   Refreshed,
		p:                       gs_dync.New(c.p.Data()),
		beansByName:             c.beansByName,
		beansByType:             c.beansByType,
		forceAutowireIsNullable: c.forceAutowireIsNullable,
	}
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	return r.wireBeanValue(v, t, NewStack())
}

// Close closes the container and cleans up resources.
func (c *Injecting) Close() {
	for _, f := range slices.Backward(c.destroyers) {
		f()
	}
}

type Injector struct {
	state                   refreshState
	p                       *gs_dync.Properties
	beansByName             map[string][]BeanRuntime
	beansByType             map[reflect.Type][]BeanRuntime
	forceAutowireIsNullable bool
}

// findBeans finds beans based on a given selector.
func (c *Injector) findBeans(s gs.BeanSelector) ([]BeanRuntime, error) {
	t, name := s.TypeAndName()
	var beans []BeanRuntime
	if t != nil {
		beans = c.beansByType[t]
	}
	if name != "" {
		var ret []BeanRuntime
		for _, b := range beans {
			if name == b.Name() {
				ret = append(ret, b)
			}
		}
		beans = ret
	}
	return beans, nil
}

// WireTag represents a parsed injection tag in the format TypeName:BeanName?.
type WireTag struct {
	beanName string // Bean name for injection.
	nullable bool   // Whether the injection can be nil.
}

// String converts a wireTag back to its string representation.
func (tag WireTag) String() string {
	b := bytes.NewBuffer(nil)
	b.WriteString(tag.beanName)
	if tag.nullable {
		b.WriteString("?")
	}
	return b.String()
}

// toWireString converts a slice of wireTags to a comma-separated string.
func toWireString(tags []WireTag) string {
	var buf bytes.Buffer
	for i, tag := range tags {
		buf.WriteString(tag.String())
		if i < len(tags)-1 {
			buf.WriteByte(',')
		}
	}
	return buf.String()
}

// parseWireTag parses a wire tag string and returns a wireTag struct.
func parseWireTag(str string) (tag WireTag, err error) {
	if str != "" {
		if n := len(str) - 1; str[n] == '?' {
			tag.beanName = str[:n]
			tag.nullable = true
		} else {
			tag.beanName = str
		}
	}
	return
}

// getSingleBean retrieves the bean corresponding to the specified tag and assigns it to `v`.
// `v` should be an uninitialized value.
func (c *Injector) getBean(t reflect.Type, tag WireTag, stack *Stack) (BeanRuntime, error) {

	// Check if the type of `v` is a valid bean receiver type.
	if !util.IsBeanInjectionTarget(t) {
		return nil, fmt.Errorf("%s is not a valid receiver type", t.String())
	}

	var foundBeans []BeanRuntime
	// Iterate through all beans of the given type and match against the tag.
	for _, b := range c.beansByType[t] {
		if tag.beanName == "" || tag.beanName == b.Name() {
			foundBeans = append(foundBeans, b)
		}
	}

	// When a specific bean name is provided, find it by name.
	if t.Kind() == reflect.Interface && tag.beanName != "" {
		for _, b := range c.beansByName[tag.beanName] {
			if !b.Type().AssignableTo(t) {
				continue
			}
			if !slices.Contains(foundBeans, b) {
				foundBeans = append(foundBeans, b)
				syslog.Warnf("you should call Export() on %s", b)
			}
		}
	}

	// If no matching beans are found and the tag allows nullable beans, return nil.
	if len(foundBeans) == 0 {
		if tag.nullable {
			return nil, nil
		}
		return nil, fmt.Errorf("can't find bean, bean:%q type:%q", tag, t)
	}

	// If more than one matching bean is found, return an error.
	if len(foundBeans) > 1 {
		msg := fmt.Sprintf("found %d beans, bean:%q type:%q [", len(foundBeans), tag, t)
		for _, b := range foundBeans {
			msg += "( " + b.String() + " ), "
		}
		msg = msg[:len(msg)-2] + "]"
		return nil, errors.New(msg)
	}

	// Retrieve the single matching bean.
	b := foundBeans[0]
	if c.state == Refreshing {
		if err := c.wireBean(b.(*gs_bean.BeanDefinition), stack); err != nil {
			return nil, err
		}
	}
	return b, nil
}

// getMultiBeans collects beans into the given slice or map value `v`.
// It supports dependency injection by resolving matching beans based on tags.
func (c *Injector) getBeans(t reflect.Type, tags []WireTag, nullable bool, stack *Stack) ([]BeanRuntime, error) {

	et := t.Elem()
	if !util.IsBeanInjectionTarget(et) {
		return nil, fmt.Errorf("%s is not a valid receiver type", t.String())
	}

	beans := c.beansByType[et]

	// Process bean tags to filter and order beans
	if len(tags) > 0 {
		var (
			anyBeans  []int
			afterAny  []int
			beforeAny []int
		)
		foundAny := false
		for _, item := range tags {

			// 是否遇到了"无序"标记
			if item.beanName == "*" {
				if foundAny {
					return nil, fmt.Errorf("more than one * in collection %q", tags)
				}
				foundAny = true
				continue
			}

			var founds []int
			for i, b := range beans {
				if item.beanName == b.Name() {
					founds = append(founds, i)
				}
			}
			if len(founds) > 1 {
				msg := fmt.Sprintf("found %d beans, bean:%q type:%q [", len(founds), item, t)
				for _, i := range founds {
					msg += "( " + beans[i].String() + " ), "
				}
				msg = msg[:len(msg)-2] + "]"
				return nil, errors.New(msg)
			}
			if len(founds) == 0 {
				if item.nullable {
					continue
				}
				return nil, fmt.Errorf("can't find bean, bean:%q type:%q", item, t)
			}

			if foundAny {
				afterAny = append(afterAny, founds[0])
			} else {
				beforeAny = append(beforeAny, founds[0])
			}
		}

		if foundAny {
			temp := append(beforeAny, afterAny...)
			for i := 0; i < len(beans); i++ {
				if slices.Contains(temp, i) {
					continue
				}
				anyBeans = append(anyBeans, i)
			}
		}

		n := len(beforeAny) + len(anyBeans) + len(afterAny)
		arr := make([]BeanRuntime, 0, n)
		for _, i := range beforeAny {
			arr = append(arr, beans[i])
		}
		for _, i := range anyBeans {
			arr = append(arr, beans[i])
		}
		for _, i := range afterAny {
			arr = append(arr, beans[i])
		}

		beans = arr
	}

	// Handle empty beans
	if len(beans) == 0 {
		if nullable {
			return nil, nil
		}
		if len(tags) == 0 {
			return nil, fmt.Errorf("no beans collected for %q", toWireString(tags))
		}
		for _, tag := range tags {
			if !tag.nullable {
				return nil, fmt.Errorf("no beans collected for %q", toWireString(tags))
			}
		}
		return nil, nil
	}

	// Wire the beans based on the current state of the container
	if c.state == Refreshing {
		for _, b := range beans {
			if err := c.wireBean(b.(*gs_bean.BeanDefinition), stack); err != nil {
				return nil, err
			}
		}
	}
	return beans, nil
}

// autowire performs dependency injection by tag.
func (c *Injector) autowire(v reflect.Value, str string, stack *Stack) error {
	if strings.Contains(str, "${") {
		var err error
		if str, err = c.p.Data().Resolve(str); err != nil {
			return err
		}
	}
	switch v.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		{
			var tags []WireTag
			if str != "" && str != "?" {
				for _, s := range strings.Split(str, ",") {
					g, err := parseWireTag(s)
					if err != nil {
						return err
					}
					tags = append(tags, g)
				}
			}
			nullable := str == "?"
			if c.forceAutowireIsNullable {
				for i := 0; i < len(tags); i++ {
					tags[i].nullable = true
				}
				nullable = true
			}
			beans, err := c.getBeans(v.Type(), tags, nullable, stack)
			if err != nil {
				return err
			}
			// Populate the slice or map with the resolved beans
			switch v.Kind() {
			case reflect.Slice:
				sort.Slice(beans, func(i, j int) bool {
					return beans[i].Name() < beans[j].Name()
				})
				ret := reflect.MakeSlice(v.Type(), 0, 0)
				for _, b := range beans {
					ret = reflect.Append(ret, b.Value())
				}
				v.Set(ret)
			case reflect.Map:
				ret := reflect.MakeMap(v.Type())
				for _, b := range beans {
					ret.SetMapIndex(reflect.ValueOf(b.Name()), b.Value())
				}
				v.Set(ret)
			default:
			}
			return nil
		}
	default:
		tag, err := parseWireTag(str)
		if err != nil {
			return err
		}
		if c.forceAutowireIsNullable {
			tag.nullable = true
		}
		// Ensure the provided value `v` is valid.
		if !v.IsValid() {
			return fmt.Errorf("receiver must be a reference type, bean:%q", str)
		}
		b, err := c.getBean(v.Type(), tag, stack)
		if err != nil {
			return err
		}
		if b != nil {
			v.Set(b.Value())
		}
		return nil
	}
}

// wireBean performs property binding and dependency injection for the specified bean.
// It also tracks its injection path. If the bean has an initialization function, it
// is executed after the injection is completed. If the bean depends on other beans,
// it attempts to instantiate and inject those dependencies first.
func (c *Injector) wireBean(b *gs_bean.BeanDefinition, stack *Stack) error {

	haveDestroy := false

	// Ensure destroy functions are cleaned up in case of failure.
	defer func() {
		if haveDestroy {
			stack.popDestroyer()
		}
	}()

	// Record the destroy function for the bean, if it exists.
	if b.Destroy() != nil {
		haveDestroy = true
		stack.pushDestroyer(b)
	}

	stack.pushBean(b)

	// Detect circular dependency.
	if b.Status() == gs_bean.StatusCreating && b.Callable() != nil {
		for _, bean := range stack.beans {
			if bean == b {
				return errors.New("found circular autowire")
			}
		}
	}

	// If the bean is already being created, return early.
	if b.Status() >= gs_bean.StatusCreating {
		stack.popBean()
		return nil
	}

	// Mark the bean as being created.
	b.SetStatus(gs_bean.StatusCreating)

	// Inject dependencies for the current bean.
	for _, s := range b.DependsOn() {
		beans, err := c.findBeans(s) // todo 唯一
		if err != nil {
			return err
		}
		for _, d := range beans {
			err = c.wireBean(d.(*gs_bean.BeanDefinition), stack)
			if err != nil {
				return err
			}
		}
	}

	// Get the value of the current bean.
	v, err := c.getBeanValue(b, stack)
	if err != nil {
		return err
	}

	b.SetStatus(gs_bean.StatusCreated)

	// Check if the bean has a value and wire it if it does.
	if v.IsValid() && !b.Mocked() {

		// Wire the value of the bean.
		err = c.wireBeanValue(v, v.Type(), stack)
		if err != nil {
			return err
		}

		// Execute the bean's initialization function, if it exists.
		if b.Init() != nil {
			fnValue := reflect.ValueOf(b.Init())
			out := fnValue.Call([]reflect.Value{b.Value()})
			if len(out) > 0 && !out[0].IsNil() {
				return out[0].Interface().(error)
			}
		}
	}

	// Mark the bean as wired and pop it from the stack.
	b.SetStatus(gs_bean.StatusWired)
	stack.popBean()
	return nil
}

// getBeanValue retrieves the value of a bean. If it is a constructor bean,
// it executes the constructor and returns the result.
func (c *Injector) getBeanValue(b BeanRuntime, stack *Stack) (reflect.Value, error) {

	// If the bean has no callable function, return its value directly.
	if b.Callable() == nil {
		return b.Value(), nil
	}

	// Call the bean's constructor and handle errors.
	out, err := b.Callable().Call(NewArgContext(c, stack))
	if err != nil {
		if c.forceAutowireIsNullable {
			return reflect.Value{}, nil
		}
		return reflect.Value{}, err
	}

	if o := out[len(out)-1]; util.IsErrorType(o.Type()) {
		if i := o.Interface(); i != nil {
			if c.forceAutowireIsNullable {
				return reflect.Value{}, nil
			}
			return reflect.Value{}, i.(error)
		}
	}

	// If the return value is of bean type, handle it accordingly.
	if val := out[0]; util.IsBeanType(val.Type()) {
		// If it's a non-pointer value type, convert it into a pointer and set it.
		if !val.IsNil() && val.Kind() == reflect.Interface && util.IsPropBindingTarget(val.Elem().Type()) {
			v := reflect.New(val.Elem().Type())
			v.Elem().Set(val.Elem())
			b.Value().Set(v)
		} else {
			b.Value().Set(val)
		}
	} else {
		b.Value().Elem().Set(val)
	}

	// Return an error if the value is nil.
	if b.Value().IsNil() {
		return reflect.Value{}, fmt.Errorf("%s return nil", b.String()) // b.GetClass(), b.FileLine())
	}

	v := b.Value()
	// If the result is an interface, extract the original value.
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v, nil
}

// wireBeanValue binds properties and injects dependencies into the value v. v should already be initialized.
func (c *Injector) wireBeanValue(v reflect.Value, t reflect.Type, stack *Stack) error {

	// Dereference pointer types and adjust the target type.
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	// If v is not a struct type, no injection is needed.
	if v.Kind() != reflect.Struct {
		return nil
	}

	typeName := t.Name()
	if typeName == "" {
		// Simple types don't have names, use their string representation.
		typeName = t.String()
	}

	param := conf.BindParam{Path: typeName}
	return c.wireStruct(v, t, param, stack)
}

// wireStruct performs dependency injection for a struct.
func (c *Injector) wireStruct(v reflect.Value, t reflect.Type, opt conf.BindParam, stack *Stack) error {
	// Loop through each field of the struct.
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		fv := v.Field(i)

		// If the field is unexported, try to patch it.
		if !fv.CanInterface() {
			fv = util.PatchValue(fv)
			if !fv.CanInterface() {
				continue
			}
		}

		fieldPath := opt.Path + "." + ft.Name

		// Check for autowire or inject tags.
		tag, ok := ft.Tag.Lookup("autowire")
		if !ok {
			tag, ok = ft.Tag.Lookup("inject")
		}
		if ok {
			// Handle lazy injection.
			if strings.HasSuffix(tag, ",lazy") {
				f := LazyField{v: fv, path: fieldPath, tag: tag}
				stack.lazyFields = append(stack.lazyFields, f)
			} else {
				if err := c.autowire(fv, tag, stack); err != nil {
					return fmt.Errorf("%q wired error: %w", fieldPath, err)
				}
			}
			continue
		}

		subParam := conf.BindParam{
			Key:  opt.Key,
			Path: fieldPath,
		}

		// Bind values if the field has a "value" tag.
		if tag, ok = ft.Tag.Lookup("value"); ok {
			if err := subParam.BindTag(tag, ft.Tag); err != nil {
				return err
			}
			if ft.Anonymous {
				// Recursively wire anonymous structs.
				err := c.wireStruct(fv, ft.Type, subParam, stack)
				if err != nil {
					return err
				}
			} else {
				// Refresh field value if needed.
				err := c.p.RefreshField(fv.Addr(), subParam)
				if err != nil {
					return err
				}
			}
			continue
		}

		// Recursively wire anonymous struct fields.
		if ft.Anonymous && ft.Type.Kind() == reflect.Struct {
			if err := c.wireStruct(fv, ft.Type, subParam, stack); err != nil {
				return err
			}
		}
	}
	return nil
}

// destroyer stores beans with destroy functions and their call order.
type destroyer struct {
	current *gs_bean.BeanDefinition   // The current bean being processed.
	depends []*gs_bean.BeanDefinition // Beans that must be destroyed before the current bean.
}

// after adds a bean to the later list, ensuring it is destroyed after the current bean.
func (d *destroyer) isDependOn(b *gs_bean.BeanDefinition) bool {
	return slices.Contains(d.depends, b)
}

// after adds a bean to the later list, ensuring it is destroyed after the current bean.
func (d *destroyer) dependOn(b *gs_bean.BeanDefinition) {
	if d.isDependOn(b) {
		return
	}
	d.depends = append(d.depends, b)
}

// LazyField represents a lazy-injected field with metadata.
type LazyField struct {
	v    reflect.Value // The value to be injected.
	path string        // Path for the field in the injection hierarchy.
	tag  string        // Associated tag for the field.
}

// Stack tracks the injection path of beans and their destroyers.
type Stack struct {
	beans        []*gs_bean.BeanDefinition
	lazyFields   []LazyField
	destroyers   *list.List
	destroyerMap map[gs.BeanID]*destroyer
}

// NewStack creates a new Stack instance.
func NewStack() *Stack {
	return &Stack{
		destroyers:   list.New(),
		destroyerMap: make(map[gs.BeanID]*destroyer),
	}
}

// pushBean adds a bean to the injection path.
func (s *Stack) pushBean(b *gs_bean.BeanDefinition) {
	syslog.Debugf("push %s %s", b, b.Status())
	s.beans = append(s.beans, b)
}

// popBean removes the last bean from the injection path.
func (s *Stack) popBean() {
	n := len(s.beans)
	b := s.beans[n-1]
	s.beans[n-1] = nil
	s.beans = s.beans[:n-1]
	syslog.Debugf("pop %s %s", b, b.Status())
}

// Path returns the injection path as a string.
func (s *Stack) Path() (path string) {
	if len(s.beans) == 0 {
		return ""
	}
	for _, b := range s.beans {
		path += fmt.Sprintf("=> %s ↩\n", b)
	}
	return path[:len(path)-1] // Remove the trailing newline.
}

// pushDestroyer tracks a bean with a destroy function, ensuring no duplicates.
func (s *Stack) pushDestroyer(b *gs_bean.BeanDefinition) {
	beanID := gs.BeanID{Name: b.Name(), Type: b.Type()}
	d, ok := s.destroyerMap[beanID]
	if !ok {
		d = &destroyer{current: b}
		s.destroyerMap[beanID] = d
	}
	if i := s.destroyers.Back(); i != nil {
		d.dependOn(i.Value.(*gs_bean.BeanDefinition))
	}
	s.destroyers.PushBack(b)
}

// popDestroyer removes the last bean from the destroyer stack.
func (s *Stack) popDestroyer() {
	s.destroyers.Remove(s.destroyers.Back())
}

// getBeforeDestroyers retrieves destroyers that should be processed before a given one for sorting purposes.
func getBeforeDestroyers(destroyers *list.List, i interface{}) *list.List {
	d := i.(*destroyer)
	result := list.New()
	for e := destroyers.Front(); e != nil; e = e.Next() {
		c := e.Value.(*destroyer)
		if d.isDependOn(c.current) {
			result.PushBack(c)
		}
	}
	return result
}

// getSortedDestroyers sorts beans with destroy functions by dependency order.
func (s *Stack) getSortedDestroyers() []func() {

	destroy := func(v reflect.Value, fn interface{}) func() {
		return func() {
			fnValue := reflect.ValueOf(fn)
			out := fnValue.Call([]reflect.Value{v})
			if len(out) > 0 && !out[0].IsNil() {
				syslog.Errorf("%s", out[0].Interface().(error).Error())
			}
		}
	}

	destroyers := list.New()
	for _, d := range s.destroyerMap {
		destroyers.PushBack(d)
	}
	// the injection process should first discover cyclic dependencies
	destroyers, _ = gs_util.TripleSort(destroyers, getBeforeDestroyers)

	var ret []func()
	for e := destroyers.Front(); e != nil; e = e.Next() {
		d := e.Value.(*destroyer).current
		ret = append(ret, destroy(d.Value(), d.Destroy()))
	}
	return ret
}

// ArgContext holds a Container and a Stack to manage dependency injection.
type ArgContext struct {
	c     *Injector
	stack *Stack
}

// NewArgContext creates a new ArgContext with a given Container and Stack.
func NewArgContext(c *Injector, stack *Stack) *ArgContext {
	return &ArgContext{c: c, stack: stack}
}

func (a *ArgContext) Has(key string) bool {
	return a.c.p.Data().Has(key)
}

func (a *ArgContext) Prop(key string, def ...string) string {
	return a.c.p.Data().Get(key, def...)
}

func (a *ArgContext) Find(s gs.BeanSelector) ([]gs.CondBean, error) {
	beans, err := a.c.findBeans(s)
	if err != nil {
		return nil, err
	}
	var ret []gs.CondBean
	for _, bean := range beans {
		ret = append(ret, bean)
	}
	return ret, nil
}

// Check checks if a given condition matches the container.
func (a *ArgContext) Check(c gs.Condition) (bool, error) {
	return c.Matches(a)
}

// Bind binds a value to a specific tag in the container.
func (a *ArgContext) Bind(v reflect.Value, tag string) error {
	return a.c.p.Data().Bind(v, tag)
}

// Wire wires a value based on a specific tag in the container.
func (a *ArgContext) Wire(v reflect.Value, tag string) error {
	return a.c.autowire(v, tag, a.stack)
}
