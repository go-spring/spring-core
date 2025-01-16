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

package gs_core_test

import (
	"reflect"
	"testing"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_core"
	"github.com/go-spring/spring-core/util/assert"
	"go.uber.org/mock/gomock"
)

func TestBind(t *testing.T) {

	t.Run("zero argument", func(t *testing.T) {
		c := container(t, nil)
		stack := gs_core.NewWiringStack()
		ctx := gs_core.NewArgContext(c.(*gs_core.Container), stack)
		fn := func() {}
		p, err := gs_arg.Bind(fn, []gs.Arg{}, 1)
		if err != nil {
			t.Fatal(err)
		}
		values, err := p.Call(ctx)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(values), 0)
	})

	t.Run("one value argument", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_core.NewArgMockContext(ctrl)
		expectInt := 0
		fn := func(i int) {
			expectInt = i
		}
		c, err := gs_arg.Bind(fn, []gs.Arg{
			gs_arg.Value(3),
		}, 1)
		if err != nil {
			t.Fatal(err)
		}
		values, err := c.Call(ctx)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, expectInt, 3)
		assert.Equal(t, len(values), 0)
	})

	t.Run("one ctx value argument", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_core.NewArgMockContext(ctrl)
		ctx.EXPECT().Bind(gomock.Any(), "${a.b.c}").DoAndReturn(func(v, tag interface{}) error {
			v.(reflect.Value).SetInt(3)
			return nil
		})
		expectInt := 0
		fn := func(i int) {
			expectInt = i
		}
		c, err := gs_arg.Bind(fn, []gs.Arg{
			"${a.b.c}",
		}, 1)
		if err != nil {
			t.Fatal(err)
		}
		values, err := c.Call(ctx)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, expectInt, 3)
		assert.Equal(t, len(values), 0)
	})

	t.Run("one ctx named bean argument", func(t *testing.T) {
		type st struct {
			i int
		}
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_core.NewArgMockContext(ctrl)
		ctx.EXPECT().Wire(gomock.Any(), "a").DoAndReturn(func(v, tag interface{}) error {
			v.(reflect.Value).Set(reflect.ValueOf(&st{3}))
			return nil
		})
		expectInt := 0
		fn := func(v *st) {
			expectInt = v.i
		}
		c, err := gs_arg.Bind(fn, []gs.Arg{
			"a",
		}, 1)
		if err != nil {
			t.Fatal(err)
		}
		values, err := c.Call(ctx)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, expectInt, 3)
		assert.Equal(t, len(values), 0)
	})

	t.Run("one ctx unnamed bean argument", func(t *testing.T) {
		type st struct {
			i int
		}
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_core.NewArgMockContext(ctrl)
		ctx.EXPECT().Wire(gomock.Any(), "").DoAndReturn(func(v, tag interface{}) error {
			v.(reflect.Value).Set(reflect.ValueOf(&st{3}))
			return nil
		})
		expectInt := 0
		fn := func(v *st) {
			expectInt = v.i
		}
		c, err := gs_arg.Bind(fn, []gs.Arg{}, 1)
		if err != nil {
			t.Fatal(err)
		}
		values, err := c.Call(ctx)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, expectInt, 3)
		assert.Equal(t, len(values), 0)
	})

}
