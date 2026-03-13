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
