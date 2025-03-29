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

package gs_cond_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/util"
	"github.com/go-spring/spring-core/util/assert"
	"go.uber.org/mock/gomock"
)

var (
	trueCond  = gs_cond.OnFunc(func(ctx gs.CondContext) (bool, error) { return true, nil })
	falseCond = gs_cond.OnFunc(func(ctx gs.CondContext) (bool, error) { return false, nil })
)

func TestConditionString(t *testing.T) {

	c := gs_cond.OnFunc(func(ctx gs.CondContext) (bool, error) { return false, nil })
	assert.Equal(t, fmt.Sprint(c), `OnFunc(fn=gs_cond_test.TestConditionString.func1)`)

	c = gs_cond.OnProperty("a").HavingValue("123")
	assert.Equal(t, fmt.Sprint(c), `OnProperty(name=a, havingValue=123)`)

	c = gs_cond.OnProperty("a").HavingValue("123").MatchIfMissing()
	assert.Equal(t, fmt.Sprint(c), `OnProperty(name=a, havingValue=123, matchIfMissing)`)

	c = gs_cond.OnMissingProperty("a")
	assert.Equal(t, fmt.Sprint(c), `OnMissingProperty(name=a)`)

	c = gs_cond.OnBean[any]("a")
	assert.Equal(t, fmt.Sprint(c), `OnBean(selector={Name:a})`)

	c = gs_cond.OnBean[error]()
	assert.Equal(t, fmt.Sprint(c), `OnBean(selector={Type:error})`)

	c = gs_cond.OnMissingBean[any]("a")
	assert.Equal(t, fmt.Sprint(c), `OnMissingBean(selector={Name:a})`)

	c = gs_cond.OnMissingBeanSelector(gs.BeanSelectorFor[error]())
	assert.Equal(t, fmt.Sprint(c), `OnMissingBean(selector={Type:error})`)

	c = gs_cond.OnSingleBean[any]("a")
	assert.Equal(t, fmt.Sprint(c), `OnSingleBean(selector={Name:a})`)

	c = gs_cond.OnSingleBeanSelector(gs.BeanSelectorFor[error]())
	assert.Equal(t, fmt.Sprint(c), `OnSingleBean(selector={Type:error})`)

	c = gs_cond.OnExpression("a")
	assert.Equal(t, fmt.Sprint(c), `OnExpression(expression=a)`)

	c = gs_cond.Not(gs_cond.OnBean[any]("a"))
	assert.Equal(t, fmt.Sprint(c), `Not(OnBean(selector={Name:a}))`)

	c = gs_cond.Or(gs_cond.OnBean[any]("a"))
	assert.Equal(t, fmt.Sprint(c), `OnBean(selector={Name:a})`)

	c = gs_cond.Or(gs_cond.OnBean[any]("a"), gs_cond.OnBean[any]("b"))
	assert.Equal(t, fmt.Sprint(c), `Or(OnBean(selector={Name:a}), OnBean(selector={Name:b}))`)

	c = gs_cond.And(gs_cond.OnBean[any]("a"))
	assert.Equal(t, fmt.Sprint(c), `OnBean(selector={Name:a})`)

	c = gs_cond.And(
		gs_cond.OnBeanSelector(gs.BeanSelectorImpl{Name: "a"}),
		gs_cond.OnBeanSelector(gs.BeanSelectorImpl{Name: "b"}),
	)
	assert.Equal(t, fmt.Sprint(c), `And(OnBean(selector={Name:a}), OnBean(selector={Name:b}))`)

	c = gs_cond.None(gs_cond.OnBean[any]("a"))
	assert.Equal(t, fmt.Sprint(c), `Not(OnBean(selector={Name:a}))`)

	c = gs_cond.None(gs_cond.OnBean[any]("a"), gs_cond.OnBean[any]("b"))
	assert.Equal(t, fmt.Sprint(c), `None(OnBean(selector={Name:a}), OnBean(selector={Name:b}))`)

	c = gs_cond.And(
		gs_cond.OnBean[any]("a"),
		gs_cond.Or(
			gs_cond.OnBean[any]("b"),
			gs_cond.Not(gs_cond.OnBean[any]("c")),
		),
	)
	assert.Equal(t, fmt.Sprint(c), `And(OnBean(selector={Name:a}), Or(OnBean(selector={Name:b}), Not(OnBean(selector={Name:c}))))`)
}

