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
	"errors"
	"testing"
	"time"

	"github.com/go-spring/spring-base/testing/assert"
	"github.com/go-spring/spring-core/util/goutil"
)

func TestGo(t *testing.T) {

	var s string
	goutil.Go(t.Context(), func(ctx context.Context) {
		panic("something is wrong")
	}).Wait()
	assert.That(t, s).Equal("")

	goutil.Go(t.Context(), func(ctx context.Context) {
		s = "hello world!"
	}).Wait()
	assert.That(t, s).Equal("hello world!")
}

func TestGoValue(t *testing.T) {

	s, err := goutil.GoValue(t.Context(), func(ctx context.Context) (string, error) {
		panic("something is wrong")
	}).Wait()
	assert.That(t, s).Equal("")
	assert.Error(t, err).Matches("panic recovered: .*")

	i, err := goutil.GoValue(t.Context(), func(ctx context.Context) (int, error) {
		return 42, nil
	}).Wait()
	assert.That(t, err).Nil()
	assert.That(t, i).Equal(42)

	s, err = goutil.GoValue(t.Context(), func(ctx context.Context) (string, error) {
		return "hello world!", nil
	}).Wait()
	assert.That(t, err).Nil()
	assert.That(t, s).Equal("hello world!")

	var arr []*goutil.ValueStatus[int]
	for i := range 3 {
		arr = append(arr, goutil.GoValue(t.Context(), func(ctx context.Context) (int, error) {
			return i, nil
		}))
	}
	for i, g := range arr {
		v, err := g.Wait()
		assert.That(t, v).Equal(i)
		assert.That(t, err).Nil()
	}

	expectedErr := errors.New("expected error")
	_, err = goutil.GoValue(t.Context(), func(ctx context.Context) (string, error) {
		return "", expectedErr
	}).Wait()
	assert.That(t, err).Equal(expectedErr)

	t.Run("context cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()
		_, err := goutil.GoValue(ctx, func(ctx context.Context) (string, error) {
			time.Sleep(100 * time.Millisecond)
			return "done", nil
		}).Wait()
		assert.That(t, err).Nil()
	})
}
