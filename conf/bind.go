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

	"github.com/go-spring/stdlib/errutil"
	"github.com/go-spring/stdlib/flatten"
	"github.com/go-spring/stdlib/typeutil"
)

var (
	ErrNotExist      = errutil.Explain(nil, "not exist")
	ErrInvalidSyntax = errutil.Explain(nil, "invalid syntax")
)

// ParsedTag represents a parsed configuration tag that encodes
// metadata for binding configuration values from property sources.
//
// A tag string generally follows the pattern:
//
//	${key:=default}
//
// - "key":        the property key used to look up a value.
// - "default":    optional fallback value if the key does not exist.
//
// Examples:
//
//	"${db.host:=localhost}"       -> key=db.host, default=localhost
//	"${ports:=8080,9090}"         -> key=ports, default=8080,9090
//	"${:=foo}"                    -> empty key, only default value "foo"
//
// The parsing logic is strict; malformed tags will result in ErrInvalidSyntax.
type ParsedTag struct {
	Key    string // short property key
	Def    string // default value string
	HasDef bool   // indicates whether a default value exists
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
	return sb.String()
}

// ParseTag parses a tag string into a ParsedTag struct.
//
// Supported syntax: `${key:=default}`
//
// - The `${...}` block is mandatory.
// - ":=" introduces an optional default value.
//
// Example parses:
//
//	"${foo}"               -> Key="foo"
//	"${foo:=bar}"          -> Key="foo", HasDef=true, Def="bar"
//	"${foo:=bar}"          -> Key="foo", HasDef=true, Def="bar"
//	"${:=fallback}"        -> Key="", HasDef=true, Def="fallback"
//
// Errors:
//   - Returns ErrInvalidSyntax if the string does not follow the pattern.
func ParseTag(tag string) (ret ParsedTag, err error) {
	j := strings.LastIndex(tag, "}")
	if j <= 0 {
		err = errutil.Explain(ErrInvalidSyntax, "parse tag '%s' error", tag)
		return
	}
	k := strings.Index(tag, "${")
	if k < 0 {
		err = errutil.Explain(ErrInvalidSyntax, "parse tag '%s' error", tag)
		return
	}
	ss := strings.SplitN(tag[k+2:j], ":=", 2)
	ret.Key = strings.TrimSpace(ss[0])
	if len(ss) > 1 {
		ret.HasDef = true
		ret.Def = strings.TrimSpace(ss[1])
	}
	return
}

// BindParam holds metadata needed to bind a single configuration value
// to a Go struct field, slice element, or map entry.
type BindParam struct {
	Key      string            // full property key
	Path     string            // full property path
	Tag      ParsedTag         // parsed tag
	Validate reflect.StructTag // original struct field tag for validation
}

// BindTag parses the tag string, stores the ParsedTag in BindParam,
// and resolves nested key expansion.
//
// Special cases:
// - "${:=default}" -> Key is empty, only default is set.
// - "${ROOT}"      -> explicitly resets Key to an empty string.
//
// If a BindParam already has a Key, new keys are appended hierarchically,
// e.g. parent Key="db", tag="${host}" -> final Key="db.host".
//
// Errors:
// - ErrInvalidSyntax if the tag string is malformed or empty without default.
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
		return errutil.Explain(ErrInvalidSyntax, "parse tag '%s' error", tag)
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

