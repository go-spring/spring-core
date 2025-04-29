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

package conf

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-spring/spring-core/util"
	"github.com/go-spring/spring-core/util/errutil"
)

var (
	ErrNotExist      = errors.New("not exist")
	ErrInvalidSyntax = errors.New("invalid syntax")
)

// ParsedTag represents a parsed configuration tag, including key,
// default value, and optional custom splitter for list parsing.
type ParsedTag struct {
	Key      string // short property key
	Def      string // default value
	HasDef   bool   // has default value
	Splitter string // splitter's name
}

func (tag ParsedTag) String() string {
	var sb strings.Builder
	sb.WriteString("${")
	sb.WriteString(tag.Key)
	if tag.HasDef {
		sb.WriteString(":=")
		sb.WriteString(tag.Def)
	}
	sb.WriteString("}")
	if tag.Splitter != "" {
		sb.WriteString(">>")
		sb.WriteString(tag.Splitter)
	}
	return sb.String()
}

// ParseTag parses a tag string in the format `${key:=default}>>splitter`
// into a ParsedTag struct. Returns an error if the format is invalid.
func ParseTag(tag string) (ret ParsedTag, err error) {
	i := strings.LastIndex(tag, ">>")
	if i == 0 {
		err = fmt.Errorf("parse tag '%s' error: %w", tag, ErrInvalidSyntax)
		return
	}
	j := strings.LastIndex(tag, "}")
	if j <= 0 {
		err = fmt.Errorf("parse tag '%s' error: %w", tag, ErrInvalidSyntax)
		return
	}
	k := strings.Index(tag, "${")
	if k < 0 {
		err = fmt.Errorf("parse tag '%s' error: %w", tag, ErrInvalidSyntax)
		return
	}
	if i > j {
		ret.Splitter = strings.TrimSpace(tag[i+2:])
	}
	ss := strings.SplitN(tag[k+2:j], ":=", 2)
	ret.Key = ss[0]
	if len(ss) > 1 {
		ret.HasDef = true
		ret.Def = ss[1]
	}
	return
}

// BindParam holds binding metadata for a single configuration value.
type BindParam struct {
	Key      string            // full key
	Path     string            // full path
	Tag      ParsedTag         // parsed tag
	Validate reflect.StructTag // full field tag
}

// BindTag parses and binds the configuration tag to the BindParam.
// It handles default values and nested key expansion.
func (param *BindParam) BindTag(tag string, validate reflect.StructTag) error {
	param.Validate = validate
	parsedTag, err := ParseTag(tag)
	if err != nil {
		return err
	}
	if parsedTag.Key == "" { // ${:=} 默认值语法
		if parsedTag.HasDef {
			param.Tag = parsedTag
			return nil
		}
		return fmt.Errorf("parse tag '%s' error: %w", tag, ErrInvalidSyntax)
	}
	if parsedTag.Key == "ROOT" {
		parsedTag.Key = ""
	}
	if param.Key == "" {
		param.Key = parsedTag.Key
	} else if parsedTag.Key != "" {
		param.Key = param.Key + "." + parsedTag.Key
	}
	param.Tag = parsedTag
	return nil
}

// Filter defines an interface for filtering configuration fields during binding.
type Filter interface {
	Do(i any, param BindParam) (bool, error)
}

