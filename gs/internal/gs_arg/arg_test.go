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

package gs_arg_test

import (
	"reflect"
	"testing"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/gsmock"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/util/assert"
)

func TestTagArg(t *testing.T) {
	r, _ := gsmock.Init(t.Context())
	c := gs.NewMockArgContext(r)
	c.MockBind().Handle(func(value reflect.Value, s string) (error, bool) {
		return conf.New().Bind(value, s), true
	})
	tag := gs_arg.Tag("${int:=3}")
	v, err := tag.GetArgValue(c, reflect.TypeFor[string]())
	assert.Nil(t, err)
	assert.Equal(t, v.String(), "3")
}
