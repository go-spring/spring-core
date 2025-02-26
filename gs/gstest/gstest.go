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

package gstest

import (
	"testing"

	"github.com/go-spring/spring-core/gs"
)

var ctx = &gs.ContextAware{}

func init() {
	gs.ForceAutowireIsNullable(true)
}

// Init initializes the test environment.
func Init() error {
	gs.Object(ctx)
	return gs.Start()
}

// Run executes test cases and ensures shutdown of the app context.
func Run(m *testing.M) (code int) {
	defer func() { gs.Stop() }()
	return m.Run()
}

// Keys retrieves all the property keys.
func Keys() []string {
	return ctx.GSContext.Keys()
}

// Has checks whether a specific property exists.
func Has(key string) bool {
	return ctx.GSContext.Has(key)
}

// SubKeys retrieves the sub-keys of a specified key.
func SubKeys(key string) ([]string, error) {
	return ctx.GSContext.SubKeys(key)
}

// Prop retrieves the value of a property specified by the key.
func Prop(key string, def ...string) string {
	return ctx.GSContext.Prop(key, def...)
}

// Resolve resolves a given string with placeholders.
func Resolve(s string) (string, error) {
	return ctx.GSContext.Resolve(s)
}

// Bind binds an object to the properties.
func Bind(i interface{}, tag ...string) error {
	return ctx.GSContext.Bind(i, tag...)
}

// Get retrieves an object using specified selectors.
func Get(i interface{}, tag ...string) error {
	return ctx.GSContext.Get(i, tag...)
}

// Wire injects dependencies into an object or constructor.
func Wire(objOrCtor interface{}, ctorArgs ...gs.Arg) (interface{}, error) {
	return ctx.GSContext.Wire(objOrCtor, ctorArgs...)
}

// Invoke calls a function with arguments injected.
func Invoke(fn interface{}, args ...gs.Arg) ([]interface{}, error) {
	return ctx.GSContext.Invoke(fn, args...)
}