// BindValue binds a value from properties `p` to the reflect.Value `v` of type `t`
// using metadata in `param`. It supports primitives, structs, maps, and slices.
func BindValue(p Properties, v reflect.Value, t reflect.Type, param BindParam, filter Filter) (RetErr error) {

	if !util.IsPropBindingTarget(t) {
		err := errors.New("target should be value type")
		return fmt.Errorf("bind path=%s type=%s error: %w", param.Path, v.Type().String(), err)
	}

	defer func() {
		if RetErr == nil {
			tag, ok := param.Validate.Lookup("expr")
			if ok && len(tag) > 0 {
				err := validateField(tag, v.Interface())
				if err != nil {
					RetErr = err
				}
			}
		}
	}()

	switch v.Kind() {
	case reflect.Map:
		return bindMap(p, v, t, param, filter)
	case reflect.Slice:
		return bindSlice(p, v, t, param, filter)
	case reflect.Array:
		err := errors.New("use slice instead of array")
		return fmt.Errorf("bind path=%s type=%s error: %w", param.Path, v.Type().String(), err)
	default: // for linter
	}

	fn := converters[t]
	if fn == nil && v.Kind() == reflect.Struct {
		if err := bindStruct(p, v, t, param, filter); err != nil {
			return err // no wrap
		}
		return nil
	}

	val, err := resolve(p, param)
	if err != nil {
		return errutil.WrapError(err, "bind path=%s type=%s error", param.Path, v.Type().String())
	}

	if fn != nil {
		fnValue := reflect.ValueOf(fn)
		out := fnValue.Call([]reflect.Value{reflect.ValueOf(val)})
		if !out[1].IsNil() {
			err = out[1].Interface().(error)
			return errutil.WrapError(err, "bind path=%s type=%s error", param.Path, v.Type().String())
		}
		v.Set(out[0])
		return nil
	}

	switch v.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var u uint64
		if u, err = strconv.ParseUint(val, 0, 0); err == nil {
			v.SetUint(u)
			return nil
		}
		return errutil.WrapError(err, "bind path=%s type=%s error", param.Path, v.Type().String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var i int64
		if i, err = strconv.ParseInt(val, 0, 0); err == nil {
			v.SetInt(i)
			return nil
		}
		return errutil.WrapError(err, "bind path=%s type=%s error", param.Path, v.Type().String())
	case reflect.Float32, reflect.Float64:
		var f float64
		if f, err = strconv.ParseFloat(val, 64); err == nil {
			v.SetFloat(f)
			return nil
		}
		return errutil.WrapError(err, "bind path=%s type=%s error", param.Path, v.Type().String())
	case reflect.Bool:
		var b bool
		if b, err = strconv.ParseBool(val); err == nil {
			v.SetBool(b)
			return nil
		}
		return errutil.WrapError(err, "bind path=%s type=%s error", param.Path, v.Type().String())
	default:
		v.SetString(val)
		return nil
	}
}

// bindSlice binds properties to a slice value.
func bindSlice(p Properties, v reflect.Value, t reflect.Type, param BindParam, filter Filter) error {

	et := t.Elem()
	p, err := getSlice(p, et, param)
	if err != nil {
		return errutil.WrapError(err, "bind path=%s type=%s error", param.Path, v.Type().String())
	}

	slice := reflect.MakeSlice(t, 0, 0)
	defer func() { v.Set(slice) }()

	if p == nil {
		return nil
	}

	for i := 0; ; i++ {
		ev := reflect.New(et).Elem()
		subParam := BindParam{
			Key:  fmt.Sprintf("%s[%d]", param.Key, i),
			Path: fmt.Sprintf("%s[%d]", param.Path, i),
		}
		err = BindValue(p, ev, et, subParam, filter)
		if errors.Is(err, ErrNotExist) {
			break
		}
		if err != nil {
			return errutil.WrapError(err, "bind path=%s type=%s error", param.Path, v.Type().String())
		}
		slice = reflect.Append(slice, ev)
	}
	return nil
}

// getSlice retrieves and splits a string into slice elements,
// creating a new Properties instance if necessary.
func getSlice(p Properties, et reflect.Type, param BindParam) (Properties, error) {

	// properties that defined as list.
	if p.Has(param.Key + "[0]") {
		return p, nil
	}

	// properties that defined as string and needs to split into []string.
	var strVal string
	{
		if p.Has(param.Key) {
			strVal = p.Get(param.Key)
		} else {
			if !param.Tag.HasDef {
				return nil, fmt.Errorf("property %q %w", param.Key, ErrNotExist)
			}
			if param.Tag.Def == "" {
				return nil, nil
			}
			if !util.IsPrimitiveValueType(et) && converters[et] == nil {
				return nil, fmt.Errorf("can't find converter for %s", et.String())
			}
			strVal = param.Tag.Def
		}
	}
	if strVal == "" {
		return nil, nil
	}

	var (
		err    error
		arrVal []string
	)

	if s := param.Tag.Splitter; s == "" {
		arrVal = strings.Split(strVal, ",")
		for i := range arrVal {
			arrVal[i] = strings.TrimSpace(arrVal[i])
		}
	} else if fn, ok := splitters[s]; ok && fn != nil {
		if arrVal, err = fn(strVal); err != nil {
			return nil, fmt.Errorf("split error: %w, value: %q", err, strVal)
		}
	} else {
		return nil, fmt.Errorf("unknown splitter '%s'", s)
	}

	r := New()
	for i, s := range arrVal {
		k := fmt.Sprintf("%s[%d]", param.Key, i)
		_ = r.Set(k, s) // always no error
	}
	return r, nil
}