// BindValue attempts to bind a property value from the property source `p`
// into the given reflect.Value `v`, based on metadata in `param`.
//
// Supported binding targets:
// - Primitive types (string, int, float, bool, etc.).
// - Structs (recursively bound field by field).
// - Maps (bound by iterating subkeys).
// - Slices (bound by either indexed keys or split strings).
//
// Errors:
// - Returns ErrNotExist if the property is missing without a default.
// - Returns type conversion errors if parsing fails.
// - Returns wrapped errors with context (path, type).
func BindValue(p flatten.Storage, v reflect.Value, t reflect.Type, param BindParam, filter Filter) (RetErr error) {

	if !typeutil.IsPropBindingTarget(t) {
		err := errutil.Explain(nil, "target should be value type")
		return errutil.Explain(err, "bind path=%s type=%s error", param.Path, v.Type().String())
	}

	// run validation if "expr" tag is defined and no prior error
	defer func() {
		if RetErr == nil {
			tag, ok := param.Validate.Lookup("expr")
			if ok && len(tag) > 0 {
				if RetErr = validateField(tag, v.Interface()); RetErr != nil {
					RetErr = errutil.Explain(RetErr, "validate path=%s type=%s error", param.Path, v.Type().String())
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
		err := errutil.Explain(nil, "use slice instead of array")
		return errutil.Explain(err, "bind path=%s type=%s error", param.Path, v.Type().String())
	default: // for linter
	}

	fn := converters[t]
	if fn == nil && v.Kind() == reflect.Struct {
		if err := bindStruct(p, v, t, param, filter); err != nil {
			return err // no wrap
		}
		return nil
	}

	// resolve property value (with default and references)
	val, err := resolve(p, param)
	if err != nil {
		return errutil.Explain(err, "bind path=%s type=%s error", param.Path, v.Type().String())
	}

	// try converter function first
	if fn != nil {
		fnValue := reflect.ValueOf(fn)
		out := fnValue.Call([]reflect.Value{reflect.ValueOf(strings.TrimSpace(val))})
		if !out[1].IsNil() {
			err = out[1].Interface().(error)
			return errutil.Explain(err, "bind path=%s type=%s error", param.Path, v.Type().String())
		}
		v.Set(out[0])
		return nil
	}

	// fallback: parse string into basic types
	switch v.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var u uint64
		if u, err = strconv.ParseUint(val, 0, 0); err == nil {
			v.SetUint(u)
			return nil
		}
		return errutil.Explain(err, "bind path=%s type=%s error", param.Path, v.Type().String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var i int64
		if i, err = strconv.ParseInt(val, 0, 0); err == nil {
			v.SetInt(i)
			return nil
		}
		return errutil.Explain(err, "bind path=%s type=%s error", param.Path, v.Type().String())
	case reflect.Float32, reflect.Float64:
		var f float64
		if f, err = strconv.ParseFloat(val, 64); err == nil {
			v.SetFloat(f)
			return nil
		}
		return errutil.Explain(err, "bind path=%s type=%s error", param.Path, v.Type().String())
	case reflect.Bool:
		var b bool
		if b, err = strconv.ParseBool(val); err == nil {
			v.SetBool(b)
			return nil
		}
		return errutil.Explain(err, "bind path=%s type=%s error", param.Path, v.Type().String())
	default:
		// treat everything else as string
		v.SetString(val)
		return nil
	}
}

// bindSlice binds configuration values into a slice of type []T.
//
// Supported input formats:
//  1. Indexed keys in the property source:
//     e.g. "list[0]=a", "list[1]=b"
//  2. A single delimited string:
//     e.g. "list=a,b,c"  (split by ",")
//
// The slice is always reset (v.Set(slice)) before return,
// even if binding fails midway.
func bindSlice(p flatten.Storage, v reflect.Value, t reflect.Type, param BindParam, filter Filter) error {

	elemType := t.Elem()
	p, err := getSlice(p, elemType, param)
	if err != nil {
		return errutil.Explain(err, "bind path=%s type=%s error", param.Path, v.Type().String())
	}

	slice := reflect.MakeSlice(t, 0, 0)
	defer func() { v.Set(slice) }()

	if p == nil {
		return nil
	}

	for i := 0; ; i++ {
		subValue := reflect.New(elemType).Elem()
		subParam := BindParam{
			Key:  fmt.Sprintf("%s[%d]", param.Key, i),
			Path: fmt.Sprintf("%s[%d]", param.Path, i),
		}
		err = BindValue(p, subValue, elemType, subParam, filter)
		if errors.Is(err, ErrNotExist) {
			// stop when no more indexed elements
			break
		}
		if err != nil {
			return errutil.Explain(err, "bind path=%s type=%s error", param.Path, v.Type().String())
		}
		slice = reflect.Append(slice, subValue)
	}
	return nil
}

// getSlice prepares a Storage object representing slice elements
// derived from either:
//
// - Explicit indexed properties (preferred).
// - A single delimited string property, split into multiple elements.
//
// Errors:
// - ErrNotExist if property is missing and no default is provided.
// - Converter missing for non-primitive element types.
func getSlice(p flatten.Storage, et reflect.Type, param BindParam) (flatten.Storage, error) {

	m := make(map[string]string)
	if p.SliceEntries(param.Key, m) {
		return flatten.NewPropertiesStorage(flatten.NewProperties(m)), nil
	}

	// case 2: property is a single string -> split into slice
	strVal, ok := p.Value(param.Key)
	if !ok {
		if !param.Tag.HasDef {
			return nil, errutil.Explain(ErrNotExist, "property %q", param.Key)
		}
		if param.Tag.Def == "" {
			return nil, nil
		}
		if !typeutil.IsPrimitiveValueType(et) && converters[et] == nil {
			return nil, errutil.Explain(nil, "can't find converter for %s", et.String())
		}
		strVal = param.Tag.Def
	}
	if strVal == "" {
		return nil, nil
	}

	arrVal := strings.Split(strVal, ",")
	for i := range arrVal {
		arrVal[i] = strings.TrimSpace(arrVal[i])
	}

	m = make(map[string]string)
	for i, s := range arrVal {
		k := fmt.Sprintf("%s[%d]", param.Key, i)
		m[k] = s
	}
	return flatten.NewPropertiesStorage(flatten.NewProperties(m)), nil
}

// bindMap binds configuration properties into a Go map[K]V.
//
// Example:
//
//	Storage:
//	  "users.alice.age" = 20
//	  "users.bob.age"   = 30
//
//	Binding into map[string]User produces:
//	  {"alice": User{Age:20}, "bob": User{Age:30}}
//
// Errors:
// - Returns error if property is missing without default.
// - Propagates binding errors from element binding.
func bindMap(p flatten.Storage, v reflect.Value, t reflect.Type, param BindParam, filter Filter) error {

	if param.Tag.HasDef && param.Tag.Def != "" {
		err := errutil.Explain(nil, "map can't have a non-empty default value")
		return errutil.Explain(err, "bind path=%s type=%s error", param.Path, v.Type().String())
	}

	elemType := t.Elem()
	ret := reflect.MakeMap(t)
	defer func() { v.Set(ret) }()

	// handle empty key as default value placeholder
	if param.Tag.Key == "" {
		if param.Tag.HasDef {
			return nil
		}
	}

	// ensure property exists
	keySet := make(map[string]struct{})
	p.MapKeys(param.Key, keySet)
	if len(keySet) == 0 {
		if param.Tag.HasDef {
			return nil
		}
		return errutil.Explain(ErrNotExist, "map property %q", param.Key)
	}

	for key := range keySet {
		subValue := reflect.New(elemType).Elem()
		subKey := key
		if param.Key != "" {
			subKey = param.Key + "." + key
		}
		subParam := BindParam{
			Key:  subKey,
			Path: param.Path,
		}
		if err := BindValue(p, subValue, elemType, subParam, filter); err != nil {
			return err // no wrap
		}
		ret.SetMapIndex(reflect.ValueOf(key), subValue)
	}
	return nil
}

// bindStruct binds configuration properties into a struct.
//
// Example:
//
//	type Config struct {
//	    Host string `value:"${db.host:=localhost}"`
//	    Port int    `value:"${db.port:=3306}"`
//	}
//
//	With properties:
//	  db.host=127.0.0.1
//	Result:
//	  Config{Host:"127.0.0.1", Port:3306}
//
// Errors:
// - Invalid syntax in tag.
// - Binding or conversion failures in nested fields.
// - Infinite recursion is avoided for embedded pointer structs.
func bindStruct(p flatten.Storage, v reflect.Value, t reflect.Type, param BindParam, filter Filter) error {

	if param.Tag.HasDef && param.Tag.Def != "" {
		err := errutil.Explain(nil, "struct can't have a non-empty default value")
		return errutil.Explain(err, "bind path=%s type=%s error", param.Path, v.Type().String())
	}

	for i := range t.NumField() {
		ft := t.Field(i)
		fv := v.Field(i)

		// skip unexported fields
		if !fv.CanInterface() {
			continue
		}

		subParam := BindParam{
			Key:  param.Key,
			Path: param.Path + "." + ft.Name,
		}

		if tag, ok := ft.Tag.Lookup("value"); ok {
			if err := subParam.BindTag(tag, ft.Tag); err != nil {
				return errutil.Explain(err, "bind path=%s type=%s error", param.Path, v.Type().String())
			}
			if filter != nil {
				ret, err := filter.Do(fv.Addr().Interface(), subParam)
				if err != nil {
					return errutil.Explain(err, "bind path=%s type=%s error", param.Path, v.Type().String())
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

// resolve fetches the final string value of a property key,
// applying default values and resolving references recursively.
//
// Example:
//
//	Storage:
//	  "host" = "localhost"
//	  "url"  = "http://${host}:8080"
//
//	resolve(url) -> "http://localhost:8080"
func resolve(p flatten.Storage, param BindParam) (string, error) {
	if val, ok := p.Value(param.Key); ok {
		return ResolveString(p, val)
	}
	//if p.Exists(param.Key) {
	//	return "", errutil.Explain(nil, "property %q isn't simple value", param.Key)
	//}
	if param.Tag.HasDef {
		return ResolveString(p, param.Tag.Def)
	}
	return "", errutil.Explain(ErrNotExist, "property %q", param.Key)
}