func TestOnFunc(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		fn := func(ctx gs.CondContext) (bool, error) { return true, nil }
		cond := gs_cond.OnFunc(fn)
		ok, err := cond.Matches(ctx)
		assert.True(t, ok)
		assert.Nil(t, err)
	})

	t.Run("error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		fn := func(ctx gs.CondContext) (bool, error) { return false, errors.New("test error") }
		cond := gs_cond.OnFunc(fn)
		_, err := cond.Matches(ctx)
		assert.Error(t, err, "test error")
	})
}

func TestOnProperty(t *testing.T) {

	t.Run("property exist and match", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		ctx.EXPECT().Has("test.prop").Return(true)
		ctx.EXPECT().Prop("test.prop").Return("42")
		cond := gs_cond.OnProperty("test.prop").HavingValue("42")
		ok, err := cond.Matches(ctx)
		assert.True(t, ok)
		assert.Nil(t, err)
	})

	t.Run("property exist but not match", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		ctx.EXPECT().Has("test.prop").Return(true)
		ctx.EXPECT().Prop("test.prop").Return("42")
		cond := gs_cond.OnProperty("test.prop").HavingValue("100")
		ok, _ := cond.Matches(ctx)
		assert.False(t, ok)
	})

	t.Run("property not exist but MatchIfMissing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		ctx.EXPECT().Has("missing.prop").Return(false)
		cond := gs_cond.OnProperty("missing.prop").MatchIfMissing()
		ok, _ := cond.Matches(ctx)
		assert.True(t, ok)
	})

	t.Run("expression", func(t *testing.T) {

		t.Run("number expression", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := gs.NewMockCondContext(ctrl)
			ctx.EXPECT().Has("test.prop").Return(true)
			ctx.EXPECT().Prop("test.prop").Return("42")
			cond := gs_cond.OnProperty("test.prop").HavingValue("expr:int($) > 40")
			ok, _ := cond.Matches(ctx)
			assert.True(t, ok)
		})

		t.Run("string expression", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := gs.NewMockCondContext(ctrl)
			ctx.EXPECT().Has("test.prop").Return(true)
			ctx.EXPECT().Prop("test.prop").Return("42")
			cond := gs_cond.OnProperty("test.prop").HavingValue("expr:$ == \"42\"")
			ok, _ := cond.Matches(ctx)
			assert.True(t, ok)
		})

		t.Run("invalid expression", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := gs.NewMockCondContext(ctrl)
			ctx.EXPECT().Has("test.prop").Return(true)
			ctx.EXPECT().Prop("test.prop").Return("42")
			cond := gs_cond.OnProperty("test.prop").HavingValue("expr:invalid syntax")
			_, err := cond.Matches(ctx)
			assert.Error(t, err, "eval \\\"invalid syntax\\\" returns error")
		})
	})
}

func TestOnMissingProperty(t *testing.T) {

	t.Run("property exist", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		ctx.EXPECT().Has(gomock.Any()).Return(true)
		cond := gs_cond.OnMissingProperty("existing")
		ok, _ := cond.Matches(ctx)
		assert.False(t, ok)
	})

	t.Run("property not exist", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		ctx.EXPECT().Has(gomock.Any()).Return(false)
		cond := gs_cond.OnMissingProperty("missing")
		ok, _ := cond.Matches(ctx)
		assert.True(t, ok)
	})
}

func TestOnBean(t *testing.T) {

	t.Run("found bean", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return([]gs.CondBean{nil}, nil)
		cond := gs_cond.OnBean[any]("b")
		ok, err := cond.Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})

	t.Run("not found bean", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return(nil, nil)
		cond := gs_cond.OnBean[any]("b")
		ok, err := cond.Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})

	t.Run("return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return(nil, errors.New("test error"))
		cond := gs_cond.OnBean[any]("b")
		ok, err := cond.Matches(ctx)
		assert.Error(t, err, "test error")
		assert.False(t, ok)
	})
}

func TestOnMissingBean(t *testing.T) {

	t.Run("not found bean", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return(nil, nil)
		cond := gs_cond.OnMissingBean[any]("bean1")
		ok, err := cond.Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})

	t.Run("found bean", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return([]gs.CondBean{nil}, nil)
		cond := gs_cond.OnMissingBean[any]("bean1")
		ok, err := cond.Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})

	t.Run("return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return(nil, errors.New("test error"))
		cond := gs_cond.OnMissingBean[any]("b")
		ok, err := cond.Matches(ctx)
		assert.Error(t, err, "test error")
		assert.False(t, ok)
	})
}

