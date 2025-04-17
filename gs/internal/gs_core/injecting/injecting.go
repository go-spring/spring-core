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

/************************************ destroyer ******************************/

// destroyer stores beans with destroy functions and their call order.
type destroyer struct {
	current *gs_bean.BeanDefinition   // The current bean being processed.
	earlier []*gs_bean.BeanDefinition // Beans that must be destroyed before the current bean.
}

// foundEarlier checks if a bean is already in the earlier list.
func (d *destroyer) foundEarlier(b *gs_bean.BeanDefinition) bool {
	for _, c := range d.earlier {
		if c == b {
			return true
		}
	}
	return false
}

// after adds a bean to the earlier list, ensuring it is destroyed before the current bean.
func (d *destroyer) after(b *gs_bean.BeanDefinition) {
	if d.foundEarlier(b) {
		return
	}
	d.earlier = append(d.earlier, b)
}

// getBeforeDestroyers retrieves destroyers that should be processed before a given one for sorting purposes.
func getBeforeDestroyers(destroyers *list.List, i interface{}) *list.List {
	d := i.(*destroyer)
	result := list.New()
	for e := destroyers.Front(); e != nil; e = e.Next() {
		c := e.Value.(*destroyer)
		if d.foundEarlier(c.current) {
			result.PushBack(c)
		}
	}
	return result
}

/****************************** injecting stack ******************************/

// lazyField represents a lazy-injected field with metadata.
type lazyField struct {
	v    reflect.Value // The value to be injected.
	path string        // Path for the field in the injection hierarchy.
	tag  string        // Associated tag for the field.
}

// Stack tracks the injection path of beans and their destroyers.
type Stack struct {
	beans        []*gs_bean.BeanDefinition
	destroyers   *list.List
	destroyerMap map[gs.BeanID]*destroyer
	lazyFields   []lazyField
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
	for _, b := range s.beans {
		path += fmt.Sprintf("=> %s ↩\n", b)
	}
	return path[:len(path)-1] // Remove the trailing newline.
}

// saveDestroyer tracks a bean with a destroy function, ensuring no duplicates.
func (s *Stack) saveDestroyer(b *gs_bean.BeanDefinition) {
	beanID := gs.BeanID{Name: b.Name(), Type: b.Type()}
	d, ok := s.destroyerMap[beanID]
	if !ok {
		d = &destroyer{current: b}
		s.destroyerMap[beanID] = d
	}
	if i := s.destroyers.Back(); i != nil {
		d.after(i.Value.(*gs_bean.BeanDefinition))
	}
	s.destroyers.PushBack(b)
}

// getSortedDestroyers sorts beans with destroy functions by dependency order.
func (s *Stack) getSortedDestroyers() ([]func(), error) {

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
	destroyers, err := gs_util.TripleSort(destroyers, getBeforeDestroyers)
	if err != nil {
		return nil, err
	}

	var ret []func()
	for e := destroyers.Front(); e != nil; e = e.Next() {
		d := e.Value.(*destroyer).current
		ret = append(ret, destroy(d.Value(), d.Destroy()))
	}
	return ret, nil
}

/************************************ arg ************************************/

// ArgContext holds a Container and a Stack to manage dependency injection.
type ArgContext struct {
	c     *Injecting
	stack *Stack
}

