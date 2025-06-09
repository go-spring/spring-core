/*
 * Copyright 2025 The Go-Spring Authors.
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

package log

import (
	"strings"
	"testing"
)

func TestIsValidTag(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want bool
	}{
		{"valid_1segments", "_def", true},
		{"valid_4segments", "service_module_submodule_component00", true},
		{"too_short_3", "abc", false},
		{"too_long_37", strings.Repeat("a", 37), false},
		{"uppercase", "Invalid_Tag", false},
		{"special_char", "tag!name", false},
		{"space", "tag name", false},
		{"hyphen", "tag-name", false},
		{"too_many_segments", "a_b_c_d_e", false},
		{"leading_underscore_1", "_service_component", true},
		{"leading_underscore_2", "__service_component", false},
		{"trailing_underscore_1", "service_component_", false},
		{"trailing_underscore_2", "service_component__", false},
		{"consecutive_underscore", "service__component", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidTag(tt.tag); got != tt.want {
				t.Errorf("isValidTag(%q) = %v, want %v", tt.tag, got, tt.want)
			}
		})
	}
}
