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

package gs_cond_test

import (
	"errors"
	"testing"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/util/assert"
	"go.uber.org/mock/gomock"
)

func TestOK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := gs_cond.NewMockContext(ctrl)
	ok, err := gs_cond.OK().Matches(ctx)
	assert.Nil(t, err)
	assert.True(t, ok)
}

func TestNot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := gs_cond.NewMockContext(ctrl)
	ok, err := gs_cond.Not(gs_cond.OK()).Matches(ctx)
	assert.Nil(t, err)
	assert.False(t, ok)
}

func TestOnProperty(t *testing.T) {
	t.Run("no property & no HavingValue & no MatchIfMissing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Has("a").Return(false)
		ok, err := gs_cond.OnProperty("a").Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
	t.Run("has property & no HavingValue & no MatchIfMissing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Has("a").Return(true)
		ok, err := gs_cond.OnProperty("a").Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
	t.Run("no property & has HavingValue & no MatchIfMissing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Has("a").Return(false)
		ok, err := gs_cond.OnProperty("a", gs_cond.HavingValue("a")).Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
	t.Run("diff property & has HavingValue & no MatchIfMissing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Has("a").Return(true)
		ctx.EXPECT().Prop("a").Return("b")
		ok, err := gs_cond.OnProperty("a", gs_cond.HavingValue("a")).Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
	t.Run("same property & has HavingValue & no MatchIfMissing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Has("a").Return(true)
		ctx.EXPECT().Prop("a").Return("a")
		ok, err := gs_cond.OnProperty("a", gs_cond.HavingValue("a")).Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
	t.Run("no property & no HavingValue & has MatchIfMissing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Has("a").Return(false)
		ok, err := gs_cond.OnProperty("a", gs_cond.MatchIfMissing()).Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
	t.Run("has property & no HavingValue & has MatchIfMissing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Has("a").Return(true)
		ok, err := gs_cond.OnProperty("a", gs_cond.MatchIfMissing()).Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
	t.Run("no property & has HavingValue & has MatchIfMissing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Has("a").Return(false)
		ok, err := gs_cond.OnProperty("a", gs_cond.HavingValue("a"), gs_cond.MatchIfMissing()).Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
	t.Run("diff property & has HavingValue & has MatchIfMissing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Has("a").Return(true)
		ctx.EXPECT().Prop("a").Return("b")
		ok, err := gs_cond.OnProperty("a", gs_cond.HavingValue("a"), gs_cond.MatchIfMissing()).Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
	t.Run("same property & has HavingValue & has MatchIfMissing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Has("a").Return(true)
		ctx.EXPECT().Prop("a").Return("a")
		ok, err := gs_cond.OnProperty("a", gs_cond.HavingValue("a"), gs_cond.MatchIfMissing()).Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
	t.Run("go expression", func(t *testing.T) {
		testcases := []struct {
			propValue    string
			expression   string
			expectResult bool
		}{
			{
				"a",
				"expr:$==\"a\"",
				true,
			},
			{
				"a",
				"expr:$==\"b\"",
				false,
			},
			{
				"3",
				"expr:$==3",
				true,
			},
			{
				"3",
				"expr:$==4",
				false,
			},
			{
				"3",
				"expr:$>1&&$<5",
				true,
			},
			{
				"false",
				"expr:$",
				false,
			},
			{
				"false",
				"expr:!$",
				true,
			},
		}
		for _, testcase := range testcases {
			ctrl := gomock.NewController(t)
			ctx := gs_cond.NewMockContext(ctrl)
			ctx.EXPECT().Has("a").Return(true)
			ctx.EXPECT().Prop("a").Return(testcase.propValue)
			ok, err := gs_cond.OnProperty("a", gs_cond.HavingValue(testcase.expression)).Matches(ctx)
			assert.Nil(t, err)
			assert.Equal(t, ok, testcase.expectResult)
			ctrl.Finish()
		}
	})
}

func TestOnMissingProperty(t *testing.T) {
	t.Run("no property", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Has("a").Return(false)
		ok, err := gs_cond.OnMissingProperty("a").Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
	t.Run("has property", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Has("a").Return(true)
		ok, err := gs_cond.OnMissingProperty("a").Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
}

func TestOnBean(t *testing.T) {
	t.Run("return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Find("a").Return(nil, errors.New("error"))
		ok, err := gs_cond.OnBean("a").Matches(ctx)
		assert.Error(t, err, "error")
		assert.False(t, ok)
	})
	t.Run("no bean", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Find("a").Return(nil, nil)
		ok, err := gs_cond.OnBean("a").Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
	t.Run("one bean", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Find("a").Return([]*gs.BeanDefinition{
			nil,
		}, nil)
		ok, err := gs_cond.OnBean("a").Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
	t.Run("more than one beans", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Find("a").Return([]*gs.BeanDefinition{
			nil, nil,
		}, nil)
		ok, err := gs_cond.OnBean("a").Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
}

func TestOnMissingBean(t *testing.T) {
	t.Run("return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Find("a").Return(nil, errors.New("error"))
		ok, err := gs_cond.OnMissingBean("a").Matches(ctx)
		assert.Error(t, err, "error")
		assert.False(t, ok)
	})
	t.Run("no bean", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Find("a").Return(nil, nil)
		ok, err := gs_cond.OnMissingBean("a").Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
	t.Run("one bean", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Find("a").Return([]*gs.BeanDefinition{
			nil,
		}, nil)
		ok, err := gs_cond.OnMissingBean("a").Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
	t.Run("more than one beans", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Find("a").Return([]*gs.BeanDefinition{
			nil, nil,
		}, nil)
		ok, err := gs_cond.OnMissingBean("a").Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
}

func TestOnSingleBean(t *testing.T) {
	t.Run("return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Find("a").Return(nil, errors.New("error"))
		ok, err := gs_cond.OnSingleBean("a").Matches(ctx)
		assert.Error(t, err, "error")
		assert.False(t, ok)
	})
	t.Run("no bean", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Find("a").Return(nil, nil)
		ok, err := gs_cond.OnSingleBean("a").Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
	t.Run("one bean", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Find("a").Return([]*gs.BeanDefinition{
			nil,
		}, nil)
		ok, err := gs_cond.OnSingleBean("a").Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
	t.Run("more than one beans", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Find("a").Return([]*gs.BeanDefinition{
			nil, nil,
		}, nil)
		ok, err := gs_cond.OnSingleBean("a").Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
}

func TestOnExpression(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := gs_cond.NewMockContext(ctrl)
	ok, err := gs_cond.OnExpression("").Matches(ctx)
	assert.Error(t, err, "unimplemented method")
	assert.False(t, ok)
}

func TestOnMatches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := gs_cond.NewMockContext(ctrl)
	ok, err := gs_cond.OnMatches(func(ctx gs.CondContext) (bool, error) {
		return false, nil
	}).Matches(ctx)
	assert.Nil(t, err)
	assert.False(t, ok)
}

func TestOnProfile(t *testing.T) {
	t.Run("no property", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Has("spring.profiles.active").Return(false)
		ok, err := gs_cond.OnProfile("test").Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
	t.Run("diff property", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Has("spring.profiles.active").Return(true)
		ctx.EXPECT().Prop("spring.profiles.active").Return("dev")
		ok, err := gs_cond.OnProfile("test").Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
	t.Run("same property", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ctx.EXPECT().Has("spring.profiles.active").Return(true)
		ctx.EXPECT().Prop("spring.profiles.active").Return("test")
		ok, err := gs_cond.OnProfile("test").Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
}

func TestConditional(t *testing.T) {
	t.Run("ok && ", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ok, err := gs_cond.On(gs_cond.OK()).And().Matches(ctx)
		assert.Error(t, err, "no condition in last node")
		assert.False(t, ok)
	})
	t.Run("ok && !ok", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ok, err := gs_cond.On(gs_cond.OK()).And().On(gs_cond.Not(gs_cond.OK())).Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
	t.Run("ok || ", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ok, err := gs_cond.On(gs_cond.OK()).Or().Matches(ctx)
		assert.Error(t, err, "no condition in last node")
		assert.False(t, ok)
	})
	t.Run("ok || !ok", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ok, err := gs_cond.On(gs_cond.OK()).Or().On(gs_cond.Not(gs_cond.OK())).Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
}

func TestGroup(t *testing.T) {
	t.Run("ok && ", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ok, err := gs_cond.Group(gs_cond.And, gs_cond.OK()).Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
	t.Run("ok && !ok", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ok, err := gs_cond.Group(gs_cond.And, gs_cond.OK(), gs_cond.Not(gs_cond.OK())).Matches(ctx)
		assert.Nil(t, err)
		assert.False(t, ok)
	})
	t.Run("ok || ", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ok, err := gs_cond.Group(gs_cond.Or, gs_cond.OK()).Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
	t.Run("ok || !ok", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ctx := gs_cond.NewMockContext(ctrl)
		ok, err := gs_cond.Group(gs_cond.Or, gs_cond.OK(), gs_cond.Not(gs_cond.OK())).Matches(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
}
