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

package gs_cond

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/util"
	"github.com/go-spring/spring-core/util/assert"
	"go.uber.org/mock/gomock"
)

var (
	trueCond  = OnFunc(func(ctx gs.CondContext) (bool, error) { return true, nil })
	falseCond = OnFunc(func(ctx gs.CondContext) (bool, error) { return false, nil })
)

func TestConditionString(t *testing.T) {

	c := OnFunc(func(ctx gs.CondContext) (bool, error) { return false, nil })
	assert.Equal(t, fmt.Sprint(c), `OnFunc(fn=gs_cond.TestConditionString.func1)`)

	c = OnProperty("a").HavingValue("123")
	assert.Equal(t, fmt.Sprint(c), `OnProperty(name=a, havingValue=123)`)

	c = OnProperty("a").HavingValue("123").MatchIfMissing()
	assert.Equal(t, fmt.Sprint(c), `OnProperty(name=a, havingValue=123, matchIfMissing)`)

	c = OnMissingProperty("a")
	assert.Equal(t, fmt.Sprint(c), `OnMissingProperty(name=a)`)

	c = OnBean[any]("a")
	assert.Equal(t, fmt.Sprint(c), `OnBean(selector={Name:a})`)

	c = OnBean[error]()
	assert.Equal(t, fmt.Sprint(c), `OnBean(selector={Type:error})`)

	c = OnMissingBean[any]("a")
	assert.Equal(t, fmt.Sprint(c), `OnMissingBean(selector={Name:a})`)

	c = OnMissingBeanSelector(gs.BeanSelectorFor[error]())
	assert.Equal(t, fmt.Sprint(c), `OnMissingBean(selector={Type:error})`)

	c = OnSingleBean[any]("a")
	assert.Equal(t, fmt.Sprint(c), `OnSingleBean(selector={Name:a})`)

	c = OnSingleBeanSelector(gs.BeanSelectorFor[error]())
	assert.Equal(t, fmt.Sprint(c), `OnSingleBean(selector={Type:error})`)

	c = OnExpression("a")
	assert.Equal(t, fmt.Sprint(c), `OnExpression(expression=a)`)

	c = Not(OnBean[any]("a"))
	assert.Equal(t, fmt.Sprint(c), `Not(OnBean(selector={Name:a}))`)

	c = Or(OnBean[any]("a"))
	assert.Equal(t, fmt.Sprint(c), `OnBean(selector={Name:a})`)

	c = Or(OnBean[any]("a"), OnBean[any]("b"))
	assert.Equal(t, fmt.Sprint(c), `Or(OnBean(selector={Name:a}), OnBean(selector={Name:b}))`)

	c = And(OnBean[any]("a"))
	assert.Equal(t, fmt.Sprint(c), `OnBean(selector={Name:a})`)

	c = And(
		OnBeanSelector(gs.BeanSelectorImpl{Name: "a"}),
		OnBeanSelector(gs.BeanSelectorImpl{Name: "b"}),
	)
	assert.Equal(t, fmt.Sprint(c), `And(OnBean(selector={Name:a}), OnBean(selector={Name:b}))`)

	c = None(OnBean[any]("a"))
	assert.Equal(t, fmt.Sprint(c), `Not(OnBean(selector={Name:a}))`)

	c = None(OnBean[any]("a"), OnBean[any]("b"))
	assert.Equal(t, fmt.Sprint(c), `None(OnBean(selector={Name:a}), OnBean(selector={Name:b}))`)

	c = And(
		OnBean[any]("a"),
		Or(
			OnBean[any]("b"),
			Not(OnBean[any]("c")),
		),
	)
	assert.Equal(t, fmt.Sprint(c), `And(OnBean(selector={Name:a}), Or(OnBean(selector={Name:b}), Not(OnBean(selector={Name:c}))))`)
}

func TestOnFunc(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		fn := func(ctx gs.CondContext) (bool, error) { return true, nil }
		cond := OnFunc(fn)
		ok, err := cond.Matches(nil)
		assert.True(t, ok)
		assert.Nil(t, err)
	})

	t.Run("error", func(t *testing.T) {
		fn := func(ctx gs.CondContext) (bool, error) { return false, errors.New("test error") }
		cond := OnFunc(fn)
		_, err := cond.Matches(nil)
		assert.Error(t, err, "test error")
	})
}

