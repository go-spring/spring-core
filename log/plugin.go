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
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-spring/spring-core/util/errutil"
)

//////////////////////////////////////////////////////////////////////////////

// 类型转换器
var converters = map[reflect.Type]any{}

func init() {
	RegisterConverter(ParseLevel)
	RegisterConverter(ParseColorStyle)
}

// Converter converts a string to a target type T.
type Converter[T any] func(string) (T, error)

// RegisterConverter registers its converter for non-primitive type such as
// time.Time, time.Duration, or other user-defined value type.
func RegisterConverter[T any](fn Converter[T]) {
	t := reflect.TypeOf(fn)
	converters[t.Out(0)] = fn
}

//////////////////////////////////////////////////////////////////////////////

// Initializer 用于 plugin 的初始化
type Initializer interface {
	Init() error
}

// LifeCycle 用于 plugin 的启动和停止
type LifeCycle interface {
	Start() error
	Stop(ctx context.Context)
}

// PluginType plugin 的类型
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

// plugins stores user registered Plugin(s) .
var plugins = map[string]*Plugin{}

// Plugin is the name of node label or XML element.
type Plugin struct {
	Name  string     // 用于检查重复的 plugin
	Type  PluginType // 用于注入时候设置类型
	Class reflect.Type
	File  string
	Line  int
}

// RegisterPlugin registers a Plugin, `i` is used to obtain the type of Plugin.
func RegisterPlugin[T any](name string, typ PluginType) {
	_, file, line, _ := runtime.Caller(1)
	if p, ok := plugins[name]; ok {
		panic(fmt.Errorf("duplicate plugin %s in %s:%d and %s:%d", typ, p.File, p.Line, file, line))
	}
	p := &Plugin{
		Name:  name,
		Type:  typ,
		Class: reflect.TypeFor[T](),
		File:  file,
		Line:  line,
	}
	plugins[name] = p
}

//////////////////////////////////////////////////////////////////////////////

// NewPlugin creates a new Plugin.
func NewPlugin(t reflect.Type, node *Node) (reflect.Value, error) {
	v := reflect.New(t)
	err := inject(v.Elem(), v.Type().Elem(), node)
	if err != nil {
		return reflect.Value{}, err
	}
	i, ok := v.Interface().(Initializer)
	if ok {
		if err = i.Init(); err != nil {
			return reflect.Value{}, err
		}
	}
	return v, nil
}

// inject handles the struct field with the PluginAttribute or PluginElement tag.
func inject(v reflect.Value, t reflect.Type, node *Node) error {
	for i := 0; i < v.NumField(); i++ {
		ft := t.Field(i)
		fv := v.Field(i)
		if tag, ok := ft.Tag.Lookup("PluginAttribute"); ok {
			if err := injectAttribute(tag, fv, ft, node); err != nil {
				return err
			}
			continue
		}
		if tag, ok := ft.Tag.Lookup("PluginElement"); ok {
			if err := injectElement(tag, fv, ft, node); err != nil {
				return err
			}
			continue
		}
		if ft.Anonymous && ft.Type.Kind() == reflect.Struct {
			if err := inject(fv, fv.Type(), node); err != nil {
				return err
			}
		}
	}
	return nil
}

type PluginTag string

func (tag PluginTag) Get(key string) string {
	v, _ := tag.Lookup(key)
	return v
}

func (tag PluginTag) Lookup(key string) (value string, ok bool) {
	kvs := strings.Split(string(tag), ",")
	if key == "" {
		return kvs[0], true
	}
	for i := 1; i < len(kvs); i++ {
		ss := strings.Split(kvs[i], "=")
		if ss[0] == key {
			if len(ss) > 1 {
				return ss[1], true
			}
			return "", true
		}
	}
	return "", false
}

// injectAttribute handles the struct field with the PluginAttribute tag.
func injectAttribute(tag string, fv reflect.Value, ft reflect.StructField, node *Node) error {

	attrTag := PluginTag(tag)
	attrName := attrTag.Get("")
	val, ok := node.Attributes[attrName]
	if !ok {
		val, ok = attrTag.Lookup("default")
		if !ok {
			return fmt.Errorf("found no attribute for %s", attrName)
		}
	}

	if fn := converters[ft.Type]; fn != nil {
		fnValue := reflect.ValueOf(fn)
		out := fnValue.Call([]reflect.Value{reflect.ValueOf(val)})
		if !out[1].IsNil() {
			err := out[1].Interface().(error)
			return errutil.WrapError(err, "inject error")
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
		return errutil.WrapError(err, "inject %s error", ft.Name)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(val, 0, 0)
		if err == nil {
			fv.SetInt(i)
			return nil
		}
		return errutil.WrapError(err, "inject %s error", ft.Name)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(val, 64)
		if err == nil {
			fv.SetFloat(f)
			return nil
		}
		return errutil.WrapError(err, "inject %s error", ft.Name)
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err == nil {
			fv.SetBool(b)
			return nil
		}
		return errutil.WrapError(err, "inject %s error", ft.Name)
	case reflect.String:
		fv.SetString(val)
		return nil
	default:
		return fmt.Errorf("unsupported inject type %s for struct field %s", ft.Type.String(), ft.Name)
	}
}

// injectElement handles the struct field with the PluginElement tag.
func injectElement(tag string, fv reflect.Value, ft reflect.StructField, node *Node) error {

	elemTag := PluginTag(tag)
	elemType := elemTag.Get("")

	var children []reflect.Value
	for _, c := range node.Children {
		p, ok := plugins[c.Label]
		if !ok {
			err := fmt.Errorf("plugin %s not found", c.Label)
			return errutil.WrapError(err, "inject element")
		}
		if string(p.Type) != elemType {
			continue
		}
		v, err := NewPlugin(p.Class, c)
		if err != nil {
			return err
		}
		children = append(children, v)
	}

	if len(children) == 0 {
		elemLabel, ok := elemTag.Lookup("default")
		if !ok {
			return nil
		}
		p, ok := plugins[elemLabel]
		if !ok {
			err := fmt.Errorf("plugin %s not found", elemLabel)
			return errutil.WrapError(err, "inject element")
		}
		v, err := NewPlugin(p.Class, &Node{Label: elemLabel})
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
