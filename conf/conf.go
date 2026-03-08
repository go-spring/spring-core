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

/*
Package conf provides a configuration binding framework with hierarchical resolution,
type-safe mapping, and validation capabilities.

# Core Concepts:

The framework enables mapping configuration data from multiple sources into Go structures with:

- Hierarchical property resolution using ${key} syntax
- Type-safe binding with automatic conversions
- Expression-based validation
- Extensible architecture via pluggable components

# Tag Syntax:

Struct tags use the following format:

	value:"${key:=default}"

Key features:
- Nested keys (e.g., service.endpoint)
- Dynamic defaults (e.g., ${DB_HOST:=localhost:${DB_PORT:=3306}})

# Data Binding:

Supports binding to various types with automatic conversion:

1. Primitives: Uses strconv for basic type conversions
2. Complex Types:
  - Slices: From indexed properties
  - Maps: Via subkey expansion
  - Structs: Recursive binding of nested structures

3. Custom Types: Register converters using RegisterConverter

# Validation System:

 1. Expression validation using expr tag:
    type Config struct {
    Port int `expr:"$ > 0 && $ < 65535"`
    }

 2. Custom validators:
    RegisterValidateFunc("futureDate", func(t time.Time) bool {
    return t.After(time.Now())
    })

# File Support:

Built-in readers handle:
- JSON (.json)
- Properties (.properties)
- YAML (.yaml/.yml)
- TOML (.toml/.tml)

Register custom readers with RegisterReader.

# Property Resolution:

- Recursive ${} substitution
- Type-aware defaults
- Chained defaults (${A:=${B:=C}})

# Extension Points:

2. RegisterConverter: Add type converters
3. RegisterReader: Support new file formats
4. RegisterValidateFunc: Add custom validators

# Examples:

Basic binding:

	type ServerConfig struct {
	    Host string `value:"${host:=localhost}"`
	    Port int    `value:"${port:=8080}"`
	}

Nested structure:

	type AppConfig struct {
	    DB      Database `value:"${db}"`
	    Timeout string   `value:"${timeout:=5s}"`
	}

Slice binding:

	type Config struct {
	    Endpoints []string `value:"${endpoints}"`
	    Features  []string `value:"${features}"`
	}

Validation:

	type UserConfig struct {
	    Age     int       `value:"${age}" expr:"$ >= 18"`
	    Email   string    `value:"${email}" expr:"contains($, '@')"`
	    Expires time.Time `value:"${expires}" expr:"futureDate($)"`
	}
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

var (
	converters = map[reflect.Type]any{}
)

func init() {
	RegisterConverter(func(s string) (time.Time, error) { return cast.ToTimeE(s) })
	RegisterConverter(func(s string) (time.Duration, error) { return time.ParseDuration(s) })
}

// RegisterReader registers its Reader for some kind of file extension.
func RegisterReader(r reader.Reader, ext ...string) {
	reader.Register(r, ext...)
}

// RegisterProvider registers a Provider for a specific configuration source.
func RegisterProvider(name string, p provider.Provider) {
	provider.Register(name, p)
}

// Converter converts a string to a target type T.
type Converter[T any] func(string) (T, error)

// RegisterConverter registers a Converter for a non-primitive type such as
// time.Time, time.Duration, or other user-defined value types.
func RegisterConverter[T any](fn Converter[T]) {
	t := reflect.TypeFor[T]()
	converters[t] = fn
}

// Load creates a MutableProperties instance from a configuration file.
// Returns an error if the file type is not supported or parsing fails.
func Load(source string) (*flatten.Properties, error) {
	data, err := provider.Load(source)
	if err != nil {
		return nil, err
	}
	return flatten.NewProperties(data), nil
}

// Bind maps property values into the provided target object.
func Bind(p flatten.Storage, i any, tag ...string) error {

	var v reflect.Value
	{
		switch refVal := i.(type) {
		case reflect.Value:
			v = refVal
		default:
			v = reflect.ValueOf(i)
			if v.Kind() != reflect.Pointer {
				return errutil.Explain(nil, "should be a pointer but %T", i)
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
	if len(tag) > 0 && tag[0] != "" {
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

// ResolveString expands property references of the form ${key}
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
// - ErrInvalidSyntax if braces are unbalanced.
// - Propagates errors from resolve().
func ResolveString(p flatten.Storage, s string) (string, error) {

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
		err := ErrInvalidSyntax
		return "", errutil.Explain(err, "resolve string %q error", s)
	}

	var param BindParam
	_ = param.BindTag(s[start:end+1], "")

	// resolve the referenced property
	resolved, err := resolve(p, param)
	if err != nil {
		return "", errutil.Explain(err, "resolve string %q error", s)
	}

	// resolve the remaining part of the string
	suffix, err := ResolveString(p, s[end+1:])
	if err != nil {
		return "", errutil.Explain(err, "resolve string %q error", s)
	}

	// combine: prefix + resolved + suffix
	return s[:start] + resolved + suffix, nil
}
