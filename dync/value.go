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

package dync

import (
	"encoding/json"
	"reflect"

	"github.com/go-spring/spring-core/conf"
	"go.uber.org/atomic"
)

type Value[T interface{}] struct {
	v atomic.Value
}

func bindValue[T any](o T, prop conf.ReadOnlyProperties, param conf.BindParam) error {
	t := reflect.TypeOf(o).Elem()
	v := reflect.ValueOf(o).Elem()
	return conf.BindValue(prop, v, t, param, nil)
}

func (r *Value[T]) OnRefresh(prop conf.ReadOnlyProperties, param conf.BindParam) error {
	var o T
	err := bindValue(&o, prop, param)
	if err != nil {
		return err
	}
	r.v.Store(o)
	return nil
}

func (r *Value[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.v.Load())
}
