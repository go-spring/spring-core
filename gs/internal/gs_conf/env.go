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
	"strings"

	"github.com/go-spring/spring-core/conf"
)

// Environment represents the environment configuration.
type Environment struct{}

// NewEnvironment initializes a new instance of Environment.
func NewEnvironment() *Environment {
	return &Environment{}
}

// CopyTo add environment variables that matches IncludeEnvPatterns and
// exclude environment variables that matches ExcludeEnvPatterns.
func (c *Environment) CopyTo(p *conf.MutableProperties) error {
	environ := os.Environ()
	if len(environ) == 0 {
		return nil
	}
	const prefix = "GS_"
	for _, env := range environ {
		ss := strings.SplitN(env, "=", 2)
		k, v := ss[0], ""
		if len(ss) > 1 {
			v = ss[1]
		}
		var propKey string
		if strings.HasPrefix(k, prefix) {
			propKey = strings.TrimPrefix(k, prefix)
			propKey = strings.ReplaceAll(propKey, "_", ".")
			propKey = strings.ToLower(propKey)
		} else {
			propKey = k
		}
		if err := p.Set(propKey, v); err != nil {
			return err
		}
	}
	return nil
}
