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

package gs_core

import (
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_util"
	"github.com/go-spring/spring-core/util"
	"github.com/go-spring/spring-core/util/syslog"
)

var (
	GsContextType = reflect.TypeOf((*gs.Context)(nil)).Elem()
)

// Find 查找符合条件的 bean 对象，注意该函数只能保证返回的 bean 是有效的，
// 即未被标记为删除的，而不能保证已经完成属性绑定和依赖注入。
func (c *Container) Find(selector gs.BeanSelector) ([]gs.CondBean, error) {

	finder := func(fn func(*gs_bean.BeanDefinition) bool) ([]gs.CondBean, error) {
		var result []gs.CondBean
		for _, b := range c.beans {
			if b.Status() == gs_bean.StatusResolving || b.Status() == gs_bean.StatusDeleted || !fn(b) {
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

	var t reflect.Type
	switch st := selector.(type) {
	case string, *gs_bean.BeanDefinition:
		tag, err := c.toWireTag(selector)
		if err != nil {
			return nil, err
		}
		return finder(func(b *gs_bean.BeanDefinition) bool {
			return b.Match(tag.typeName, tag.beanName)
		})
	case reflect.Type:
		t = st
	default:
		t = reflect.TypeOf(st)
	}

	if t.Kind() == reflect.Ptr {
		if e := t.Elem(); e.Kind() == reflect.Interface {
			t = e // 指 (*error)(nil) 形式的 bean 选择器
		}
	}

	return finder(func(b *gs_bean.BeanDefinition) bool {
		if b.Type() == t {
			return true
		}
		return false
	})
}

// lazyField represents a lazy-injected field with metadata.
type lazyField struct {
	v    reflect.Value // The value to be injected.
	path string        // Path for the field in the injection hierarchy.
	tag  string        // Associated tag for the field.
}

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
func getBeforeDestroyers(destroyers *list.List, i interface{}) (*list.List, error) {
	d := i.(*destroyer)
	result := list.New()
	for e := destroyers.Front(); e != nil; e = e.Next() {
		c := e.Value.(*destroyer)
		if d.foundEarlier(c.current) {
			result.PushBack(c)
		}
	}
	return result, nil
}

// WiringStack tracks the injection path of beans and their destroyers.
type WiringStack struct {
	destroyers   *list.List
	destroyerMap map[string]*destroyer
	beans        []*gs_bean.BeanDefinition
	lazyFields   []lazyField
}

// NewWiringStack creates a new WiringStack instance.
func NewWiringStack() *WiringStack {
	return &WiringStack{
		destroyers:   list.New(),
		destroyerMap: make(map[string]*destroyer),
	}
}

// pushBack adds a bean to the injection path.
func (s *WiringStack) pushBack(b *gs_bean.BeanDefinition) {
	syslog.Debugf("push %s %s", b, b.Status())
	s.beans = append(s.beans, b)
}

// popBack removes the last bean from the injection path.
func (s *WiringStack) popBack() {
	n := len(s.beans)
	b := s.beans[n-1]
	s.beans = s.beans[:n-1]
	syslog.Debugf("pop %s %s", b, b.Status())
}

// path returns the injection path as a string.
func (s *WiringStack) path() (path string) {
	for _, b := range s.beans {
		path += fmt.Sprintf("=> %s ↩\n", b)
	}
	return path[:len(path)-1] // Remove the trailing newline.
}

// saveDestroyer tracks a bean with a destroy function, ensuring no duplicates.
func (s *WiringStack) saveDestroyer(b *gs_bean.BeanDefinition) *destroyer {
	d, ok := s.destroyerMap[b.ID()]
	if !ok {
		d = &destroyer{current: b}
		s.destroyerMap[b.ID()] = d
	}
	return d
}

// sortDestroyers sorts beans with destroy functions by dependency order.
func (s *WiringStack) sortDestroyers() ([]func(), error) {

	destroy := func(v reflect.Value, fn interface{}) func() {
		return func() {
			if fn == nil {
				v.Interface().(gs_bean.BeanDestroy).OnBeanDestroy()
			} else {
				fnValue := reflect.ValueOf(fn)
				out := fnValue.Call([]reflect.Value{v})
				if len(out) > 0 && !out[0].IsNil() {
					syslog.Errorf("%s", out[0].Interface().(error).Error())
				}
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

// wireTag represents a parsed injection tag in the format TypeName:BeanName?.
type wireTag struct {
	typeName string // Full type name.
	beanName string // Bean name for injection.
	nullable bool   // Whether the injection can be nil.
}

// parseWireTag parses a wire tag from its string representation.
func parseWireTag(str string) (tag wireTag) {

	if str == "" {
		return
	}

	if n := len(str) - 1; str[n] == '?' {
		tag.nullable = true
		str = str[:n]
	}

	i := strings.Index(str, ":")
	if i < 0 {
		tag.beanName = str
		return
	}

	tag.typeName = str[:i]
	tag.beanName = str[i+1:]
	return
}

// String converts a wireTag back to its string representation.
func (tag wireTag) String() string {
	b := bytes.NewBuffer(nil)
	if tag.typeName != "" {
		b.WriteString(tag.typeName)
		b.WriteString(":")
	}
	b.WriteString(tag.beanName)
	if tag.nullable {
		b.WriteString("?")
	}
	return b.String()
}

// toWireString converts a slice of wireTags to a comma-separated string.
func toWireString(tags []wireTag) string {
	var buf bytes.Buffer
	for i, tag := range tags {
		buf.WriteString(tag.String())
		if i < len(tags)-1 {
			buf.WriteByte(',')
		}
	}
	return buf.String()
}

// resolveTag preprocesses a tag string, resolving properties if necessary.
func (c *Container) resolveTag(tag string) (string, error) {
	if strings.HasPrefix(tag, "${") {
		s, err := c.p.Data().Resolve(tag)
		if err != nil {
			return "", err
		}
		return s, nil
	}
	return tag, nil
}

// toWireTag converts a BeanSelector to a wireTag.
func (c *Container) toWireTag(selector gs.BeanSelector) (wireTag, error) {
	switch s := selector.(type) {
	case string:
		s, err := c.resolveTag(s)
		if err != nil {
			return wireTag{}, err
		}
		return parseWireTag(s), nil
	case gs_bean.BeanDefinition:
		return parseWireTag(s.ID()), nil
	case *gs_bean.BeanDefinition:
		return parseWireTag(s.ID()), nil
	default:
		return parseWireTag(util.TypeName(s) + ":"), nil
	}
}

// autowire injects dependencies into a given value based on tags.
func (c *Container) autowire(v reflect.Value, tags []wireTag, nullable bool, stack *WiringStack) error {
	if c.ForceAutowireIsNullable {
		for i := 0; i < len(tags); i++ {
			tags[i].nullable = true
		}
	}
	switch v.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		return c.collectBeans(v, tags, nullable, stack)
	default:
		var tag wireTag
		if len(tags) > 0 {
			tag = tags[0]
		} else if nullable {
			tag.nullable = true
		}
		return c.getBean(v, tag, stack)
	}
}

type byBeanName []BeanRuntime

func (b byBeanName) Len() int           { return len(b) }
func (b byBeanName) Less(i, j int) bool { return b[i].Name() < b[j].Name() }
func (b byBeanName) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

// filterBean searches for a bean matching the given tag and type.
// Returns the index of the matched bean in the array or -1 if no match is found.
// If multiple matches are found, an error is returned.
func filterBean(beans []BeanRuntime, tag wireTag, t reflect.Type) (int, error) {
	var found []int
	for i, b := range beans {
		if b.Match(tag.typeName, tag.beanName) {
			found = append(found, i)
		}
	}

	if len(found) > 1 {
		msg := fmt.Sprintf("found %d beans, bean:%q type:%q [", len(found), tag, t)
		for _, i := range found {
			msg += "( " + beans[i].String() + " ), "
		}
		msg = msg[:len(msg)-2] + "]"
		return -1, errors.New(msg)
	}

	if len(found) > 0 {
		return found[0], nil
	}

	if tag.nullable {
		return -1, nil
	}

	return -1, fmt.Errorf("can't find bean, bean:%q type:%q", tag, t)
}

// collectBeans collects beans into the given slice or map value `v`.
// It supports dependency injection by resolving matching beans based on tags.
func (c *Container) collectBeans(v reflect.Value, tags []wireTag, nullable bool, stack *WiringStack) error {
	t := v.Type()
	if t.Kind() != reflect.Slice && t.Kind() != reflect.Map {
		return fmt.Errorf("should be slice or map in collection mode")
	}

	et := t.Elem()
	if !util.IsBeanReceiver(et) {
		return fmt.Errorf("%s is not a valid receiver type", t.String())
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
			anyBeans  []BeanRuntime
			afterAny  []BeanRuntime
			beforeAny []BeanRuntime
		)
		foundAny := false
		for _, item := range tags {

			// 是否遇到了"无序"标记
			if item.beanName == "*" {
				if foundAny {
					return fmt.Errorf("more than one * in collection %q", tags)
				}
				foundAny = true
				continue
			}

			index, err := filterBean(beans, item, et)
			if err != nil {
				return err
			}
			if index < 0 {
				continue
			}

			if foundAny {
				afterAny = append(afterAny, beans[index])
			} else {
				beforeAny = append(beforeAny, beans[index])
			}

			tmpBeans := append([]BeanRuntime{}, beans[:index]...)
			beans = append(tmpBeans, beans[index+1:]...)
		}

		if foundAny {
			anyBeans = append(anyBeans, beans...)
		}

		n := len(beforeAny) + len(anyBeans) + len(afterAny)
		arr := make([]BeanRuntime, 0, n)
		arr = append(arr, beforeAny...)
		arr = append(arr, anyBeans...)
		arr = append(arr, afterAny...)
		beans = arr
	}

	// Handle empty beans
	if len(beans) == 0 && !nullable {
		if len(tags) == 0 {
			return fmt.Errorf("no beans collected for %q", toWireString(tags))
		}
		for _, tag := range tags {
			if !tag.nullable {
				return fmt.Errorf("no beans collected for %q", toWireString(tags))
			}
		}
		return nil
	}

	// Wire the beans based on the current state of the container
	for _, b := range beans {
		switch c.state {
		case Refreshing:
			if err := c.wireBean(b.(*gs_bean.BeanDefinition), stack); err != nil {
				return err
			}
		case Refreshed:
			if b.Status() != gs_bean.StatusWired {
				return fmt.Errorf("unexpected bean status %d", b.Status())
			}
		default:
			return fmt.Errorf("state is error for wiring")
		}
	}

	// Populate the slice or map with the resolved beans
	var ret reflect.Value
	switch t.Kind() {
	case reflect.Slice:
		sort.Sort(byBeanName(beans))
		ret = reflect.MakeSlice(t, 0, 0)
		for _, b := range beans {
			ret = reflect.Append(ret, b.Value())
		}
	case reflect.Map:
		ret = reflect.MakeMap(t)
		for _, b := range beans {
			ret.SetMapIndex(reflect.ValueOf(b.Name()), b.Value())
		}
	default:
	}
	v.Set(ret)
	return nil
}

// getBean retrieves the bean corresponding to the specified tag and assigns it to `v`.
// `v` should be an uninitialized value.
func (c *Container) getBean(v reflect.Value, tag wireTag, stack *WiringStack) error {

	// Ensure the provided value `v` is valid.
	if !v.IsValid() {
		return fmt.Errorf("receiver must be a reference type, bean:%q", tag)
	}

	t := v.Type()
	// Check if the type of `v` is a valid bean receiver type.
	if !util.IsBeanReceiver(t) {
		return fmt.Errorf("%s is not a valid receiver type", t.String())
	}

	var foundBeans []BeanRuntime
	// Iterate through all beans of the given type and match against the tag.
	for _, b := range c.beansByType[t] {
		if b.Status() == gs_bean.StatusDeleted {
			continue
		}
		if !b.Match(tag.typeName, tag.beanName) {
			continue
		}
		foundBeans = append(foundBeans, b)
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
			if !b.Match(tag.typeName, tag.beanName) {
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
			return nil
		}
		return fmt.Errorf("can't find bean, bean:%q type:%q", tag, t)
	}

	// If more than one matching bean is found, return an error.
	if len(foundBeans) > 1 {
		msg := fmt.Sprintf("found %d beans, bean:%q type:%q [", len(foundBeans), tag, t)
		for _, b := range foundBeans {
			msg += "( " + b.String() + " ), "
		}
		msg = msg[:len(msg)-2] + "]"
		return errors.New(msg)
	}

	// Retrieve the single matching bean.
	b := foundBeans[0]

	// Ensure the found bean has completed dependency injection.
	switch c.state {
	case Refreshing:
		if err := c.wireBean(b.(*gs_bean.BeanDefinition), stack); err != nil {
			return err
		}
	case Refreshed:
		if b.Status() != gs_bean.StatusWired {
			return fmt.Errorf("unexpected bean status %d", b.Status())
		}
	default:
		return fmt.Errorf("state is invalid for wiring")
	}

	// Set the value of `v` to the bean's value.
	v.Set(b.Value())
	return nil
}

// wireBean performs property binding and dependency injection for the specified bean.
// It also tracks its injection path. If the bean has an initialization function, it
// is executed after the injection is completed. If the bean depends on other beans,
// it attempts to instantiate and inject those dependencies first.
func (c *Container) wireBean(b *gs_bean.BeanDefinition, stack *WiringStack) error {

	// Check if the bean is deleted.
	if b.Status() == gs_bean.StatusDeleted {
		return fmt.Errorf("bean:%q has been deleted", b.ID())
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
	if _, ok := b.Interface().(gs_bean.BeanDestroy); ok || b.Destroy() != nil {
		haveDestroy = true
		d := stack.saveDestroyer(b)
		if i := stack.destroyers.Back(); i != nil {
			d.after(i.Value.(*gs_bean.BeanDefinition))
		}
		stack.destroyers.PushBack(b)
	}

	stack.pushBack(b)

	// Detect circular dependency.
	if b.Status() == gs_bean.StatusCreating && b.Callable() != nil {
		prev := stack.beans[len(stack.beans)-2]
		if prev.Status() == gs_bean.StatusCreating {
			return errors.New("found circular autowire")
		}
	}

	// If the bean is already being created, return early.
	if b.Status() >= gs_bean.StatusCreating {
		stack.popBack()
		return nil
	}

	// Mark the bean as being created.
	b.SetStatus(gs_bean.StatusCreating)

	// Inject dependencies for the current bean.
	for _, s := range b.DependsOn() {
		beans, err := c.Find(s)
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

	// Validate that the bean exports the appropriate interfaces.
	t := v.Type()
	for _, typ := range b.Exports() {
		if !t.Implements(typ) {
			return fmt.Errorf("%s doesn't implement interface %s", b, typ)
		}
	}

	// Wire the value of the bean.
	err = c.wireBeanValue(v, t, stack)
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

	// If the bean implements the BeanInit interface, execute its OnBeanInit method.
	if f, ok := b.Interface().(gs_bean.BeanInit); ok {
		if err = f.OnBeanInit(c); err != nil {
			return err
		}
	}

	// If the bean is refreshable, add it to the refreshable list.
	if b.Refreshable() {
		i := b.Interface().(gs.Refreshable)
		var param conf.BindParam
		err = param.BindTag(b.RefreshTag(), "")
		if err != nil {
			return err
		}
		watch := c.state == Refreshing
		if err = c.p.RefreshBean(i, param, watch); err != nil {
			return err
		}
	}

	// Mark the bean as wired and pop it from the stack.
	b.SetStatus(gs_bean.StatusWired)
	stack.popBack()
	return nil
}

// ArgContext holds a Container and a WiringStack to manage dependency injection.
type ArgContext struct {
	c     *Container
	stack *WiringStack
}

// NewArgContext creates a new ArgContext with a given Container and WiringStack.
func NewArgContext(c *Container, stack *WiringStack) *ArgContext {
	return &ArgContext{c: c, stack: stack}
}

// Matches checks if a given condition matches the container.
func (a *ArgContext) Matches(c gs.Condition) (bool, error) {
	return c.Matches(a.c)
}

// Bind binds a value to a specific tag in the container.
func (a *ArgContext) Bind(v reflect.Value, tag string) error {
	return a.c.p.Data().Bind(v, conf.Tag(tag))
}

// Wire wires a value based on a specific tag in the container.
func (a *ArgContext) Wire(v reflect.Value, tag string) error {
	return a.c.wireByTag(v, tag, a.stack)
}

// getBeanValue retrieves the value of a bean. If it is a constructor bean,
// it executes the constructor and returns the result.
func (c *Container) getBeanValue(b BeanRuntime, stack *WiringStack) (reflect.Value, error) {

	// If the bean has no callable function, return its value directly.
	if b.Callable() == nil {
		return b.Value(), nil
	}

	// Call the bean's constructor and handle errors.
	out, err := b.Callable().Call(NewArgContext(c, stack))
	if err != nil {
		return reflect.Value{}, err /* fmt.Errorf("%s:%s return error: %v", b.getClass(), b.ID(), err) */
	}

	// If the return value is of bean type, handle it accordingly.
	if val := out[0]; util.IsBeanType(val.Type()) {
		// If it's a non-pointer value type, convert it into a pointer and set it.
		if !val.IsNil() && val.Kind() == reflect.Interface && util.IsValueType(val.Elem().Type()) {
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
func (c *Container) wireBeanValue(v reflect.Value, t reflect.Type, stack *WiringStack) error {

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
func (c *Container) wireStruct(v reflect.Value, t reflect.Type, opt conf.BindParam, stack *WiringStack) error {

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
				// Handle context-aware injection.
				if ft.Type == GsContextType {
					c.ContextAware = true
				}
				if err := c.wireByTag(fv, tag, stack); err != nil {
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
				watch := c.state == Refreshing
				err := c.p.RefreshField(fv.Addr(), subParam, watch)
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

// wireByTag performs dependency injection by tag.
func (c *Container) wireByTag(v reflect.Value, tag string, stack *WiringStack) error {

	tag, err := c.resolveTag(tag)
	if err != nil {
		return err
	}

	if tag == "" {
		return c.autowire(v, nil, false, stack)
	}

	var tags []wireTag
	if tag != "?" {
		for _, s := range strings.Split(tag, ",") {
			var g wireTag
			g, err = c.toWireTag(s)
			if err != nil {
				return err
			}
			tags = append(tags, g)
		}
	}
	return c.autowire(v, tags, tag == "?", stack)
}
