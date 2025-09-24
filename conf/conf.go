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

	value:"${key:=default}>>splitter"

Key features:
- Nested keys (e.g., service.endpoint)
- Dynamic defaults (e.g., ${DB_HOST:=localhost:${DB_PORT:=3306}})
- Splitter chaining (e.g., >>json for JSON parsing)

# Data Binding:

Supports binding to various types with automatic conversion:

1. Primitives: Uses strconv for basic type conversions
2. Complex Types:
  - Slices: From indexed properties or custom splitters
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

1. RegisterSplitter: Add custom string splitters
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
	    Endpoints []string `value:"${endpoints}"`       // Indexed properties
	    Features  []string `value:"${features}>>split"` // Custom splitter
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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/go-spring/spring-base/barky"
	"github.com/go-spring/spring-core/conf/reader/json"
	"github.com/go-spring/spring-core/conf/reader/prop"
	"github.com/go-spring/spring-core/conf/reader/toml"
	"github.com/go-spring/spring-core/conf/reader/yaml"
	"github.com/spf13/cast"
)

var (
	readers    = map[string]Reader{}
	splitters  = map[string]Splitter{}
	converters = map[reflect.Type]any{}
)

func init() {

	// built-in readers
	RegisterReader(json.Read, ".json")
	RegisterReader(prop.Read, ".properties")
	RegisterReader(yaml.Read, ".yaml", ".yml")
	RegisterReader(toml.Read, ".toml", ".tml")

	// time.Time
	RegisterConverter(func(s string) (time.Time, error) {
		return cast.ToTimeE(strings.TrimSpace(s))
	})

	// time.Duration
	RegisterConverter(func(s string) (time.Duration, error) {
		return time.ParseDuration(strings.TrimSpace(s))
	})
}

// Reader parses raw bytes into a nested map[string]any.
type Reader func(b []byte) (map[string]any, error)

// RegisterReader registers its Reader for some kind of file extension.
func RegisterReader(r Reader, ext ...string) {
	for _, s := range ext {
		readers[s] = r
	}
}

// Splitter splits a string into a slice of strings using custom logic.
type Splitter func(string) ([]string, error)

// RegisterSplitter registers a Splitter with a given name.
func RegisterSplitter(name string, fn Splitter) {
	splitters[name] = fn
}

// Converter converts a string to a target type T.
type Converter[T any] func(string) (T, error)

// RegisterConverter registers a Converter for a non-primitive type such as
// time.Time, time.Duration, or other user-defined value types.
func RegisterConverter[T any](fn Converter[T]) {
	t := reflect.TypeFor[T]()
	converters[t] = fn
}

// Properties defines the read-only interface for accessing configuration data.
type Properties interface {
	// Data returns all key-value pairs as a flat map.
	Data() map[string]string
	// Keys returns all keys.
	Keys() []string
	// SubKeys returns the sorted sub-keys of a given key.
	SubKeys(key string) ([]string, error)
	// Has checks whether a key exists.
	Has(key string) bool
	// Get returns the value for a given key, with an optional default.
	Get(key string, def ...string) string
	// Resolve resolves placeholders inside a string (e.g. ${key:=default}).
	Resolve(s string) (string, error)
	// Bind binds property values into a target object (struct, map, slice, or primitive).
	Bind(i any, tag ...string) error
	// CopyTo copies properties into another instance, overriding existing values.
	CopyTo(out *MutableProperties) error
}

var _ Properties = (*MutableProperties)(nil)

// MutableProperties stores the data with map[string]string and the keys are case-sensitive,
// you can get one of them by its key, or bind some of them to a value.
//
// There are too many formats of configuration files, and too many conflicts between
// them. Each format of configuration file provides its special characteristics, but
// usually they are not all necessary, and complementary. For example, `conf` disabled
// Java properties' expansion when reading file, but also provides similar function
// when getting or binding properties.
//
// A good rule of thumb is that treating application configuration as a tree, but not
// all formats of configuration files designed as a tree or not ideal, for instance
// Java properties isn't strictly verified. Although configuration can store as a tree,
// but it costs more CPU time when getting properties because it reads property node
// by node. So `conf` uses a tree to strictly verify and a flat map to store.
type MutableProperties struct {
	*barky.Storage
}

// New creates a new empty MutableProperties instance.
func New() *MutableProperties {
	return &MutableProperties{
		Storage: barky.NewStorage(),
	}
}

// Load creates a MutableProperties instance from a configuration file.
// Returns an error if the file type is not supported or parsing fails.
func Load(file string) (*MutableProperties, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	ext := filepath.Ext(file)
	r, ok := readers[ext]
	if !ok {
		return nil, fmt.Errorf("unsupported file type %s", ext)
	}
	m, err := r(b)
	if err != nil {
		return nil, err
	}
	p := New()
	_ = p.merge(barky.FlattenMap(m), file)
	return p, nil
}

// Map creates a MutableProperties instance directly from a map.
func Map(data map[string]any) *MutableProperties {
	p := New()
	_, file, _, _ := runtime.Caller(1)
	_ = p.merge(barky.FlattenMap(data), file)
	return p
}

// merge flattens the map and sets all keys and values.
func (p *MutableProperties) merge(m map[string]string, file string) error {
	fileID := p.AddFile(file)
	for key, val := range m {
		if err := p.Set(key, val, fileID); err != nil {
			return err
		}
	}
	return nil
}

// Resolve resolves placeholders in a string, replacing references like
// ${key:=default} with their actual values from the properties.
func (p *MutableProperties) Resolve(s string) (string, error) {
	return resolveString(p, s)
}

// Bind maps property values into the provided target object.
// Supported targets: primitive values, maps, slices, and structs.
// Struct binding uses the `value` tag in the form:
//
//	value:"${key:=default}>>splitter"
//
// - key: property key
// - default: default value if key is missing
// - splitter: registered splitter name for splitting into []string
func (p *MutableProperties) Bind(i any, tag ...string) error {

	var v reflect.Value
	{
		switch refVal := i.(type) {
		case reflect.Value:
			v = refVal
		default:
			v = reflect.ValueOf(i)
			if v.Kind() != reflect.Ptr {
				return errors.New("should be a ptr")
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
		return err
	}
	param.Path = typeName
	return BindValue(p, v, t, param, nil)
}

// CopyTo copies all properties into another MutableProperties instance,
// overriding values if keys already exist.
func (p *MutableProperties) CopyTo(out *MutableProperties) error {
	rawFile := p.RawFile()
	newfile := make(map[string]int8)
	oldFile := make([]string, len(rawFile))
	for k, v := range rawFile {
		oldFile[v] = k
		newfile[k] = out.AddFile(k)
	}
	for key, v := range p.RawData() {
		fileID := newfile[oldFile[v.File]]
		if err := out.Set(key, v.Value, fileID); err != nil {
			return err
		}
	}
	return nil
}
