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

//
//func TestNot(t *testing.T) {
//	c := container(t, nil)
//	ok, err := gs_cond.Not(gs_cond.OnMissingProperty("a")).Matches(c.(gs.CondContext))
//	assert.Nil(t, err)
//	assert.False(t, ok)
//	ok, err = gs_cond.Not(gs_cond.Not(gs_cond.OnMissingProperty("a"))).Matches(c.(gs.CondContext))
//	assert.Nil(t, err)
//	assert.True(t, ok)
//}
//
//func TestOnProperty(t *testing.T) {
//	t.Run("no property & no HavingValue & no MatchIfMissing", func(t *testing.T) {
//		c := container(t, nil)
//		ok, err := gs_cond.OnProperty("a").Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.False(t, ok)
//	})
//	t.Run("has property & no HavingValue & no MatchIfMissing", func(t *testing.T) {
//		c := container(t, func(p *conf.Properties, c *gs_core.Container) error {
//			return p.Set("a", "b")
//		})
//		ok, err := gs_cond.OnProperty("a").Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.True(t, ok)
//	})
//	t.Run("no property & has HavingValue & no MatchIfMissing", func(t *testing.T) {
//		c := container(t, nil)
//		ok, err := gs_cond.OnProperty("a", gs_cond.HavingValue("a")).Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.False(t, ok)
//	})
//	t.Run("diff property & has HavingValue & no MatchIfMissing", func(t *testing.T) {
//		c := container(t, func(p *conf.Properties, c *gs_core.Container) error {
//			return p.Set("a", "b")
//		})
//		ok, err := gs_cond.OnProperty("a", gs_cond.HavingValue("a")).Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.False(t, ok)
//	})
//	t.Run("same property & has HavingValue & no MatchIfMissing", func(t *testing.T) {
//		c := container(t, func(p *conf.Properties, c *gs_core.Container) error {
//			return p.Set("a", "a")
//		})
//		ok, err := gs_cond.OnProperty("a", gs_cond.HavingValue("a")).Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.True(t, ok)
//	})
//	t.Run("no property & no HavingValue & has MatchIfMissing", func(t *testing.T) {
//		c := container(t, nil)
//		ok, err := gs_cond.OnProperty("a", gs_cond.MatchIfMissing()).Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.True(t, ok)
//	})
//	t.Run("has property & no HavingValue & has MatchIfMissing", func(t *testing.T) {
//		c := container(t, func(p *conf.Properties, c *gs_core.Container) error {
//			return p.Set("a", "b")
//		})
//		ok, err := gs_cond.OnProperty("a", gs_cond.MatchIfMissing()).Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.True(t, ok)
//	})
//	t.Run("no property & has HavingValue & has MatchIfMissing", func(t *testing.T) {
//		c := container(t, nil)
//		ok, err := gs_cond.OnProperty("a", gs_cond.HavingValue("a"), gs_cond.MatchIfMissing()).Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.True(t, ok)
//	})
//	t.Run("diff property & has HavingValue & has MatchIfMissing", func(t *testing.T) {
//		c := container(t, func(p *conf.Properties, c *gs_core.Container) error {
//			return p.Set("a", "b")
//		})
//		ok, err := gs_cond.OnProperty("a", gs_cond.HavingValue("a"), gs_cond.MatchIfMissing()).Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.False(t, ok)
//	})
//	t.Run("same property & has HavingValue & has MatchIfMissing", func(t *testing.T) {
//		c := container(t, func(p *conf.Properties, c *gs_core.Container) error {
//			return p.Set("a", "a")
//		})
//		ok, err := gs_cond.OnProperty("a", gs_cond.HavingValue("a"), gs_cond.MatchIfMissing()).Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.True(t, ok)
//	})
//	t.Run("go expression", func(t *testing.T) {
//		testcases := []struct {
//			propValue    string
//			expression   string
//			expectResult bool
//		}{
//			{
//				"a",
//				"expr:$==\"a\"",
//				true,
//			},
//			{
//				"a",
//				"expr:$==\"b\"",
//				false,
//			},
//			{
//				"3",
//				"expr:$==3",
//				true,
//			},
//			{
//				"3",
//				"expr:$==4",
//				false,
//			},
//			{
//				"3",
//				"expr:$>1&&$<5",
//				true,
//			},
//			{
//				"false",
//				"expr:$",
//				false,
//			},
//			{
//				"false",
//				"expr:!$",
//				true,
//			},
//		}
//		for _, testcase := range testcases {
//			c := container(t, func(p *conf.Properties, c *gs_core.Container) error {
//				return p.Set("a", testcase.propValue)
//			})
//			ok, err := gs_cond.OnProperty("a", gs_cond.HavingValue(testcase.expression)).Matches(c.(gs.CondContext))
//			assert.Nil(t, err)
//			assert.Equal(t, ok, testcase.expectResult)
//		}
//	})
//}
//
//func TestOnMissingProperty(t *testing.T) {
//	t.Run("no property", func(t *testing.T) {
//		c := container(t, nil)
//		ok, err := gs_cond.OnMissingProperty("a").Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.True(t, ok)
//	})
//	t.Run("has property", func(t *testing.T) {
//		c := container(t, func(p *conf.Properties, c *gs_core.Container) error {
//			return p.Set("a", "b")
//		})
//		ok, err := gs_cond.OnMissingProperty("a").Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.False(t, ok)
//	})
//}
//
//func TestOnBean(t *testing.T) {
//	t.Run("return error", func(t *testing.T) {
//		c := container(t, nil)
//		ok, err := gs_cond.OnBean("${a}").Matches(c.(gs.CondContext))
//		assert.Error(t, err, "property \"a\" not exist")
//		assert.False(t, ok)
//	})
//	t.Run("no bean", func(t *testing.T) {
//		c := container(t, nil)
//		ok, err := gs_cond.OnBean("a").Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.False(t, ok)
//	})
//	t.Run("one bean", func(t *testing.T) {
//		c := container(t, func(p *conf.Properties, c *gs_core.Container) error {
//			c.Provide(conf.New).Name("a")
//			return nil
//		})
//		ok, err := gs_cond.OnBean("a").Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.True(t, ok)
//	})
//	t.Run("more than one beans", func(t *testing.T) {
//		c := container(t, func(p *conf.Properties, c *gs_core.Container) error {
//			c.Provide(conf.New).Name("a")
//			c.Provide(NewVarInterfaceObj).Name("a")
//			return nil
//		})
//		ok, err := gs_cond.OnBean("a").Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.True(t, ok)
//	})
//}
//
//func TestOnMissingBean(t *testing.T) {
//	t.Run("return error", func(t *testing.T) {
//		c := container(t, nil)
//		ok, err := gs_cond.OnMissingBean("${x}").Matches(c.(gs.CondContext))
//		assert.Error(t, err, "property \"x\" not exist")
//		assert.False(t, ok)
//	})
//	t.Run("no bean", func(t *testing.T) {
//		c := container(t, nil)
//		ok, err := gs_cond.OnMissingBean("a").Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.True(t, ok)
//	})
//	t.Run("one bean", func(t *testing.T) {
//		c := container(t, func(p *conf.Properties, c *gs_core.Container) error {
//			c.Provide(conf.New).Name("a")
//			return nil
//		})
//		ok, err := gs_cond.OnMissingBean("a").Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.False(t, ok)
//	})
//	t.Run("more than one beans", func(t *testing.T) {
//		c := container(t, func(p *conf.Properties, c *gs_core.Container) error {
//			c.Provide(conf.New).Name("a")
//			c.Provide(NewVarInterfaceObj).Name("a")
//			return nil
//		})
//		ok, err := gs_cond.OnMissingBean("a").Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.False(t, ok)
//	})
//}
//
//func TestOnSingleBean(t *testing.T) {
//	t.Run("return error", func(t *testing.T) {
//		c := container(t, nil)
//		ok, err := gs_cond.OnSingleBean("${x}").Matches(c.(gs.CondContext))
//		assert.Error(t, err, "property \"x\" not exist")
//		assert.False(t, ok)
//	})
//	t.Run("no bean", func(t *testing.T) {
//		c := container(t, nil)
//		ok, err := gs_cond.OnSingleBean("a").Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.False(t, ok)
//	})
//	t.Run("one bean", func(t *testing.T) {
//		c := container(t, func(p *conf.Properties, c *gs_core.Container) error {
//			c.Provide(conf.New).Name("a")
//			return nil
//		})
//		ok, err := gs_cond.OnSingleBean("a").Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.True(t, ok)
//	})
//	t.Run("more than one beans", func(t *testing.T) {
//		c := container(t, func(p *conf.Properties, c *gs_core.Container) error {
//			c.Provide(conf.New).Name("a")
//			c.Provide(NewVarInterfaceObj).Name("a")
//			return nil
//		})
//		ok, err := gs_cond.OnSingleBean("a").Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.False(t, ok)
//	})
//}
//
//func TestOnExpression(t *testing.T) {
//	c := container(t, nil)
//	ok, err := gs_cond.OnExpression("").Matches(c.(gs.CondContext))
//	assert.Error(t, err, "unimplemented method")
//	assert.False(t, ok)
//}
//
//func TestOnMatches(t *testing.T) {
//	c := container(t, nil)
//	ok, err := gs_cond.OnFunc(func(ctx gs.CondContext) (bool, error) {
//		return false, nil
//	}).Matches(c.(gs.CondContext))
//	assert.Nil(t, err)
//	assert.False(t, ok)
//}
//
//func TestGroup(t *testing.T) {
//	t.Run("ok && ", func(t *testing.T) {
//		c := container(t, nil)
//		ok, err := gs_cond.And(gs_cond.OnMissingProperty("a")).Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.True(t, ok)
//	})
//	t.Run("ok && !ok", func(t *testing.T) {
//		c := container(t, nil)
//		ok, err := gs_cond.And(gs_cond.OnMissingProperty("a"), gs_cond.Not(gs_cond.OnMissingProperty("a"))).Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.False(t, ok)
//	})
//	t.Run("ok || ", func(t *testing.T) {
//		c := container(t, nil)
//		ok, err := gs_cond.Or(gs_cond.OnMissingProperty("a")).Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.True(t, ok)
//	})
//	t.Run("ok || !ok", func(t *testing.T) {
//		c := container(t, nil)
//		ok, err := gs_cond.Or(gs_cond.OnMissingProperty("a"), gs_cond.Not(gs_cond.OnMissingProperty("a"))).Matches(c.(gs.CondContext))
//		assert.Nil(t, err)
//		assert.True(t, ok)
//	})
//}
