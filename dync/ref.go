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

type RefValidateFunc func(v interface{}) error

type Ref[T interface{}] struct {
	v atomic.Value
	f RefValidateFunc
}

func (r *Ref[T]) OnValidate(f RefValidateFunc) {
	r.f = f
}

func bindRef[T any](o T, prop conf.ReadOnlyProperties, param conf.BindParam) error {
	t := reflect.TypeOf(o).Elem()
	v := reflect.ValueOf(o).Elem()
	return conf.BindValue(prop, v, t, param, nil)
}

func (r *Ref[T]) Refresh(prop conf.ReadOnlyProperties, param conf.BindParam) error {
	var o T
	err := bindRef(&o, prop, param)
	if err != nil {
		return err
	}
	r.v.Store(o)
	return nil
}

func (r *Ref[T]) Validate(prop conf.ReadOnlyProperties, param conf.BindParam) error {
	var o T
	err := bindRef(&o, prop, param)
	if err != nil {
		return err
	}
	if r.f != nil {
		return r.f(o)
	}
	return err
}

func (r *Ref[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.v.Load())
}
