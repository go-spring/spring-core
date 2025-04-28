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
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/go-spring/spring-core/conf/reader/json"
	"github.com/go-spring/spring-core/conf/reader/prop"
	"github.com/go-spring/spring-core/conf/reader/toml"
	"github.com/go-spring/spring-core/conf/reader/yaml"
	"github.com/go-spring/spring-core/conf/storage"
	"github.com/go-spring/spring-core/util"
	"github.com/spf13/cast"
)

var (
	readers    = map[string]Reader{}
	splitters  = map[string]Splitter{}
	converters = map[reflect.Type]any{}
)

func init() {

	RegisterReader(json.Read, ".json")
	RegisterReader(prop.Read, ".properties")
	RegisterReader(yaml.Read, ".yaml", ".yml")
	RegisterReader(toml.Read, ".toml", ".tml")

	RegisterConverter(func(s string) (time.Time, error) {
		return cast.ToTimeE(strings.TrimSpace(s))
	})

	RegisterConverter(func(s string) (time.Duration, error) {
		return time.ParseDuration(strings.TrimSpace(s))
	})
}

// Reader parses []byte into nested map[string]any.
type Reader func(b []byte) (map[string]any, error)

// RegisterReader registers its Reader for some kind of file extension.
func RegisterReader(r Reader, ext ...string) {
	for _, s := range ext {
		readers[s] = r
	}
}

// Splitter splits string into []string by some characters.
type Splitter func(string) ([]string, error)

// RegisterSplitter registers a Splitter and named it.
func RegisterSplitter(name string, fn Splitter) {
	splitters[name] = fn
}

// Converter converts a string to a target type T.
type Converter[T any] func(string) (T, error)

// RegisterConverter registers its converter for non-primitive type such as
// time.Time, time.Duration, or other user-defined value type.
func RegisterConverter[T any](fn Converter[T]) {
	t := reflect.TypeOf(fn)
	converters[t.Out(0)] = fn
}

// Properties is the interface for read-only properties.
type Properties interface {
	// Data returns key-value pairs of the properties.
	Data() map[string]string
	// Keys returns keys of the properties.
	Keys() []string
	// SubKeys returns the sorted sub keys of the key.
	SubKeys(key string) ([]string, error)
	// Has returns whether the key exists.
	Has(key string) bool
	// Get returns key's value.
	Get(key string, def ...string) string
	// Resolve resolves string that contains references.
	Resolve(s string) (string, error)
	// Bind binds properties into a value.
	Bind(i any, tag ...string) error
	// CopyTo copies properties into another by override.
	CopyTo(out *MutableProperties) error
}

var _ Properties = (*MutableProperties)(nil)

// MutableProperties stores the data with map[string]string and the keys are case-sensitive,
// you can get one of them by its key, or bind some of them to a value.
// There are too many formats of configuration files, and too many conflicts between
// them. Each format of configuration file provides its special characteristics, but
// usually they are not all necessary, and complementary. For example, `conf` disabled
// Java properties' expansion when reading file, but also provides similar function
// when getting or binding properties.
// A good rule of thumb is that treating application configuration as a tree, but not
// all formats of configuration files designed as a tree or not ideal, for instance
// Java properties isn't strictly verified. Although configuration can store as a tree,
// but it costs more CPU time when getting properties because it reads property node
// by node. So `conf` uses a tree to strictly verify and a flat map to store.
type MutableProperties struct {
	*storage.Storage
}

// New creates empty *MutableProperties.
func New() *MutableProperties {
	return &MutableProperties{
		Storage: storage.NewStorage(),
	}
}

// Load creates *MutableProperties from file.
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
	return Map(m), nil
}

// Map creates *MutableProperties from map.
func Map(m map[string]any) *MutableProperties {
	p := New()
	_ = p.merge(util.FlattenMap(m))
	return p
}

// merge flattens the map and sets all keys and values.
func (p *MutableProperties) merge(m map[string]string) error {
	for key, val := range m {
		if err := p.Set(key, val); err != nil {
			return err
		}
	}
	return nil
}

// Data returns key-value pairs of the properties.
func (p *MutableProperties) Data() map[string]string {
	m := make(map[string]string)
	maps.Copy(m, p.RawData())
	return m
}

// Keys returns keys of the properties.
func (p *MutableProperties) Keys() []string {
	return util.OrderedMapKeys(p.RawData())
}

// Get returns key's value, using Def to return a default value.
func (p *MutableProperties) Get(key string, def ...string) string {
	val, ok := p.RawData()[key]
	if !ok && len(def) > 0 {
		return def[0]
	}
	return val
}

// Resolve resolves string value that contains references to other
// properties, the references are defined by ${key:=def}.
func (p *MutableProperties) Resolve(s string) (string, error) {
	return resolveString(p, s)
}

// Bind binds properties to a value, the bind value can be primitive type,
// map, slice, struct. When binding to struct, the tag 'value' indicates
// which properties should be bind. The 'value' tag are defined by
// value:"${a:=b}>>splitter", 'a' is the key, 'b' is the default value,
// 'splitter' is the Splitter's name when you want split string value
// into []string value.
func (p *MutableProperties) Bind(i any, tag ...string) error {

	var v reflect.Value
	{
		switch e := i.(type) {
		case reflect.Value:
			v = e
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
	if typeName == "" { // primitive type has no name
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

// CopyTo copies properties into another by override.
func (p *MutableProperties) CopyTo(out *MutableProperties) error {
	return out.merge(p.RawData())
}
