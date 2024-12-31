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

	"github.com/go-spring/spring-core/conf"
	"github.com/spf13/cast"
	"go.uber.org/atomic"
)

type Float64ValidateFunc func(v float64) error

type Float64 struct {
	v atomic.Float64
	f Float64ValidateFunc
}

func (x *Float64) Value() float64 {
	return x.v.Load()
}

func (x *Float64) OnValidate(f Float64ValidateFunc) {
	x.f = f
}

func (x *Float64) getFloat64(prop conf.ReadOnlyProperties, param conf.BindParam) (float64, error) {
	s, err := GetProperty(prop, param)
	if err != nil {
		return 0, err
	}
	v, err := cast.ToFloat64E(s)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func (x *Float64) Refresh(prop conf.ReadOnlyProperties, param conf.BindParam) error {
	v, err := x.getFloat64(prop, param)
	if err != nil {
		return err
	}
	x.v.Store(v)
	return nil
}

func (x *Float64) Validate(prop conf.ReadOnlyProperties, param conf.BindParam) error {
	v, err := x.getFloat64(prop, param)
	if err != nil {
		return err
	}
	err = Validate(v, param)
	if err != nil {
		return err
	}
	if x.f != nil {
		return x.f(v)
	}
	return nil
}

func (x *Float64) MarshalJSON() ([]byte, error) {
	return json.Marshal(x.Value())
}
