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

package yaml

import (
	"gopkg.in/yaml.v2"
)

// Read parses []byte in the yaml format into map.
func Read(b []byte) (map[string]any, error) {
	ret := make(map[string]any)
	err := yaml.Unmarshal(b, &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