func TestOnSingleBean(t *testing.T) {

	t.Run("found only one bean", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return([]gs.CondBean{nil}, nil)
		cond := gs_cond.OnSingleBean[any]("b")
		ok, _ := cond.Matches(ctx)
		assert.True(t, ok)
	})

	t.Run("found two beans", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return([]gs.CondBean{nil, nil}, nil)
		cond := gs_cond.OnSingleBean[any]("b")
		ok, _ := cond.Matches(ctx)
		assert.False(t, ok)
	})

	t.Run("return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return(nil, errors.New("test error"))
		cond := gs_cond.OnSingleBean[any]("b")
		ok, err := cond.Matches(ctx)
		assert.Error(t, err, "test error")
		assert.False(t, ok)
	})
}

func TestOnExpression(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := gs.NewMockCondContext(ctrl)
	cond := gs_cond.OnExpression("1+1==2")
	_, err := cond.Matches(ctx)
	assert.True(t, errors.Is(err, util.UnimplementedMethod))
}

func TestNot(t *testing.T) {

	t.Run("true", func(t *testing.T) {
		cond := gs_cond.Not(trueCond)
		ok, err := cond.Matches(nil)
		assert.Nil(t, err)
		assert.False(t, ok)
	})

	t.Run("false", func(t *testing.T) {
		cond := gs_cond.Not(falseCond)
		ok, err := cond.Matches(nil)
		assert.Nil(t, err)
		assert.True(t, ok)
	})

	t.Run("return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return(nil, errors.New("test error"))
		cond := gs_cond.OnSingleBean[any]("b")
		ok, err := gs_cond.Not(cond).Matches(ctx)
		assert.Error(t, err, "test error")
		assert.False(t, ok)
	})
}

func TestAnd(t *testing.T) {

	t.Run("nil", func(t *testing.T) {
		cond := gs_cond.And()
		assert.Nil(t, cond)
	})

	t.Run("one condition", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		cond := gs_cond.And(trueCond)
		ok, err := cond.Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})

	t.Run("two conditions | true", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		cond := gs_cond.And(trueCond, trueCond)
		ok, err := cond.Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})

	t.Run("two conditions | false", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		cond := gs_cond.And(trueCond, falseCond)
		ok, err := cond.Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})

	t.Run("return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return(nil, errors.New("test error"))
		cond := gs_cond.OnSingleBean[any]("b")
		ok, err := gs_cond.And(cond, trueCond).Matches(ctx)
		assert.Error(t, err, "test error")
		assert.False(t, ok)
	})
}

func TestOr(t *testing.T) {

	t.Run("nil", func(t *testing.T) {
		cond := gs_cond.Or()
		assert.Nil(t, cond)
	})

	t.Run("one condition", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		cond := gs_cond.Or(trueCond)
		ok, err := cond.Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})

	t.Run("two conditions | true", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		cond := gs_cond.Or(trueCond, falseCond)
		ok, err := cond.Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})

	t.Run("two conditions | false", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		cond := gs_cond.Or(falseCond, falseCond)
		ok, err := cond.Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})

	t.Run("return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return(nil, errors.New("test error"))
		cond := gs_cond.OnSingleBean[any]("b")
		ok, err := gs_cond.Or(cond, trueCond).Matches(ctx)
		assert.Error(t, err, "test error")
		assert.False(t, ok)
	})
}

func TestNone(t *testing.T) {

	t.Run("nil", func(t *testing.T) {
		cond := gs_cond.None()
		assert.Nil(t, cond)
	})

	t.Run("one condition", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		cond := gs_cond.None(trueCond)
		ok, err := cond.Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})

	t.Run("two conditions | true", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		cond := gs_cond.None(trueCond, falseCond)
		ok, err := cond.Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})

	t.Run("two conditions | false", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		cond := gs_cond.None(falseCond, falseCond)
		ok, err := cond.Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})

	t.Run("return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs.NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return(nil, errors.New("test error"))
		cond := gs_cond.OnSingleBean[any]("b")
		ok, err := gs_cond.None(cond, trueCond).Matches(ctx)
		assert.Error(t, err, "test error")
		assert.False(t, ok)
	})
}
