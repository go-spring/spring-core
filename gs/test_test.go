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
	// Register GreetingService as a regular bean in the IoC container.
	gs.Object(&GreetingService{})

	// Register GreetingTester as a tester so its test methods will be discovered
	// and executed automatically by gs.TestMain.
	gs.AddTester(&GreetingTester{})
}

func init() {
	// Replace the real GreetingService bean with a mock version for testing.
	gs.MockFor[*GreetingService]().With(
		&GreetingService{
			Message: "Hello, World!",
		},
	)
}

func TestMain(m *testing.M) {
	gs.TestMain(m)
}

// TestGreeting is a regular Go test function.
// It is discovered by the standard testing framework.
func TestGreeting(t *testing.T) {
	fmt.Println(t.Name()) // Expected output: TestGreeting
}

// TestFarewell is another regular Go test function.
func TestFarewell(t *testing.T) {
	fmt.Println(t.Name()) // Expected output: TestFarewell
}

type GreetingService struct {
	Message string `value:"${app.greeting:=Hello, Go-Spring!}"`
}

// GreetingTester is a tester struct used to define test methods
// that will be automatically registered and run by gs.TestMain.
// The Svc field is injected (autowired) from the IoC container.
type GreetingTester struct {
	Svc *GreetingService `autowire:""`
}

// TestGreeting is a test method defined on GreetingTester.
// It will be automatically discovered by gs.TestMain.
func (u *GreetingTester) TestGreeting(t *testing.T) {
	fmt.Println(t.Name()) // Expected output: gs_test.GreetingTester.TestGreeting
	assert.That(t, u.Svc).NotNil()
	assert.String(t, u.Svc.Message).Equal("Hello, World!")
}

// TestFarewell is another test method defined on GreetingTester.
// Like TestGreeting, it is automatically discovered by gs.TestMain.
func (u *GreetingTester) TestFarewell(t *testing.T) {
	fmt.Println(t.Name()) // Expected output: gs_test.GreetingTester.TestFarewell
	assert.That(t, u.Svc).NotNil()
	assert.String(t, u.Svc.Message).Equal("Hello, World!")
}
