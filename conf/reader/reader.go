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

package reader

import (
	"os"
	"path/filepath"

	"github.com/go-spring/spring-core/conf/reader/json"
	"github.com/go-spring/spring-core/conf/reader/prop"
	"github.com/go-spring/spring-core/conf/reader/toml"
	"github.com/go-spring/spring-core/conf/reader/yaml"
	"github.com/go-spring/stdlib/errutil"
)

var readers = map[string]Reader{}

func init() {
	Register(json.Read, ".json")
	Register(prop.Read, ".properties")
	Register(yaml.Read, ".yaml", ".yml")
	Register(toml.Read, ".toml", ".tml")
}

// Reader parses raw bytes into a nested map[string]any.
type Reader func(b []byte) (map[string]any, error)

// Register registers its Reader for some kind of file extension.
func Register(r Reader, ext ...string) {
	for _, s := range ext {
		readers[s] = r
	}
}

// ReadFile reads a file and parses its content into a map[string]any.
func ReadFile(file string) (map[string]any, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, errutil.Explain(err, "read file %s error", file)
	}
	ext := filepath.Ext(file)
	r, ok := readers[ext]
	if !ok {
		err = errutil.Explain(nil, "unsupported file type %s", ext)
		return nil, errutil.Explain(err, "read file %s error", file)
	}
	m, err := r(b)
	if err != nil {
		return nil, errutil.Explain(err, "read file %s error", file)
	}
	return m, nil
}