// NewArgContext creates a new ArgContext with a given Container and Stack.
func NewArgContext(c *Injecting, stack *Stack) *ArgContext {
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

/************************************ wire ***********************************/

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
func parseWireTag(p conf.Properties, str string) (tag WireTag, err error) {
	if str == "" {
		return
	}
	if strings.HasPrefix(str, "${") {
		if str, err = p.Resolve(str); err != nil {
			return
		}
	}
	if n := len(str) - 1; str[n] == '?' {
		tag.nullable = true
		str = str[:n]
	}
	tag.beanName = str
	return
}

type Injecting struct {
	state refreshState

	beansByName map[string][]BeanRuntime // 用于查找未导出接口
	beansByType map[reflect.Type][]BeanRuntime

	p *gs_dync.Properties

	destroyers []func()

	allowCircularReferences bool
	forceAutowireIsNullable bool
}

func New(p conf.Properties) *Injecting {
	return &Injecting{
		state:       RefreshDefault,
		p:           gs_dync.New(p),
		beansByName: make(map[string][]BeanRuntime),
		beansByType: make(map[reflect.Type][]BeanRuntime),
	}
}

func (c *Injecting) RefreshProperties(p conf.Properties) error {
	return c.p.Refresh(p)
}

func (c *Injecting) Refresh(beans []*gs_bean.BeanDefinition) (err error) {

	c.allowCircularReferences = cast.ToBool(c.p.Data().Get("spring.allow-circular-references"))
	c.forceAutowireIsNullable = cast.ToBool(c.p.Data().Get("spring.force-autowire-is-nullable"))

	// registers all beans
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

	// injects all beans
	c.state = Refreshing
	for _, b := range beans {
		if err = c.wireBean(b, stack); err != nil {
			return err
		}
	}
	c.state = Refreshed

	if c.allowCircularReferences {
		// processes the bean fields that are marked for lazy injection.
		for _, f := range stack.lazyFields {
			tag := strings.TrimSuffix(f.tag, ",lazy")
			if err = c.autowire(f.v, tag, stack); err != nil {
				return fmt.Errorf("%q wired error: %s", f.path, err.Error())
			}
		}
	} else if len(stack.lazyFields) > 0 {
		return errors.New("found circular references in beans")
	}

	c.destroyers, err = stack.getSortedDestroyers()
	if err != nil {
		return err
	}

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
	stack := NewStack()
	defer func() {
		if len(stack.beans) > 0 {
			syslog.Infof("injecting path %s", stack.Path())
		}
	}()
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	return c.wireBeanValue(v, t, false, stack)
}

// Close closes the container and cleans up resources.
func (c *Injecting) Close() {
	for _, f := range c.destroyers {
		f()
	}
}

// findBeans finds beans based on a given selector.
func (c *Injecting) findBeans(s gs.BeanSelector) ([]BeanRuntime, error) {
	t, name := s.TypeAndName()
	var beans []BeanRuntime
	if t != nil {
		beans = c.beansByType[t]
	}
	if name != "" {
		if t == nil {
			return c.beansByName[name], nil
		}
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

// getSingleBean retrieves the bean corresponding to the specified tag and assigns it to `v`.
// `v` should be an uninitialized value.
func (c *Injecting) getBean(t reflect.Type, tag WireTag, stack *Stack) (BeanRuntime, error) {

	// Check if the type of `v` is a valid bean receiver type.
	if !util.IsBeanInjectionTarget(t) {
		return nil, fmt.Errorf("%s is not a valid receiver type", t.String())
	}

	var foundBeans []BeanRuntime
	// Iterate through all beans of the given type and match against the tag.
	for _, b := range c.beansByType[t] {
		if b.Status() == gs_bean.StatusDeleted {
			continue
		}
		if tag.beanName == "" || tag.beanName == b.Name() {
			foundBeans = append(foundBeans, b)
		}
	}

	// When a specific bean name is provided, find it by name.
	if t.Kind() == reflect.Interface && tag.beanName != "" {
		for _, b := range c.beansByName[tag.beanName] {
			if b.Status() == gs_bean.StatusDeleted {
				continue
			}
			if !b.Type().AssignableTo(t) {
				continue
			}
			if tag.beanName != "" && tag.beanName != b.Name() {
				continue
			}

			// Deduplicate the results.
			found := false
			for _, r := range foundBeans {
				if r == b {
					found = true
					break
				}
			}
			if !found {
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

	// Ensure the found bean has completed dependency injection.
	switch c.state {
	case Refreshing:
		if err := c.wireBean(b.(*gs_bean.BeanDefinition), stack); err != nil {
			return nil, err
		}
	case Refreshed:
		if b.Status() != gs_bean.StatusWired {
			return nil, fmt.Errorf("unexpected bean status %d", b.Status())
		}
	default:
		return nil, fmt.Errorf("state is invalid for injecting")
	}
	return b, nil
}

// getMultiBeans collects beans into the given slice or map value `v`.
// It supports dependency injection by resolving matching beans based on tags.
func (c *Injecting) getBeans(t reflect.Type, tags []WireTag, nullable bool, stack *Stack) ([]BeanRuntime, error) {

	if t.Kind() != reflect.Slice && t.Kind() != reflect.Map {
		return nil, fmt.Errorf("should be slice or map in collection mode")
	}

	et := t.Elem()
	if !util.IsBeanInjectionTarget(et) {
		return nil, fmt.Errorf("%s is not a valid receiver type", t.String())
	}

	var beans []BeanRuntime
	beans = c.beansByType[et]

	// Filter out deleted beans
	{
		var arr []BeanRuntime
		for _, b := range beans {
			if b.Status() == gs_bean.StatusDeleted {
				continue
			}
			arr = append(arr, b)
		}
		beans = arr
	}

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
				found := false
				for _, j := range temp {
					if i == j {
						found = true
						break
					}
				}
				if found {
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
	if len(beans) == 0 && !nullable {
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
	for _, b := range beans {
		switch c.state {
		case Refreshing:
			if err := c.wireBean(b.(*gs_bean.BeanDefinition), stack); err != nil {
				return nil, err
			}
		case Refreshed:
			if b.Status() != gs_bean.StatusWired {
				return nil, fmt.Errorf("unexpected bean status %d", b.Status())
			}
		default:
			return nil, fmt.Errorf("state is error for injecting")
		}
	}
	return beans, nil
}

// wireBean performs property binding and dependency injection for the specified bean.
// It also tracks its injection path. If the bean has an initialization function, it
// is executed after the injection is completed. If the bean depends on other beans,
// it attempts to instantiate and inject those dependencies first.
func (c *Injecting) wireBean(b *gs_bean.BeanDefinition, stack *Stack) error {

	// Check if the bean is deleted.
	if b.Status() == gs_bean.StatusDeleted {
		return fmt.Errorf("bean:%q has been deleted", b.String())
	}

	// If the container is refreshed and the bean is already wired, do nothing.
	if c.state == Refreshed && b.Status() == gs_bean.StatusWired {
		return nil
	}

	haveDestroy := false

	// Ensure destroy functions are cleaned up in case of failure.
	defer func() {
		if haveDestroy {
			stack.destroyers.Remove(stack.destroyers.Back())
		}
	}()

	// Record the destroy function for the bean, if it exists.
	if b.Destroy() != nil {
		haveDestroy = true
		stack.saveDestroyer(b)
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
		err = c.wireBeanValue(v, v.Type(), true, stack)
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
func (c *Injecting) getBeanValue(b BeanRuntime, stack *Stack) (reflect.Value, error) {

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
	if b.Type().Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v, nil
}

// wireBeanValue binds properties and injects dependencies into the value v. v should already be initialized.
func (c *Injecting) wireBeanValue(v reflect.Value, t reflect.Type, watchRefresh bool, stack *Stack) error {

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
	return c.wireStruct(v, t, watchRefresh, param, stack)
}

// wireStruct performs dependency injection for a struct.
func (c *Injecting) wireStruct(v reflect.Value, t reflect.Type, watchRefresh bool, opt conf.BindParam, stack *Stack) error {
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
				f := lazyField{v: fv, path: fieldPath, tag: tag}
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
				err := c.wireStruct(fv, ft.Type, watchRefresh, subParam, stack)
				if err != nil {
					return err
				}
			} else {
				// Refresh field value if needed.
				err := c.p.RefreshField(fv.Addr(), subParam, watchRefresh)
				if err != nil {
					return err
				}
			}
			continue
		}

		// Recursively wire anonymous struct fields.
		if ft.Anonymous && ft.Type.Kind() == reflect.Struct {
			if err := c.wireStruct(fv, ft.Type, watchRefresh, subParam, stack); err != nil {
				return err
			}
		}
	}
	return nil
}

// autowire performs dependency injection by tag.
func (c *Injecting) autowire(v reflect.Value, str string, stack *Stack) error {
	switch v.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		{
			var tags []WireTag
			if str != "" && str != "?" {
				for _, s := range strings.Split(str, ",") {
					g, err := parseWireTag(c.p.Data(), s)
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
		tag, err := parseWireTag(c.p.Data(), str)
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