// bindMap binds properties to a map value.
func bindMap(p Properties, v reflect.Value, t reflect.Type, param BindParam, filter Filter) error {

	if param.Tag.HasDef && param.Tag.Def != "" {
		err := errors.New("map can't have a non-empty default value")
		return fmt.Errorf("bind path=%s type=%s error: %w", param.Path, v.Type().String(), err)
	}

	et := t.Elem()
	ret := reflect.MakeMap(t)
	defer func() { v.Set(ret) }()

	// 当成默认值处理
	if param.Tag.Key == "" {
		if param.Tag.HasDef {
			return nil
		}
	}

	if !p.Has(param.Key) {
		if param.Tag.HasDef {
			return nil
		}
		return fmt.Errorf("property %q %w", param.Key, ErrNotExist)
	}

	keys, err := p.SubKeys(param.Key)
	if err != nil {
		return errutil.WrapError(err, "bind path=%s type=%s error", param.Path, v.Type().String())
	}

	for _, key := range keys {
		e := reflect.New(et).Elem()
		subKey := key
		if param.Key != "" {
			subKey = param.Key + "." + key
		}
		subParam := BindParam{
			Key:  subKey,
			Path: param.Path,
		}
		if err = BindValue(p, e, et, subParam, filter); err != nil {
			return err // no wrap
		}
		ret.SetMapIndex(reflect.ValueOf(key), e)
	}
	return nil
}

// bindStruct binds properties to a struct value.
func bindStruct(p Properties, v reflect.Value, t reflect.Type, param BindParam, filter Filter) error {

	if param.Tag.HasDef && param.Tag.Def != "" {
		err := errors.New("struct can't have a non-empty default value")
		return fmt.Errorf("bind path=%s type=%s error: %w", param.Path, v.Type().String(), err)
	}

	for i := range t.NumField() {
		ft := t.Field(i)
		fv := v.Field(i)

		if !fv.CanInterface() {
			continue
		}

		subParam := BindParam{
			Key:  param.Key,
			Path: param.Path + "." + ft.Name,
		}

		if tag, ok := ft.Tag.Lookup("value"); ok {
			if err := subParam.BindTag(tag, ft.Tag); err != nil {
				return errutil.WrapError(err, "bind path=%s type=%s error", param.Path, v.Type().String())
			}
			if filter != nil {
				ret, err := filter.Do(fv.Addr().Interface(), subParam)
				if err != nil {
					return err
				}
				if ret {
					continue
				}
			}
			if err := BindValue(p, fv, ft.Type, subParam, filter); err != nil {
				return err // no wrap
			}
			continue
		}

		if ft.Anonymous {
			// embed pointer type may lead to infinite recursion.
			if ft.Type.Kind() != reflect.Struct {
				continue
			}
			if err := bindStruct(p, fv, ft.Type, subParam, filter); err != nil {
				return err // no wrap
			}
		}
	}
	return nil
}

// resolve returns property references processed property value.
func resolve(p Properties, param BindParam) (string, error) {
	const defVal = "@@def@@"
	val := p.Get(param.Key, defVal)
	if val != defVal {
		return resolveString(p, val)
	}
	if p.Has(param.Key) {
		return "", fmt.Errorf("property %s isn't simple value", param.Key)
	}
	if param.Tag.HasDef {
		return resolveString(p, param.Tag.Def)
	}
	return "", fmt.Errorf("property %s %w", param.Key, ErrNotExist)
}

// resolveString returns property references processed string.
func resolveString(p Properties, s string) (string, error) {

	// If there is no property reference, return the original string.
	start := strings.Index(s, "${")
	if start < 0 {
		return s, nil
	}

	var (
		length = len(s)
		count  = 1
		end    = -1
	)

	for i := start + 2; i < length; i++ {
		if s[i] == '$' {
			if i+1 < length && s[i+1] == '{' {
				count++
			}
		} else if s[i] == '}' {
			count--
			if count == 0 {
				end = i
				break
			}
		}
	}

	if end < 0 {
		err := ErrInvalidSyntax
		return "", fmt.Errorf("resolve string %q error: %w", s, err)
	}

	var param BindParam
	_ = param.BindTag(s[start:end+1], "")

	s1, err := resolve(p, param)
	if err != nil {
		return "", errutil.WrapError(err, "resolve string %q error", s)
	}

	s2, err := resolveString(p, s[end+1:])
	if err != nil {
		return "", errutil.WrapError(err, "resolve string %q error", s)
	}

	return s[:start] + s1 + s2, nil
}
