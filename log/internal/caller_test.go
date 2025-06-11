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

package internal_test

import (
	"testing"

	"github.com/go-spring/spring-core/log/internal"
	"github.com/lvan100/go-assert"
)

func TestCaller(t *testing.T) {

	t.Run("error skip", func(t *testing.T) {
		file, line := internal.Caller(100, true)
		assert.ThatString(t, file).Equal("")
		assert.That(t, line).Equal(0)
	})

	t.Run("fast false", func(t *testing.T) {
		file, line := internal.Caller(0, false)
		assert.ThatString(t, file).Matches(".*/caller_test.go")
		assert.That(t, line).Equal(35)
	})

	t.Run("fast true", func(t *testing.T) {
		for range 2 {
			file, line := internal.Caller(0, true)
			assert.ThatString(t, file).Matches(".*/caller_test.go")
			assert.That(t, line).Equal(42)
		}
	})
}

func BenchmarkCaller(b *testing.B) {

	// BenchmarkCaller/fast-8  12433761  95.05 ns/op
	// BenchmarkCaller/slow-8   6314623  190.3 ns/op

	b.Run("fast", func(b *testing.B) {
		for b.Loop() {
			internal.Caller(0, true)
		}
	})

	b.Run("slow", func(b *testing.B) {
		for b.Loop() {
			internal.Caller(0, false)
		}
	})
}
