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
	"reflect"
	"strings"
	"time"

	"github.com/go-spring/spring-core/conf/provider"
	"github.com/go-spring/spring-core/conf/reader"
	"github.com/go-spring/stdlib/errutil"
	"github.com/go-spring/stdlib/flatten"
	"github.com/spf13/cast"
)

var converters = map[reflect.Type]any{}

func init() {
	RegisterConverter(func(s string) (time.Time, error) { return cast.ToTimeE(s) })
	RegisterConverter(func(s string) (time.Duration, error) { return time.ParseDuration(s) })
}

// Converter converts a string to a target type T.
type Converter[T any] func(string) (T, error)

// RegisterConverter registers a Converter for a type T, such as
// time.Time, time.Duration, or other user-defined types.
// Must be called in init functions only.
func RegisterConverter[T any](fn Converter[T]) {
	t := reflect.TypeFor[T]()
	if _, ok := converters[t]; ok {
		panic("converter for type " + t.String() + " already exists")
	}
	converters[t] = fn
}

// RegisterReader registers its Reader for some kind of file extension.
// Must be called in init functions only.
func RegisterReader(r reader.Reader, ext ...string) {
	reader.Register(r, ext...)
}

// RegisterProvider registers a Provider for a specific configuration source.
// Must be called in init functions only.
func RegisterProvider(name string, p provider.Provider) {
	provider.Register(name, p)
}

// Load creates a Properties instance from a configuration source.
// The source format is [optional:]<provider>:<path> or just <path>.
// Returns an error if the file type is not supported or parsing fails.
func Load(source string) (*flatten.Properties, error) {
	data, err := provider.Load(source)
	if err != nil {
		return nil, err
	}
	return flatten.NewProperties(data), nil
}

// Bind maps property values from storage into the target object.
// The target must be a pointer to a struct.
// An optional tag specifies the root property key using ${key} syntax.
// Supports default values: ${key:=default}.
// If no tag is provided, uses ${ROOT} (binds from the root).
func Bind(p flatten.Storage, i any, tag ...string) error {

	var v reflect.Value
	{
		switch refVal := i.(type) {
		case reflect.Value:
			v = refVal
		default:
			v = reflect.ValueOf(i)
			if v.Kind() != reflect.Pointer {
				return errutil.Explain(nil, "target should be a pointer but got %T", i)
			}
			v = v.Elem()
		}
	}

	t := v.Type()
	typeName := t.Name()
	if typeName == "" { // primitive types have no name
		typeName = t.String()
	}

	s := "${ROOT}"
	if len(tag) > 0 {
		s = tag[0]
	}

	var param BindParam
	err := param.BindTag(s, "")
	if err != nil {
		return errutil.Explain(err, "bind tag '%s' error", s)
	}
	param.Path = typeName
	return BindValue(p, v, t, param, nil)
}

// Resolve expands property references of the form ${key}
// inside a string, recursively resolving nested expressions.
//
// Supported features:
// - Nested references: e.g. "${outer${inner}}"
// - Default values:    "${key:=fallback}"
// - Arbitrary string concatenation around references.
//
// Example:
//
//	Storage:
//	  "host" = "localhost"
//	  "port" = "8080"
//	Input:
//	  "http://${host}:${port}"
//	Output:
//	  "http://localhost:8080"
//
// Errors:
// - Returns invalid syntax if braces are unbalanced.
func Resolve(p flatten.Storage, s string) (string, error) {

	// If there is no property reference, return the original string.
	start := strings.Index(s, "${")
	if start < 0 {
		return s, nil
	}

	var (
		level = 1
		end   = -1
	)

	// scan for matching closing brace, handling nested references
	for i := start + 2; i < len(s); i++ {
		if s[i] == '$' {
			if i+1 < len(s) && s[i+1] == '{' {
				level++
			}
		} else if s[i] == '}' {
			level--
			if level == 0 {
				end = i
				break
			}
		}
	}

	if end < 0 {
		return "", errutil.Explain(nil, "invalid syntax tag '%s'", s)
	}

	var param BindParam
	err := param.BindTag(s[start:end+1], "")
	if err != nil {
		return "", errutil.Explain(err, "resolve string %q error", s)
	}

	// resolve the referenced property
	resolved, err := resolve(p, param)
	if err != nil {
		return "", errutil.Explain(err, "resolve string %q error", s)
	}

	// resolve the remaining part of the string
	suffix, err := Resolve(p, s[end+1:])
	if err != nil {
		return "", errutil.Explain(err, "resolve string %q error", s)
	}

	// combine: prefix + resolved + suffix
	return s[:start] + resolved + suffix, nil
}
