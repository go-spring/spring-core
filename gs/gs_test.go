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

package gs_test

import (
	"fmt"
	"testing"

	"github.com/go-spring/spring-core/gs"
)

func TestMain(m *testing.M) {
	gs.AddTester(&Tester{})
	gs.Object(&Dep{})
	gs.TestMain(m)
}

func TestAA(t *testing.T) {

}

type Dep struct {
	Name string `value:"${name:=TestAA}"`
}

type Tester struct {
	Dep *Dep `autowire:""`
}

func (o *Tester) TestAA(t *testing.T) {
	fmt.Println(o.Dep.Name)
}