func TestOnProperty(t *testing.T) {

	t.Run("property exist", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockCondContext(ctrl)
		ctx.EXPECT().Has("test.prop").Return(true)
		cond := OnProperty("test.prop")
		ok, err := cond.Matches(ctx)
		assert.True(t, ok)
		assert.Nil(t, err)
	})

	t.Run("property exist and match", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockCondContext(ctrl)
		ctx.EXPECT().Has("test.prop").Return(true)
		ctx.EXPECT().Prop("test.prop").Return("42")
		cond := OnProperty("test.prop").HavingValue("42")
		ok, err := cond.Matches(ctx)
		assert.True(t, ok)
		assert.Nil(t, err)
	})

	t.Run("property exist but not match", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockCondContext(ctrl)
		ctx.EXPECT().Has("test.prop").Return(true)
		ctx.EXPECT().Prop("test.prop").Return("42")
		cond := OnProperty("test.prop").HavingValue("100")
		ok, _ := cond.Matches(ctx)
		assert.False(t, ok)
	})

	t.Run("property not exist but MatchIfMissing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockCondContext(ctrl)
		ctx.EXPECT().Has("missing.prop").Return(false)
		cond := OnProperty("missing.prop").MatchIfMissing()
		ok, _ := cond.Matches(ctx)
		assert.True(t, ok)
	})

	t.Run("expression", func(t *testing.T) {

		t.Run("number expression", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := NewMockCondContext(ctrl)
			ctx.EXPECT().Has("test.prop").Return(true)
			ctx.EXPECT().Prop("test.prop").Return("42")
			cond := OnProperty("test.prop").HavingValue("expr:int($) > 40")
			ok, _ := cond.Matches(ctx)
			assert.True(t, ok)
		})

		t.Run("string expression", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := NewMockCondContext(ctrl)
			ctx.EXPECT().Has("test.prop").Return(true)
			ctx.EXPECT().Prop("test.prop").Return("42")
			cond := OnProperty("test.prop").HavingValue(`expr:$ == "42"`)
			ok, _ := cond.Matches(ctx)
			assert.True(t, ok)
		})

		t.Run("invalid expression", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := NewMockCondContext(ctrl)
			ctx.EXPECT().Has("test.prop").Return(true)
			ctx.EXPECT().Prop("test.prop").Return("42")
			cond := OnProperty("test.prop").HavingValue("expr:invalid syntax")
			_, err := cond.Matches(ctx)
			assert.Error(t, err, "eval \\\"invalid syntax\\\" returns error")
		})
	})
}

func TestOnMissingProperty(t *testing.T) {

	t.Run("property exist", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockCondContext(ctrl)
		ctx.EXPECT().Has(gomock.Any()).Return(true)
		cond := OnMissingProperty("existing")
		ok, _ := cond.Matches(ctx)
		assert.False(t, ok)
	})

	t.Run("property not exist", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockCondContext(ctrl)
		ctx.EXPECT().Has(gomock.Any()).Return(false)
		cond := OnMissingProperty("missing")
		ok, _ := cond.Matches(ctx)
		assert.True(t, ok)
	})
}

func TestOnBean(t *testing.T) {

	t.Run("found bean", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return([]gs.CondBean{nil}, nil)
		cond := OnBean[any]("b")
		ok, err := cond.Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})

	t.Run("not found bean", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return(nil, nil)
		cond := OnBean[any]("b")
		ok, err := cond.Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})

	t.Run("return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return(nil, errors.New("test error"))
		cond := OnBean[any]("b")
		ok, err := cond.Matches(ctx)
		assert.Error(t, err, "test error")
		assert.False(t, ok)
	})
}

func TestOnMissingBean(t *testing.T) {

	t.Run("not found bean", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return(nil, nil)
		cond := OnMissingBean[any]("bean1")
		ok, err := cond.Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})

	t.Run("found bean", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return([]gs.CondBean{nil}, nil)
		cond := OnMissingBean[any]("bean1")
		ok, err := cond.Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})

	t.Run("return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return(nil, errors.New("test error"))
		cond := OnMissingBean[any]("b")
		ok, err := cond.Matches(ctx)
		assert.Error(t, err, "test error")
		assert.False(t, ok)
	})
}

