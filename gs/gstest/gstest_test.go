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

package gstest_test

import (
	"fmt"
	"testing"

	"github.com/go-spring/spring-core/gs/gstest"
	"github.com/go-spring/spring-core/gs/gstest/testdata/app"
	"github.com/go-spring/spring-core/gs/gstest/testdata/biz"
	"github.com/lvan100/go-assert"
)

func init() {
	gstest.MockFor[*app.App]().With(&app.App{Name: "test"})
}

func TestMain(m *testing.M) {
	var opts []gstest.RunOption
	opts = append(opts, gstest.BeforeRun(func() {
		fmt.Println("before run")
	}))
	opts = append(opts, gstest.AfterRun(func() {
		fmt.Println("after run")
	}))
	gstest.TestMain(m, opts...)
}

func TestGSTest(t *testing.T) {
	a := gstest.Get[*app.App](t)
	assert.Equal(t, a.Name, "test")
	s := gstest.Wire(t, new(struct {
		App     *app.App     `autowire:""`
		Service *biz.Service `autowire:""`
	}))
	assert.Equal(t, s.App.Name, "test")
	assert.Equal(t, s.Service.Hello("xyz"), "hello xyz")
}
