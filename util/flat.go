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

package util

import (
	"fmt"
	"reflect"

	"github.com/spf13/cast"
)

// FlattenMap flattens a nested map, array, or slice into a single-level map
// with string keys and string values. It recursively processes each element
// of the input map and adds its flattened representation to the result map.
func FlattenMap(m map[string]any) map[string]string {
	result := make(map[string]string)
	for key, val := range m {
		FlattenValue(key, val, result)
	}
	return result
}

// FlattenValue flattens a single value (which can be a map, array, slice,
// or other types) into the result map.
func FlattenValue(key string, val any, result map[string]string) {
	if val == nil {
		return
	}
	switch v := reflect.ValueOf(val); v.Kind() {
	case reflect.Map:
		if v.Len() == 0 {
			result[key] = ""
			return
		}
		iter := v.MapRange()
		for iter.Next() {
			mapKey := cast.ToString(iter.Key().Interface())
			mapValue := iter.Value().Interface()
			FlattenValue(key+"."+mapKey, mapValue, result)
		}
	case reflect.Array, reflect.Slice:
		if v.Len() == 0 {
			result[key] = ""
			return
		}
		for i := range v.Len() {
			subKey := fmt.Sprintf("%s[%d]", key, i)
			subValue := v.Index(i).Interface()
			// If an element is nil, treat it as an empty value and assign an empty string.
			// Note: We do not remove the nil element to avoid changing the array's size.
			if subValue == nil {
				result[subKey] = ""
				continue
			}
			FlattenValue(subKey, subValue, result)
		}
	default:
		result[key] = cast.ToString(val)
	}
}
