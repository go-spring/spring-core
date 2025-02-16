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

package my_service_test

import (
	"context"
	"os"
	"sort"
	"testing"

	"github.com/go-spring/spring-core/gs"
	"github.com/go-spring/spring-core/gs/gstest"
	"github.com/go-spring/spring-core/gs/gstest/testcase/internal/service/my_service"
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
		assert.Nil(t, s.DoB(t.Context()))
	}
	{
		var s struct {
			Service *my_service.Service `autowire:""`
		}
		_, err := gstest.Wire(&s)
		assert.Nil(t, err)
		assert.Nil(t, s.Service.DoB(t.Context()))
		assert.Panic(t, func() { _ = s.Service.DoA(t.Context()) }, "ModelA is nil")
	}
	{
		_, err := gstest.Invoke(func(ctx context.Context, s *my_service.Service) error {
			assert.Panic(t, func() { _ = s.DoA(ctx) }, "ModelA is nil")
			return s.DoB(ctx)
		}, gs.ValueArg(t.Context()))
		assert.Nil(t, err)
	}
}
