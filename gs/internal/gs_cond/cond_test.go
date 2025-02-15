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
	"fmt"
	"testing"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_cond"
	"github.com/go-spring/spring-core/util/assert"
)

func TestConditionString(t *testing.T) {

	c := gs_cond.OnFunc(func(ctx gs.CondContext) (bool, error) { return false, nil })
	assert.Equal(t, fmt.Sprint(c), `OnFunc(fn=gs_cond_test.TestConditionString.func1)`)

	c = gs_cond.OnProperty("a").HavingValue("123")
	assert.Equal(t, fmt.Sprint(c), `OnProperty(name=a, havingValue=123)`)

	c = gs_cond.OnMissingProperty("a")
	assert.Equal(t, fmt.Sprint(c), `OnMissingProperty(name=a)`)

	c = gs_cond.OnBean(gs.BeanSelector{Name: "a"})
	assert.Equal(t, fmt.Sprint(c), `OnBean(selector={Name:a})`)

	c = gs_cond.OnBean(gs.BeanSelectorForType[error]())
	assert.Equal(t, fmt.Sprint(c), `OnBean(selector={Type:error})`)

	c = gs_cond.OnMissingBean(gs.BeanSelector{Name: "a"})
	assert.Equal(t, fmt.Sprint(c), `OnMissingBean(selector={Name:a})`)

	c = gs_cond.OnMissingBean(gs.BeanSelectorForType[error]())
	assert.Equal(t, fmt.Sprint(c), `OnMissingBean(selector={Type:error})`)

	c = gs_cond.OnSingleBean(gs.BeanSelector{Name: "a"})
	assert.Equal(t, fmt.Sprint(c), `OnSingleBean(selector={Name:a})`)

	c = gs_cond.OnSingleBean(gs.BeanSelectorForType[error]())
	assert.Equal(t, fmt.Sprint(c), `OnSingleBean(selector={Type:error})`)

	c = gs_cond.OnExpression("a")
	assert.Equal(t, fmt.Sprint(c), `OnExpression(expression=a)`)

	c = gs_cond.Not(gs_cond.OnBean(gs.BeanSelector{Name: "a"}))
	assert.Equal(t, fmt.Sprint(c), `Not(OnBean(selector={Name:a}))`)

	c = gs_cond.Or(gs_cond.OnBean(gs.BeanSelector{Name: "a"}))
	assert.Equal(t, fmt.Sprint(c), `OnBean(selector={Name:a})`)

	c = gs_cond.Or(gs_cond.OnBean(gs.BeanSelector{Name: "a"}), gs_cond.OnBean(gs.BeanSelector{Name: "b"}))
	assert.Equal(t, fmt.Sprint(c), `Or(OnBean(selector={Name:a}), OnBean(selector={Name:b}))`)

	c = gs_cond.And(gs_cond.OnBean(gs.BeanSelector{Name: "a"}))
	assert.Equal(t, fmt.Sprint(c), `OnBean(selector={Name:a})`)

	c = gs_cond.And(gs_cond.OnBean(gs.BeanSelector{Name: "a"}), gs_cond.OnBean(gs.BeanSelector{Name: "b"}))
	assert.Equal(t, fmt.Sprint(c), `And(OnBean(selector={Name:a}), OnBean(selector={Name:b}))`)

	c = gs_cond.None(gs_cond.OnBean(gs.BeanSelector{Name: "a"}))
	assert.Equal(t, fmt.Sprint(c), `Not(OnBean(selector={Name:a}))`)

	c = gs_cond.None(gs_cond.OnBean(gs.BeanSelector{Name: "a"}), gs_cond.OnBean(gs.BeanSelector{Name: "b"}))
	assert.Equal(t, fmt.Sprint(c), `None(OnBean(selector={Name:a}), OnBean(selector={Name:b}))`)

	c = gs_cond.And(
		gs_cond.OnBean(gs.BeanSelector{Name: "a"}),
		gs_cond.Or(
			gs_cond.OnBean(gs.BeanSelector{Name: "b"}),
			gs_cond.Not(gs_cond.OnBean(gs.BeanSelector{Name: "c"})),
		),
	)
	assert.Equal(t, fmt.Sprint(c), `And(OnBean(selector={Name:a}), Or(OnBean(selector={Name:b}), Not(OnBean(selector={Name:c}))))`)
}
