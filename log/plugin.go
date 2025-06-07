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

package log

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-spring/spring-core/util/errutil"
)

var converters = map[reflect.Type]any{}

// Converter function type that converts string to a specific type T.
type Converter[T any] func(string) (T, error)

// RegisterConverter Registers a converter for a specific type T.
func RegisterConverter[T any](fn Converter[T]) {
	t := reflect.TypeFor[T]()
	converters[t] = fn
}

// Lifecycle Optional lifecycle interface for plugin instances.
type Lifecycle interface {
	Start() error
	Stop()
}

// PluginType Defines types of plugins supported by the logging system.
type PluginType string

const (
	PluginTypeAppender    PluginType = "Appender"
	PluginTypeLayout      PluginType = "Layout"
	PluginTypeAppenderRef PluginType = "AppenderRef"
	PluginTypeRoot        PluginType = "Root"
	PluginTypeAsyncRoot   PluginType = "AsyncRoot"
	PluginTypeLogger      PluginType = "Logger"
	PluginTypeAsyncLogger PluginType = "AsyncLogger"
)

var plugins = map[string]*Plugin{}

// Plugin metadata structure
type Plugin struct {
	Name  string       // Name of plugin
	Type  PluginType   // Type of plugin
	Class reflect.Type // Underlying struct type
	File  string       // Source file of registration
	Line  int          // Line number of registration
}

// RegisterPlugin Registers a plugin with a given name and type.
func RegisterPlugin[T any](name string, typ PluginType) {
	_, file, line, _ := runtime.Caller(1)
	if p, ok := plugins[name]; ok {
		panic(fmt.Errorf("duplicate plugin %s in %s:%d and %s:%d", typ, p.File, p.Line, file, line))
	}
	t := reflect.TypeFor[T]()
	if t.Kind() != reflect.Struct {
		panic("T must be struct")
	}
	plugins[name] = &Plugin{
		Name:  name,
		Type:  typ,
		Class: t,
		File:  file,
		Line:  line,
	}
}

// NewPlugin Creates and initializes a plugin instance.
func NewPlugin(t reflect.Type, node *Node, properties map[string]string) (reflect.Value, error) {
	v := reflect.New(t)
	if err := inject(v.Elem(), t, node, properties); err != nil {
		return reflect.Value{}, errutil.WrapError(err, "create plugin %s error", t.String())
	}
	return v, nil
}

// inject Recursively injects values into struct fields based on tags.
func inject(v reflect.Value, t reflect.Type, node *Node, properties map[string]string) error {
	for i := 0; i < v.NumField(); i++ {
		ft := t.Field(i)
		fv := v.Field(i)
		if tag, ok := ft.Tag.Lookup("PluginAttribute"); ok {
			if err := injectAttribute(tag, fv, ft, node, properties); err != nil {
				return err
			}
			continue
		}
		if tag, ok := ft.Tag.Lookup("PluginElement"); ok {
			if err := injectElement(tag, fv, ft, node, properties); err != nil {
				return err
			}
			continue
		}
		// Recursively process anonymous embedded structs
		if ft.Anonymous && ft.Type.Kind() == reflect.Struct {
			if err := inject(fv, fv.Type(), node, properties); err != nil {
				return err
			}
		}
	}
	return nil
}

type PluginTag string

// Get Gets the value of a key or the first unnamed value.
func (tag PluginTag) Get(key string) string {
	v, _ := tag.Lookup(key)
	return v
}

// Lookup Looks up a key-value pair in the tag.
func (tag PluginTag) Lookup(key string) (value string, ok bool) {
	kvs := strings.Split(string(tag), ",")
	if key == "" {
		return kvs[0], true
	}
	for i := 1; i < len(kvs); i++ {
		ss := strings.Split(kvs[i], "=")
		if len(ss) != 2 {
			return "", false
		} else if ss[0] == key {
			return ss[1], true
		}
	}
	return "", false
}

