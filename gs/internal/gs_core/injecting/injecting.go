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
	"context"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/go-spring/log"
	"github.com/go-spring/spring-base/util"
	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_dync"
	"github.com/go-spring/spring-core/gs/internal/gs_util"
	"github.com/spf13/cast"
)

// BeanRuntime defines an interface that provides runtime metadata.
type BeanRuntime interface {
	Name() string               // The name of the bean
	Type() reflect.Type         // The reflect.Type of the bean
	Value() reflect.Value       // The reflect.Value of the bean
	Interface() any             // The underlying Go interface of the bean
	Callable() *gs_arg.Callable // Optional constructor or factory metadata
	Status() gs_bean.BeanStatus // Lifecycle status of the bean
	String() string             // A readable string representation
}

// refreshState represents the state of a refresh operation.
type refreshState int

const (
	RefreshDefault = refreshState(iota) // Not refreshed yet
	Refreshing                          // Currently refreshing
	Refreshed                           // Successfully refreshed
)

// Injecting is the IoC component that handles dependency injection and
// lifecycle management for beans once they have been resolved.
type Injecting struct {
	p           *gs_dync.Properties            // Dynamic properties provider
	beansByName map[string][]BeanRuntime       // Beans indexed by name
	beansByType map[reflect.Type][]BeanRuntime // Beans indexed by type
	destroyers  []func()                       // Cleanup functions in reverse order
}

// New creates a new Injecting instance.
func New(p conf.Properties) *Injecting {
	return &Injecting{
		p: gs_dync.New(p),
	}
}

// RefreshProperties updates the dynamic property source for the container.
func (c *Injecting) RefreshProperties(p conf.Properties) error {
	return c.p.Refresh(p)
}

