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

	"github.com/go-spring/spring-base/testing/assert"
	"github.com/go-spring/spring-core/gs"
)

func init() {
	gs.Object(&GreetingService{})
	gs.AddTester(&GreetingTester{})
}

func init() {
	gs.MockFor[*GreetingService]().With(
		&GreetingService{
			Message: "Hello, World!",
		},
	)
}

func TestMain(m *testing.M) {
	gs.TestMain(m)
}

func TestGreeting(t *testing.T) {
	fmt.Println(t.Name()) // TestGreeting
}

func TestFarewell(t *testing.T) {
	fmt.Println(t.Name()) // TestFarewell
}

type GreetingService struct {
	Message string `value:"${app.greeting:=Hello, Go-Spring!}"`
}

type GreetingTester struct {
	Svc *GreetingService `autowire:""`
}

func (u *GreetingTester) TestGreeting(t *testing.T) {
	fmt.Println(t.Name()) // gs_test.GreetingTester.TestFarewell
	assert.That(t, u.Svc).NotNil()
	assert.ThatString(t, u.Svc.Message).Equal("Hello, World!")
}

func (u *GreetingTester) TestFarewell(t *testing.T) {
	fmt.Println(t.Name()) // gs_test.GreetingTester.TestGreeting
	assert.That(t, u.Svc).NotNil()
	assert.ThatString(t, u.Svc.Message).Equal("Hello, World!")
}
