/*
 * Copyright 2012-2024 the original author or authors.
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
	"fmt"
	"reflect"

	"github.com/spf13/cast"
)

// FlattenMap can expand the nested array, slice and map.
func FlattenMap(m map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for key, val := range m {
		FlattenValue(key, val, result)
	}
	return result
}

// FlattenValue can expand the nested array, slice and map.
func FlattenValue(key string, val interface{}, result map[string]string) {
	switch v := reflect.ValueOf(val); v.Kind() {
	case reflect.Map:
		if v.Len() == 0 {
			result[key] = ""
			return
		}
		for _, k := range v.MapKeys() {
			mapKey := cast.ToString(k.Interface())
			mapValue := v.MapIndex(k).Interface()
			FlattenValue(key+"."+mapKey, mapValue, result)
		}
	case reflect.Array, reflect.Slice:
		if v.Len() == 0 {
			result[key] = ""
			return
		}
		for i := 0; i < v.Len(); i++ {
			subKey := fmt.Sprintf("%s[%d]", key, i)
			subValue := v.Index(i).Interface()
			FlattenValue(subKey, subValue, result)
		}
	default:
		result[key] = cast.ToString(val)
	}
}
