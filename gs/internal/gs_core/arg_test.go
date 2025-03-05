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

package gs_core_test

import (
	"testing"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_arg"
	"github.com/go-spring/spring-core/gs/internal/gs_core/wiring"
	"github.com/go-spring/spring-core/util/assert"
)

func TestBind(t *testing.T) {

	t.Run("zero argument", func(t *testing.T) {
		stack := wiring.NewStack()
		ctx := wiring.NewArgContext(&wiring.Wiring{}, stack)
		fn := func() {}
		p, err := gs_arg.Bind(fn, []gs.Arg{})
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
		stack := wiring.NewStack()
		ctx := wiring.NewArgContext(&wiring.Wiring{}, stack)
		expectInt := 0
		fn := func(i int) {
			expectInt = i
		}
		p, err := gs_arg.Bind(fn, []gs.Arg{
			gs_arg.Value(3),
		})
		if err != nil {
			t.Fatal(err)
		}
		values, err := p.Call(ctx)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, expectInt, 3)
		assert.Equal(t, len(values), 0)
	})

	//t.Run("one ctx value argument", func(t *testing.T) {
	//	x := gs_dync.New()
	//	prop, err := conf.Map(map[string]interface{}{"a.b.c": 3})
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	err = x.Refresh(prop)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	stack := wiring.NewWiringStack()
	//	ctx := wiring.NewArgContext(&wiring.Wiring{P: x}, stack)
	//	expectInt := 0
	//	fn := func(i int) {
	//		expectInt = i
	//	}
	//	p, err := gs_arg.Bind(fn, []gs.Arg{
	//		gs_arg.Tag("${a.b.c}"),
	//	})
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	values, err := p.Call(ctx)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	assert.Equal(t, expectInt, 3)
	//	assert.Equal(t, len(values), 0)
	//})

	//t.Run("one ctx named bean argument", func(t *testing.T) {
	//	type st struct {
	//		i int
	//	}
	//	b := gs_core.NewBean(&st{3}).Name("a").BeanRegistration().(*gs_bean.BeanDefinition).BeanRuntime
	//	stack := wiring.NewWiringStack()
	//	ctx := wiring.NewArgContext(&wiring.Wiring{
	//		BeansByName: map[string][]wiring.BeanRuntime{
	//			"a": {b},
	//		},
	//		BeansByType: map[reflect.Type][]wiring.BeanRuntime{
	//			reflect.TypeOf(&st{}): {b},
	//		},
	//		P:     gs_dync.New(),
	//		State: wiring.Refreshed,
	//	}, stack)
	//	expectInt := 0
	//	fn := func(v *st) {
	//		expectInt = v.i
	//	}
	//	p, err := gs_arg.Bind(fn, []gs.Arg{
	//		gs_arg.Tag("a"),
	//	})
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	values, err := p.Call(ctx)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	assert.Equal(t, expectInt, 3)
	//	assert.Equal(t, len(values), 0)
	//})

	//t.Run("one ctx unnamed bean argument", func(t *testing.T) {
	//	type st struct {
	//		i int
	//	}
	//	b := gs_core.NewBean(&st{3}).Name("a").BeanRegistration().(*gs_bean.BeanDefinition).BeanRuntime
	//	stack := wiring.NewWiringStack()
	//	ctx := wiring.NewArgContext(&wiring.Wiring{
	//		BeansByName: map[string][]wiring.BeanRuntime{
	//			"a": {b},
	//		},
	//		BeansByType: map[reflect.Type][]wiring.BeanRuntime{
	//			reflect.TypeOf(&st{}): {b},
	//		},
	//		P:     gs_dync.New(),
	//		State: wiring.Refreshed,
	//	}, stack)
	//	expectInt := 0
	//	fn := func(v *st) {
	//		expectInt = v.i
	//	}
	//	p, err := gs_arg.Bind(fn, []gs.Arg{})
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	values, err := p.Call(ctx)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	assert.Equal(t, expectInt, 3)
	//	assert.Equal(t, len(values), 0)
	//})

}
