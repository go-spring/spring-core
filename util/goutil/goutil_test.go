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

package goutil_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-spring/spring-core/util/goutil"
)

func TestGo(t *testing.T) {

	goutil.Go(t.Context(), func(ctx context.Context) {
		fmt.Println("hello")
	}).Wait()

	goutil.GoFunc(func() {
		fmt.Println("hello")
	}).Wait()
}

func TestGoValue(t *testing.T) {

	s, err := goutil.GoValue(t.Context(), func(ctx context.Context) (string, error) {
		return "hello", nil
	}).Wait()
	fmt.Println(s, err)

	var arr []*goutil.ValueStatus[int]
	for i := 0; i < 3; i++ {
		arr = append(arr, goutil.GoValue(t.Context(), func(ctx context.Context) (int, error) {
			return i, nil
		}))
	}
	for _, g := range arr {
		fmt.Println(g.Wait())
	}
}
