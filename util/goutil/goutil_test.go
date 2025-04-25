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

package goutil_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-spring/spring-core/util/goutil"
	"github.com/lvan100/go-assert"
)

func TestGo(t *testing.T) {

	var s string
	goutil.Go(t.Context(), func(ctx context.Context) {
		panic("something is wrong")
	}).Wait()
	assert.Equal(t, s, "")

	goutil.Go(t.Context(), func(ctx context.Context) {
		s = "hello world!"
	}).Wait()
	assert.Equal(t, s, "hello world!")
}

func TestGoFunc(t *testing.T) {

	var s string
	goutil.GoFunc(func() {
		panic("something is wrong")
	}).Wait()
	assert.Equal(t, s, "")

	goutil.GoFunc(func() {
		s = "hello world!"
	}).Wait()
	assert.Equal(t, s, "hello world!")
}

func TestGoValue(t *testing.T) {

	s, err := goutil.GoValue(t.Context(), func(ctx context.Context) (string, error) {
		panic("something is wrong")
	}).Wait()
	assert.Equal(t, s, "")
	assert.Equal(t, err, fmt.Errorf("panic occurred"))

	s, err = goutil.GoValue(t.Context(), func(ctx context.Context) (string, error) {
		return "hello world!", nil
	}).Wait()
	assert.Equal(t, s, "hello world!")
	assert.Nil(t, err)

	var arr []*goutil.ValueStatus[int]
	for i := 0; i < 3; i++ {
		arr = append(arr, goutil.GoValue(t.Context(), func(ctx context.Context) (int, error) {
			return i, nil
		}))
	}

	var v int
	for i, g := range arr {
		v, err = g.Wait()
		assert.Equal(t, v, i)
		assert.Nil(t, err)
	}
}
