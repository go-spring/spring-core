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

package testcase_test

import (
	"os"
	"sort"
	"testing"

	"github.com/go-spring/spring-core/gs"
	"github.com/go-spring/spring-core/gs/gstest"
	"github.com/go-spring/spring-core/gs/testcase/service/my_service"
	"github.com/go-spring/spring-core/util/assert"
)

func TestMain(m *testing.M) {
	err := gstest.Init()
	if err != nil {
		panic(err)
	}
	os.Exit(gstest.Run(m))
}

func TestProp(t *testing.T) {
	assert.True(t, sort.SearchStrings(gstest.Keys(), "spring.app.name") > 0)
	assert.True(t, gstest.Has("spring.app.name"))
	subKeys, err := gstest.SubKeys("spring")
	assert.Nil(t, err)
	assert.Equal(t, subKeys, []string{"app", "force-autowire-is-nullable"})
	assert.Equal(t, gstest.Prop("spring.app.name"), "test_app")
	str, err := gstest.Resolve("my_${spring.app.name}")
	assert.Nil(t, err)
	assert.Equal(t, str, "my_test_app")
	var s string
	err = gstest.Bind(&s, "${spring.app.name}")
	assert.Nil(t, err)
	assert.Equal(t, s, "test_app")
}

func TestBean(t *testing.T) {
	{
		var s *my_service.Service
		assert.Nil(t, gstest.Get(&s))
		assert.Nil(t, s.ModelA)
		assert.Equal(t, s.ModelB.Value, "456")
		assert.Equal(t, s.AppName, "test_app")
		assert.Equal(t, s.SvrName, "svr_test")
	}
	{
		var s struct {
			Service *my_service.Service `autowire:""`
		}
		_, err := gstest.Wire(&s)
		assert.Nil(t, err)
		assert.Nil(t, s.Service.ModelA)
		assert.Equal(t, s.Service.ModelB.Value, "456")
		assert.Equal(t, s.Service.AppName, "test_app")
		assert.Equal(t, s.Service.SvrName, "svr_test")
	}
	{
		ret, err := gstest.Invoke(func(i int, s *my_service.Service) error {
			assert.Equal(t, i, 1000)
			assert.Nil(t, s.ModelA)
			assert.Equal(t, s.ModelB.Value, "456")
			assert.Equal(t, s.AppName, "test_app")
			assert.Equal(t, s.SvrName, "svr_test")
			return nil
		}, gs.ValueArg(1000))
		assert.Nil(t, err)
		assert.Equal(t, len(ret), 0)
	}
}