// Refresh wires all provided beans and prepares them for use.
//
// It performs the following steps:
//  1. Builds indexes for bean lookup by name and type.
//  2. Wires root beans (entry points of the dependency graph).
//  3. Handles lazy wiring for circular dependencies if allowed.
//  4. Captures all registered destroyer callbacks for proper shutdown order.
//  5. Optionally cleans up metadata if running outside testing.
func (c *Injecting) Refresh(roots, beans []*gs_bean.BeanDefinition) (err error) {
	allowCircularReferences := cast.ToBool(c.p.Data().Get("spring.allow-circular-references"))
	forceAutowireIsNullable := cast.ToBool(c.p.Data().Get("spring.force-autowire-is-nullable"))

	// Index beans by name and type for lookup
	c.beansByName = make(map[string][]BeanRuntime)
	c.beansByType = make(map[reflect.Type][]BeanRuntime)
	for _, b := range beans {
		c.beansByName[b.Name()] = append(c.beansByName[b.Name()], b)
		c.beansByType[b.Type()] = append(c.beansByType[b.Type()], b)
		for _, t := range b.Exports() { // Register additional exported types
			c.beansByType[t] = append(c.beansByType[t], b)
		}
	}

	stack := NewStack()
	defer func() {
		// If an error occurred, or there are unresolved beans in the stack,
		// enrich the error message with the dependency path for easier debugging.
		if err != nil || len(stack.beans) > 0 {
			err = util.FormatError(nil, "%s ↩\n%s", err, stack.Path())
			log.Errorf(context.Background(), log.TagAppDef, "%s", err)
		}
	}()

	r := &Injector{
		state:                   RefreshDefault,
		p:                       c.p,
		beansByName:             c.beansByName,
		beansByType:             c.beansByType,
		forceAutowireIsNullable: forceAutowireIsNullable,
	}

	// Step 1: Wire all root beans.
	r.state = Refreshing
	for _, b := range roots {
		if err = r.wireBean(b, stack); err != nil {
			return err
		}
	}
	r.state = Refreshed

	// Step 2: Handle lazy fields caused by circular dependencies.
	if allowCircularReferences {
		for _, f := range stack.lazyFields {
			tag := strings.TrimSuffix(f.tag, ",lazy")
			if err = r.autowire(f.v, tag, stack); err != nil {
				return err
			}
		}
	} else if len(stack.lazyFields) > 0 {
		return util.FormatError(nil, "found circular autowire")
	}

	// Step 3: Collect destroyer callbacks in dependency-safe order.
	c.destroyers = stack.getSortedDestroyers()

	// Optional cleanup in non-testing environments.
	forceClean := cast.ToBool(c.p.Data().Get("spring.force-clean"))
	if !testing.Testing() || forceClean {
		if c.p.ObjectsCount() == 0 {
			c.p = nil
		}
		c.beansByName = nil
		c.beansByType = nil
		return nil
	}

	// In testing mode, retain bean indexes to allow further Wire() calls.
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

// Wire injects dependencies into an externally provided object.
func (c *Injecting) Wire(obj any) error {
	r := &Injector{
		state:                   Refreshed,
		p:                       gs_dync.New(c.p.Data()),
		beansByName:             c.beansByName,
		beansByType:             c.beansByType,
		forceAutowireIsNullable: true,
	}
	stack := NewStack()
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	if err := r.wireBeanValue(v, t, stack); err != nil {
		return err
	}
	for _, f := range stack.lazyFields {
		tag := strings.TrimSuffix(f.tag, ",lazy")
		if err := r.autowire(f.v, tag, stack); err != nil {
			return err
		}
	}
	return nil
}

// Close shuts down the container by invoking all registered destroyer
// callbacks in reverse registration order, ensuring dependent resources
// are released safely.
func (c *Injecting) Close() {
	for _, f := range slices.Backward(c.destroyers) {
		f()
	}
}

// Injector is the component that executes core autowiring and
// bean lifecycle management (constructor, field, and method injection).
type Injector struct {
	state                   refreshState                   // Current wiring state
	p                       *gs_dync.Properties            // Property resolver
	beansByName             map[string][]BeanRuntime       // Beans indexed by name
	beansByType             map[reflect.Type][]BeanRuntime // Beans indexed by type
	forceAutowireIsNullable bool                           // Treat missing references as nullable
}

// findBeans retrieves all beans that match a given selector.
func (c *Injector) findBeans(s gs.BeanSelector) []BeanRuntime {
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
	return beans
}

// WireTag represents the parsed structure of an injection tag
// in the form "BeanName?" where "?" indicates that the dependency is optional.
type WireTag struct {
	beanName string // The target bean's name
	nullable bool   // Whether the injection can be nil
}

// String converts a WireTag back to its string representation.
func (tag WireTag) String() string {
	var sb strings.Builder
	sb.WriteString(tag.beanName)
	if tag.nullable {
		sb.WriteString("?")
	}
	return sb.String()
}

// parseWireTag parses a raw wire tag string into a structured WireTag.
func parseWireTag(str string) (tag WireTag) {
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

// toWireString converts a slice of WireTags into a comma-separated string.
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

// getBean locates a single bean that matches the given type and WireTag.
// If the container is still in the Refreshing state, the matched bean will
// be wired before it is returned.
func (c *Injector) getBean(t reflect.Type, tag WireTag, stack *Stack) (BeanRuntime, error) {
	// Ensure the target type is valid for injection.
	if !util.IsBeanInjectionTarget(t) {
		return nil, util.FormatError(nil, "%s is not a valid receiver type", t.String())
	}

	var foundBeans []BeanRuntime
	for _, b := range c.beansByType[t] {
		if tag.beanName == "" || tag.beanName == b.Name() {
			foundBeans = append(foundBeans, b)
		}
	}

	// Special handling for interface types with explicit bean names.
	if t.Kind() == reflect.Interface && tag.beanName != "" {
		for _, b := range c.beansByName[tag.beanName] {
			if !b.Type().AssignableTo(t) {
				continue
			}
			if !slices.Contains(foundBeans, b) {
				foundBeans = append(foundBeans, b)
				log.Warnf(context.Background(), log.TagAppDef, "call Export() on %s", b)
			}
		}
	}

	if len(foundBeans) == 0 {
		if tag.nullable {
			return nil, nil
		}
		return nil, util.FormatError(nil, "can't find bean, bean:%q type:%q", tag, t)
	}

	if len(foundBeans) > 1 {
		msg := fmt.Sprintf("found %d beans, bean:%q type:%q [", len(foundBeans), tag, t)
		for _, b := range foundBeans {
			msg += "( " + b.String() + " ), "
		}
		msg = msg[:len(msg)-2] + "]"
		return nil, util.FormatError(nil, "%s", msg)
	}

	b := foundBeans[0]
	if c.state == Refreshing {
		if err := c.wireBean(b.(*gs_bean.BeanDefinition), stack); err != nil {
			return nil, err
		}
	}
	return b, nil
}

// getBeans retrieves a slice or map of beans that match the required element type and optional WireTags.
// It supports filtering and ordering via tags, including the "*" wildcard to include unordered beans.
func (c *Injector) getBeans(t reflect.Type, tags []WireTag, nullable bool, stack *Stack) ([]BeanRuntime, error) {

	et := t.Elem()
	if !util.IsBeanInjectionTarget(et) {
		return nil, util.FormatError(nil, "%s is not a valid receiver type", t.String())
	}

	beans := c.beansByType[et]

	// Process bean tags to filter and order beans
	if len(tags) > 0 {
		var (
			anyBeans  []int // indices of beans to be placed in the '*' section
			afterAny  []int // beans to appear after the '*'
			beforeAny []int // beans to appear before the '*'
		)
		foundAny := false
		for _, item := range tags {

			// If we see the "*" wildcard, record its presence
			if item.beanName == "*" {
				if foundAny {
					return nil, util.FormatError(nil, "more than one * in collection %q", tags)
				}
				foundAny = true
				continue
			}

			// Find beans with the specified name
			var founds []int
			for i, b := range beans {
				if item.beanName == b.Name() {
					founds = append(founds, i)
				}
			}

			// Error if there are multiple beans with the same name
			if len(founds) > 1 {
				msg := fmt.Sprintf("found %d beans, bean:%q type:%q [", len(founds), item, t)
				for _, i := range founds {
					msg += "( " + beans[i].String() + " ), "
				}
				msg = msg[:len(msg)-2] + "]"
				return nil, util.FormatError(nil, "%s", msg)
			}

			// Error if no matching bean is found (unless the tag is nullable)
			if len(founds) == 0 {
				if item.nullable {
					continue
				}
				return nil, util.FormatError(nil, "can't find bean, bean:%q type:%q", item, t)
			}

			// Classify beans as before or after the '*'
			if foundAny {
				afterAny = append(afterAny, founds[0])
			} else {
				beforeAny = append(beforeAny, founds[0])
			}
		}

		// For the '*' wildcard, include all other beans that were not explicitly listed
		if foundAny {
			temp := append(beforeAny, afterAny...)
			for i := range len(beans) {
				if slices.Contains(temp, i) {
					continue
				}
				anyBeans = append(anyBeans, i)
			}
		}

		// Assemble beans in the correct order: beforeAny -> anyBeans -> afterAny
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

	// Handle the case where no beans were found
	if len(beans) == 0 {
		if nullable {
			return nil, nil
		}
		return nil, util.FormatError(nil, "no beans collected for %q", toWireString(tags))
	}

	// If the container is in the refreshing state, wire the beans before returning them
	if c.state == Refreshing {
		for _, b := range beans {
			if err := c.wireBean(b.(*gs_bean.BeanDefinition), stack); err != nil {
				return nil, err
			}
		}
	}
	return beans, nil
}

// autowire injects dependencies into a single field or a collection (slice/map) based on its kind and tag.
func (c *Injector) autowire(v reflect.Value, str string, stack *Stack) error {
	// Resolve placeholder expressions (e.g., ${...}) from configuration
	str, err := c.p.Data().Resolve(str)
	if err != nil {
		return err
	}

	switch v.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		{
			// Handle collection types
			var nullable bool
			var tags []WireTag

			// Parse the tag string to determine nullability and tag list
			if str != "" {
				nullable = true
				if str != "?" {
					for s := range strings.SplitSeq(str, ",") {
						g := parseWireTag(s)
						tags = append(tags, g)
						if !g.nullable {
							nullable = false
						}
					}
				}
			}

			// If forced nullable mode is enabled, override all tags
			if c.forceAutowireIsNullable {
				for i := range len(tags) {
					tags[i].nullable = true
				}
				nullable = true
			}

			// Retrieve the beans matching the tag and type
			beans, err := c.getBeans(v.Type(), tags, nullable, stack)
			if err != nil {
				return err
			}

			// Populate the collection field with the resolved beans
			switch v.Kind() {
			case reflect.Slice:
				// Sort beans by name for deterministic order
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
			default: // for linter
			}
			return nil
		}
	default:
		// Handle single bean injection
		g := parseWireTag(str)
		if c.forceAutowireIsNullable {
			g.nullable = true
		}
		b, err := c.getBean(v.Type(), g, stack)
		if err != nil {
			return err
		}
		if b != nil {
			v.Set(b.Value())
		}
		return nil
	}
}

// wireBean ensures that the specified BeanDefinition is fully constructed,
// injected, initialized, and registered for destruction.
func (c *Injector) wireBean(b *gs_bean.BeanDefinition, stack *Stack) error {

	haveDestroy := false

	// Ensure that the destroyer is popped from the stack
	defer func() {
		if haveDestroy {
			stack.popDestroyer()
		}
	}()

	// If the bean has a destroy callback, record it for later execution
	if b.Destroy() != nil {
		haveDestroy = true
		stack.pushDestroyer(b)
	}

	stack.pushBean(b)

	// Detect circular dependencies
	if b.Status() == gs_bean.StatusCreating && b.Callable() != nil {
		if slices.Contains(stack.beans, b) {
			return util.FormatError(nil, "found circular autowire")
		}
	}

	// If the bean is already being created, return early.
	if b.Status() >= gs_bean.StatusCreating {
		stack.popBean()
		return nil
	}

	// Mark the bean as currently being created
	b.SetStatus(gs_bean.StatusCreating)

	// Wire all dependent beans before creating the current bean
	for _, s := range b.DependsOn() {
		beans := c.findBeans(s)
		for _, d := range beans {
			err := c.wireBean(d.(*gs_bean.BeanDefinition), stack)
			if err != nil {
				return err
			}
		}
	}

	// Retrieve the actual value for the bean (e.g., via its factory method)
	v, err := c.getBeanValue(b, stack)
	if err != nil {
		return err
	}

	b.SetStatus(gs_bean.StatusCreated)

	// If the bean is valid and not mocked, inject its internal dependencies
	if v.IsValid() && !b.Mocked() {

		// Perform field-level wiring on the bean value
		if err = c.wireBeanValue(v, v.Type(), stack); err != nil {
			return err
		}

		// Invoke the bean's initialization method if defined
		if b.Init() != nil {
			fnValue := reflect.ValueOf(b.Init())
			out := fnValue.Call([]reflect.Value{b.Value()})
			if len(out) > 0 && !out[0].IsNil() {
				return out[0].Interface().(error)
			}
		}
	}

	// Mark the bean as fully wired and remove it from the stack
	b.SetStatus(gs_bean.StatusWired)
	stack.popBean()
	return nil
}

// getBeanValue invokes the constructor (if present) of a bean and handles return values and errors.
func (c *Injector) getBeanValue(b BeanRuntime, stack *Stack) (reflect.Value, error) {

	// If there is no constructor, return the pre-existing value
	if b.Callable() == nil {
		return b.Value(), nil
	}

	// Invoke the constructor
	out, err := b.Callable().Call(NewArgContext(c, stack))
	if err != nil {
		if c.forceAutowireIsNullable {
			log.Warnf(context.Background(), log.TagAppDef, "autowire error: %v", err)
			return reflect.Value{}, nil
		}
		return reflect.Value{}, err
	}

	// Check if the last return value is an error
	if o := out[len(out)-1]; util.IsErrorType(o.Type()) {
		if err, ok := o.Interface().(error); ok && err != nil {
			if c.forceAutowireIsNullable {
				log.Warnf(context.Background(), log.TagAppDef, "autowire error: %v", err)
				return reflect.Value{}, nil
			}
			return reflect.Value{}, err
		}
	}

	// Assign the returned value to the bean
	if val := out[0]; util.IsBeanType(val.Type()) {
		// Convert interface values to pointers if necessary
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

	// Ensure the value is not nil
	if b.Value().IsNil() {
		return reflect.Value{}, util.FormatError(nil, "%s return nil", b.String())
	}

	// If the value is an interface, unwrap it
	v := b.Value()
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v, nil
}

// wireBeanValue injects dependencies into a bean's struct fields.
func (c *Injector) wireBeanValue(v reflect.Value, t reflect.Type, stack *Stack) error {

	// Dereference pointers to obtain the underlying struct
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	// If it's not a struct, nothing to wire
	if v.Kind() != reflect.Struct {
		return nil
	}

	// Use the type name for binding paths
	typeName := t.Name()
	if typeName == "" {
		typeName = t.String()
	}

	param := conf.BindParam{Path: typeName}
	return c.wireStruct(v, t, param, stack)
}

// wireStruct inspects struct fields and performs autowiring or configuration binding as needed.
func (c *Injector) wireStruct(v reflect.Value, t reflect.Type, opt conf.BindParam, stack *Stack) error {
	for i := range t.NumField() {
		ft := t.Field(i)
		fv := v.Field(i)

		// Patch unexported fields so they can be set via reflection
		if !fv.CanInterface() {
			fv = util.PatchValue(fv)
		}

		fieldPath := opt.Path + "." + ft.Name

		// Look for "autowire" or "inject" tags
		tag, ok := ft.Tag.Lookup("autowire")
		if !ok {
			tag, ok = ft.Tag.Lookup("inject")
		}
		if ok {
			// Handle lazy-injected fields
			if strings.HasSuffix(tag, ",lazy") {
				f := LazyField{v: fv, path: fieldPath, tag: tag}
				stack.lazyFields = append(stack.lazyFields, f)
			} else {
				if err := c.autowire(fv, tag, stack); err != nil {
					return util.FormatError(err, "%q wired error", fieldPath)
				}
			}
			continue
		}

		subParam := conf.BindParam{
			Key:  opt.Key,
			Path: fieldPath,
		}

		// If the field has a "value" tag, bind configuration to it
		if tag, ok = ft.Tag.Lookup("value"); ok {
			if err := subParam.BindTag(tag, ft.Tag); err != nil {
				return err
			}
			if ft.Anonymous {
				// Recursively process embedded structs
				if err := c.wireStruct(fv, ft.Type, subParam, stack); err != nil {
					return err
				}
			} else {
				// Refresh the field value from configuration
				if err := c.p.RefreshField(fv.Addr(), subParam); err != nil {
					return err
				}
			}
			continue
		}

		// Recursively process anonymous struct fields
		if ft.Anonymous && ft.Type.Kind() == reflect.Struct {
			if err := c.wireStruct(fv, ft.Type, subParam, stack); err != nil {
				return err
			}
		}
	}
	return nil
}

// destroyer represents a bean's cleanup (destroy) function
// and the dependencies that must be destroyed before it.
type destroyer struct {
	current *gs_bean.BeanDefinition   // Bean that provides this destroyer
	depends []*gs_bean.BeanDefinition // Beans that must be destroyed before current
}

// isDependOn reports whether this destroyer depends on the given bean.
func (d *destroyer) isDependOn(b *gs_bean.BeanDefinition) bool {
	return slices.Contains(d.depends, b)
}

// dependOn adds the given bean to this destroyer's dependency list
// if it is not already present.
func (d *destroyer) dependOn(b *gs_bean.BeanDefinition) {
	if d.isDependOn(b) {
		return
	}
	d.depends = append(d.depends, b)
}

// LazyField represents a field in a struct that should be injected lazily.
type LazyField struct {
	v    reflect.Value // The field value that will be injected later
	path string        // Hierarchical path of the field
	tag  string        // Original tag (e.g. "autowire") for this field
}

// Stack represents the runtime context during bean wiring.
// It keeps track of the current wiring call stack, lazily injected fields,
// and the ordering of destroyers for proper shutdown.
type Stack struct {
	beans        []*gs_bean.BeanDefinition // The stack of beans currently being wired
	lazyFields   []LazyField               // Fields deferred due to lazy injection
	destroyers   *list.List                // Ordered list of destroyers
	destroyerMap map[gs.BeanID]*destroyer  // Fast lookup map for destroyers by bean ID
}

// NewStack creates and initializes a new Stack for a fresh Refresh or Wire operation.
func NewStack() *Stack {
	return &Stack{
		destroyers:   list.New(),
		destroyerMap: make(map[gs.BeanID]*destroyer),
	}
}

// pushBean pushes a bean onto the wiring stack.
// Used to keep track of current wiring path for cycle detection.
func (s *Stack) pushBean(b *gs_bean.BeanDefinition) {
	log.Debugf(context.Background(), log.TagAppDef, "push %s %s", b, b.Status())
	s.beans = append(s.beans, b)
}

// popBean pops the most recently added bean from the wiring stack.
func (s *Stack) popBean() {
	n := len(s.beans)
	b := s.beans[n-1]
	s.beans[n-1] = nil // avoid memory leak
	s.beans = s.beans[:n-1]
	log.Debugf(context.Background(), log.TagAppDef, "pop %s %s", b, b.Status())
}

// Path returns a formatted string representation of the current wiring stack,
// which is useful for debugging and error messages.
func (s *Stack) Path() (path string) {
	if len(s.beans) == 0 {
		return ""
	}
	for _, b := range s.beans {
		path += fmt.Sprintf("=> %s ↩\n", b)
	}
	return path[:len(path)-1] // Trim the trailing newline
}

// pushDestroyer registers a destroyer for the given bean.
// It also records dependencies so that beans are destroyed in the correct order.
func (s *Stack) pushDestroyer(b *gs_bean.BeanDefinition) {
	beanID := gs.BeanID{Name: b.Name(), Type: b.Type()}

	// Get or create the destroyer entry for this bean
	d, ok := s.destroyerMap[beanID]
	if !ok {
		d = &destroyer{current: b}
		s.destroyerMap[beanID] = d
	}

	// If there is a previously registered destroyer, current depends on it
	if i := s.destroyers.Back(); i != nil {
		d.dependOn(i.Value.(*gs_bean.BeanDefinition))
	}

	// Add the current bean to the end of the destroyer list
	s.destroyers.PushBack(b)
}

// popDestroyer removes the last registered destroyer from the ordering list.
func (s *Stack) popDestroyer() {
	s.destroyers.Remove(s.destroyers.Back())
}

// getBeforeDestroyers returns a list of destroyers that the given destroyer depends on.
// This helper is used during topological sorting of destroyers.
func getBeforeDestroyers(destroyers *list.List, i any) *list.List {
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

// getSortedDestroyers computes and returns a slice of destroyer functions
// in the correct order, respecting declared dependencies between beans.
func (s *Stack) getSortedDestroyers() []func() {

	// Helper to wrap a bean's destroy method as a no-argument function
	destroy := func(v reflect.Value, fn any) func() {
		return func() {
			fnValue := reflect.ValueOf(fn)
			out := fnValue.Call([]reflect.Value{v})
			if len(out) > 0 && !out[0].IsNil() {
				log.Errorf(context.Background(), log.TagAppDef, "%v", out[0].Interface())
			}
		}
	}

	// Copy all destroyers into a new list for sorting
	destroyers := list.New()
	for _, d := range s.destroyerMap {
		destroyers.PushBack(d)
	}

	// Perform a topological sort to respect dependencies
	// (e.g. a bean must be destroyed after the beans it depends on)
	destroyers, _ = gs_util.TripleSort(destroyers, getBeforeDestroyers)

	// Convert the sorted destroyers into a slice of executable cleanup functions
	var ret []func()
	for e := destroyers.Front(); e != nil; e = e.Next() {
		d := e.Value.(*destroyer).current
		ret = append(ret, destroy(d.Value(), d.Destroy()))
	}
	return ret
}

// ArgContext provides runtime context when calling bean factory functions.
// It exposes access to configuration properties, bean lookups, condition checks,
// and allows wiring of parameters during construction.
type ArgContext struct {
	c     *Injector
	stack *Stack
}

// NewArgContext constructs a new ArgContext for a wiring operation.
func NewArgContext(c *Injector, stack *Stack) *ArgContext {
	return &ArgContext{c: c, stack: stack}
}

// Has checks whether a configuration key is present.
func (a *ArgContext) Has(key string) bool {
	return a.c.p.Data().Has(key)
}

// Prop retrieves a property value, with optional default.
func (a *ArgContext) Prop(key string, def ...string) string {
	return a.c.p.Data().Get(key, def...)
}

// Find retrieves beans matching the given selector.
func (a *ArgContext) Find(s gs.BeanSelector) ([]gs.ConditionBean, error) {
	beans := a.c.findBeans(s)
	var ret []gs.ConditionBean
	for _, bean := range beans {
		ret = append(ret, bean)
	}
	return ret, nil
}

// Check evaluates a condition against the current ArgContext.
func (a *ArgContext) Check(c gs.Condition) (bool, error) {
	return c.Matches(a)
}

// Bind binds configuration data into the provided reflect.Value
// based on the given struct tag.
func (a *ArgContext) Bind(v reflect.Value, tag string) error {
	return a.c.p.Data().Bind(v, tag)
}

// Wire performs dependency injection on the given reflect.Value
// using the specified tag, leveraging the current wiring stack.
func (a *ArgContext) Wire(v reflect.Value, tag string) error {
	return a.c.autowire(v, tag, a.stack)
}
