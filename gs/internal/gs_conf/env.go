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

package gs_conf

import (
	"os"
	"regexp"
	"strings"

	"github.com/go-spring/spring-core/conf"
)

const (
	EnvironmentPrefix  = "GS_ENVS_PREFIX"
	IncludeEnvPatterns = "INCLUDE_ENV_PATTERNS"
	ExcludeEnvPatterns = "EXCLUDE_ENV_PATTERNS"
)

// Environment represents the environment configuration.
type Environment struct{}

// NewEnvironment initializes a new instance of Environment.
func NewEnvironment() *Environment {
	return &Environment{}
}

// lookupEnv searches for an environment variable by key in the environ slice.
func lookupEnv(environ []string, key string) (value string, found bool) {
	key = strings.TrimSpace(key) + "="
	for _, s := range environ {
		if strings.HasPrefix(s, key) {
			v := strings.TrimPrefix(s, key)
			return strings.TrimSpace(v), true
		}
	}
	return "", false
}

// CopyTo add environment variables that matches IncludeEnvPatterns and
// exclude environment variables that matches ExcludeEnvPatterns.
func (c *Environment) CopyTo(p *conf.MutableProperties) error {
	environ := os.Environ()
	if len(environ) == 0 {
		return nil
	}

	prefix := "GS_"
	if s := strings.TrimSpace(os.Getenv(EnvironmentPrefix)); s != "" {
		prefix = s
	}

	toRex := func(patterns []string) ([]*regexp.Regexp, error) {
		var rex []*regexp.Regexp
		for _, v := range patterns {
			exp, err := regexp.Compile(v)
			if err != nil {
				return nil, err
			}
			rex = append(rex, exp)
		}
		return rex, nil
	}

	includes := []string{".*"}
	if s, ok := lookupEnv(environ, IncludeEnvPatterns); ok {
		includes = strings.Split(s, ",")
	}
	includeRex, err := toRex(includes)
	if err != nil {
		return err
	}

	var excludes []string
	if s, ok := lookupEnv(environ, ExcludeEnvPatterns); ok {
		excludes = strings.Split(s, ",")
	}
	excludeRex, err := toRex(excludes)
	if err != nil {
		return err
	}

	matches := func(rex []*regexp.Regexp, s string) bool {
		for _, r := range rex {
			if r.MatchString(s) {
				return true
			}
		}
		return false
	}

	for _, env := range environ {
		ss := strings.SplitN(env, "=", 2)
		k, v := ss[0], ""
		if len(ss) > 1 {
			v = ss[1]
		}

		var propKey string
		if strings.HasPrefix(k, prefix) {
			propKey = strings.TrimPrefix(k, prefix)
			propKey = strings.ToLower(replaceKey(propKey))
		} else if matches(includeRex, k) && !matches(excludeRex, k) {
			propKey = k
		} else {
			continue
		}

		if err = p.Set(propKey, v); err != nil {
			return err
		}
	}
	return nil
}

// replaceKey replace '_' with '.'
func replaceKey(s string) string {
	b := make([]byte, len(s)+2)
	b[0] = '_'
	b[len(b)-1] = '_'
	copy(b[1:len(b)-1], s)
	for i := 1; i < len(b)-1; i++ {
		if b[i] == '_' {
			if b[i-1] != '_' && b[i+1] != '_' {
				b[i] = '.'
			}
		}
	}
	return string(b[1 : len(b)-1])
}