// injectAttribute Injects a value into a struct field from plugin attribute.
func injectAttribute(tag string, fv reflect.Value, ft reflect.StructField, node *Node, properties map[string]string) error {

	attrTag := PluginTag(tag)
	attrName := attrTag.Get("")
	if attrName == "" {
		return fmt.Errorf("found no attribute for struct field %s", ft.Name)
	}
	val, ok := node.Attributes[attrName]
	if !ok {
		val, ok = attrTag.Lookup("default")
		if !ok {
			return fmt.Errorf("found no attribute for struct field %s", ft.Name)
		}
	}

	// Use a property if available
	val = strings.TrimSpace(val)
	if strings.HasPrefix(val, "${") && strings.HasSuffix(val, "}") {
		s, exist := properties[val[2:len(val)-1]]
		if !exist {
			return fmt.Errorf("property %s not found", val)
		}
		val = s
	}

	// Use a custom converter if available
	if fn := converters[ft.Type]; fn != nil {
		fnValue := reflect.ValueOf(fn)
		out := fnValue.Call([]reflect.Value{reflect.ValueOf(val)})
		if !out[1].IsNil() {
			err := out[1].Interface().(error)
			return errutil.WrapError(err, "inject struct field %s error", ft.Name)
		}
		fv.Set(out[0])
		return nil
	}

	switch fv.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(val, 0, 0)
		if err == nil {
			fv.SetUint(u)
			return nil
		}
		return errutil.WrapError(err, "inject struct field %s error", ft.Name)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(val, 0, 0)
		if err == nil {
			fv.SetInt(i)
			return nil
		}
		return errutil.WrapError(err, "inject struct field %s error", ft.Name)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(val, 64)
		if err == nil {
			fv.SetFloat(f)
			return nil
		}
		return errutil.WrapError(err, "inject struct field %s error", ft.Name)
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err == nil {
			fv.SetBool(b)
			return nil
		}
		return errutil.WrapError(err, "inject struct field %s error", ft.Name)
	case reflect.String:
		fv.SetString(val)
		return nil
	default:
		return fmt.Errorf("unsupported inject type %s for struct field %s", ft.Type.String(), ft.Name)
	}
}

// injectElement Injects plugin elements (child nodes) into struct fields.
func injectElement(tag string, fv reflect.Value, ft reflect.StructField, node *Node, properties map[string]string) error {

	elemTag := PluginTag(tag)
	elemType := elemTag.Get("")
	if elemType == "" {
		return fmt.Errorf("found no element for struct field %s", ft.Name)
	}

	var children []reflect.Value
	for _, c := range node.Children {
		p, ok := plugins[c.Label]
		if !ok {
			return fmt.Errorf("plugin %s not found for struct field %s", c.Label, ft.Name)
		}
		if string(p.Type) != elemType {
			continue
		}
		v, err := NewPlugin(p.Class, c, properties)
		if err != nil {
			return err
		}
		children = append(children, v)
	}

	if len(children) == 0 {
		elemLabel, ok := elemTag.Lookup("default")
		if !ok {
			return fmt.Errorf("found no plugin elements for struct field %s", ft.Name)
		}
		p, ok := plugins[elemLabel]
		if !ok {
			return fmt.Errorf("plugin %s not found for struct field %s", elemLabel, ft.Name)
		}
		v, err := NewPlugin(p.Class, &Node{Label: elemLabel}, properties)
		if err != nil {
			return err
		}
		children = append(children, v)
	}

	switch fv.Kind() {
	case reflect.Slice:
		slice := reflect.MakeSlice(ft.Type, 0, len(children))
		for j := 0; j < len(children); j++ {
			slice = reflect.Append(slice, children[j])
		}
		fv.Set(slice)
		return nil
	case reflect.Interface:
		if len(children) > 1 {
			return fmt.Errorf("found %d plugin elements for struct field %s", len(children), ft.Name)
		}
		fv.Set(children[0])
		return nil
	default:
		return fmt.Errorf("unsupported inject type %s for struct field %s", ft.Type.String(), ft.Name)
	}
}
