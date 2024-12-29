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
	"time"

	"github.com/go-spring/spring-core/atomic"
	"github.com/go-spring/spring-core/conf"
	"github.com/spf13/cast"
)

type DurationValidateFunc func(v time.Duration) error

type Duration struct {
	v atomic.Duration
	f DurationValidateFunc
}

func (x *Duration) Value() time.Duration {
	return x.v.Load()
}

func (x *Duration) OnValidate(f DurationValidateFunc) {
	x.f = f
}

func (x *Duration) getDuration(prop *conf.Properties, param conf.BindParam) (time.Duration, error) {
	s, err := GetProperty(prop, param)
	if err != nil {
		return 0, err
	}
	v, err := cast.ToDurationE(s)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func (x *Duration) Refresh(prop *conf.Properties, param conf.BindParam) error {
	v, err := x.getDuration(prop, param)
	if err != nil {
		return err
	}
	x.v.Store(v)
	return nil
}

func (x *Duration) Validate(prop *conf.Properties, param conf.BindParam) error {
	v, err := x.getDuration(prop, param)
	if err != nil {
		return err
	}
	if x.f != nil {
		return x.f(v)
	}
	return nil
}

func (x *Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(x.Value())
}