func TestOnSingleBean(t *testing.T) {

	t.Run("found only one bean", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return([]gs.CondBean{nil}, nil)
		cond := OnSingleBean[any]("b")
		ok, _ := cond.Matches(ctx)
		assert.True(t, ok)
	})

	t.Run("found two beans", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return([]gs.CondBean{nil, nil}, nil)
		cond := OnSingleBean[any]("b")
		ok, _ := cond.Matches(ctx)
		assert.False(t, ok)
	})

	t.Run("return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return(nil, errors.New("test error"))
		cond := OnSingleBean[any]("b")
		ok, err := cond.Matches(ctx)
		assert.Error(t, err, "test error")
		assert.False(t, ok)
	})
}

func TestOnExpression(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := NewMockCondContext(ctrl)
	cond := OnExpression("1+1==2")
	_, err := cond.Matches(ctx)
	assert.True(t, errors.Is(err, util.UnimplementedMethod))
}

func TestNot(t *testing.T) {

	t.Run("true", func(t *testing.T) {
		cond := Not(trueCond)
		ok, err := cond.Matches(nil)
		assert.Nil(t, err)
		assert.False(t, ok)
	})

	t.Run("false", func(t *testing.T) {
		cond := Not(falseCond)
		ok, err := cond.Matches(nil)
		assert.Nil(t, err)
		assert.True(t, ok)
	})

	t.Run("return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return(nil, errors.New("test error"))
		cond := OnSingleBean[any]("b")
		ok, err := Not(cond).Matches(ctx)
		assert.Error(t, err, "test error")
		assert.False(t, ok)
	})
}

func TestAnd(t *testing.T) {

	t.Run("nil", func(t *testing.T) {
		cond := And()
		assert.Nil(t, cond)
	})

	t.Run("one condition", func(t *testing.T) {
		cond := And(trueCond)
		assert.Equal(t, trueCond, cond)
	})

	t.Run("two conditions | true", func(t *testing.T) {
		cond := And(trueCond, trueCond)
		ok, err := cond.Matches(nil)
		assert.Nil(t, err)
		assert.True(t, ok)
	})

	t.Run("two conditions | false", func(t *testing.T) {
		cond := And(trueCond, falseCond)
		ok, err := cond.Matches(nil)
		assert.Nil(t, err)
		assert.False(t, ok)
	})

	t.Run("return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return(nil, errors.New("test error"))
		cond := OnSingleBean[any]("b")
		ok, err := And(cond, trueCond).Matches(ctx)
		assert.Error(t, err, "test error")
		assert.False(t, ok)
	})
}

func TestOr(t *testing.T) {

	t.Run("nil", func(t *testing.T) {
		cond := Or()
		assert.Nil(t, cond)
	})

	t.Run("one condition", func(t *testing.T) {
		cond := Or(trueCond)
		assert.Equal(t, trueCond, cond)
	})

	t.Run("two conditions | true", func(t *testing.T) {
		cond := Or(trueCond, falseCond)
		ok, err := cond.Matches(nil)
		assert.Nil(t, err)
		assert.True(t, ok)
	})

	t.Run("two conditions | false", func(t *testing.T) {
		cond := Or(falseCond, falseCond)
		ok, err := cond.Matches(nil)
		assert.Nil(t, err)
		assert.False(t, ok)
	})

	t.Run("return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return(nil, errors.New("test error"))
		cond := OnSingleBean[any]("b")
		ok, err := Or(cond, trueCond).Matches(ctx)
		assert.Error(t, err, "test error")
		assert.False(t, ok)
	})
}

func TestNone(t *testing.T) {

	t.Run("nil", func(t *testing.T) {
		cond := None()
		assert.Nil(t, cond)
	})

	t.Run("one condition", func(t *testing.T) {
		cond := None(trueCond)
		ok, err := cond.Matches(nil)
		assert.Nil(t, err)
		assert.False(t, ok)
	})

	t.Run("two conditions | false", func(t *testing.T) {
		cond := None(trueCond, falseCond)
		ok, err := cond.Matches(nil)
		assert.Nil(t, err)
		assert.False(t, ok)
	})

	t.Run("two conditions | true", func(t *testing.T) {
		cond := None(falseCond, falseCond)
		ok, err := cond.Matches(nil)
		assert.Nil(t, err)
		assert.True(t, ok)
	})

	t.Run("return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := NewMockCondContext(ctrl)
		ctx.EXPECT().Find(gomock.Any()).Return(nil, errors.New("test error"))
		cond := OnSingleBean[any]("b")
		ok, err := None(cond, trueCond).Matches(ctx)
		assert.Error(t, err, "test error")
		assert.False(t, ok)
	})
}
