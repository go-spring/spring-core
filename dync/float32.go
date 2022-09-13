/*
 * Copyright 2012-2019 the original author or authors.
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

	"github.com/go-spring/spring-base/atomic"
	"github.com/go-spring/spring-base/cast"
	"github.com/go-spring/spring-core/conf"
)

type Float32ValidateFunc func(v float32) error

type Float32 struct {
	Base
	v atomic.Float32
	f Float32ValidateFunc
}

func (x *Float32) Value() float32 {
	return x.v.Load()
}

func (x *Float32) OnValidate(f Float32ValidateFunc) {
	x.f = f
}

func (x *Float32) getFloat32(prop *conf.Properties) (float32, error) {
	s, err := x.Property(prop)
	if err != nil {
		return 0, err
	}
	v, err := cast.ToFloat64E(s)
	if err != nil {
		return 0, err
	}
	return float32(v), nil
}

func (x *Float32) onRefresh(prop *conf.Properties) error {
	v, err := x.getFloat32(prop)
	if err != nil {
		return err
	}
	x.v.Store(v)
	return nil
}

func (x *Float32) onValidate(prop *conf.Properties) error {
	v, err := x.getFloat32(prop)
	if err != nil {
		return err
	}
	err = x.Validate(v)
	if err != nil {
		return err
	}
	if x.f != nil {
		return x.f(v)
	}
	return nil
}

func (x *Float32) MarshalJSON() ([]byte, error) {
	return json.Marshal(x.Value())
}