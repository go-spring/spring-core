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

package value_struct

import (
	"bytes"
	"testing"

	"benchmark-fields/encoder"
)

func BenchmarkBools(b *testing.B) {

	// bools-8      11117685    105.6 ns/op
	// int64s-8      9201427	136.9 ns/op
	// float64s-8    1878270	618.2 ns/op
	// strings-8     8103530	146.1 ns/op

	arrBools := []bool{true, false, true, false, true, false}
	arrInt64s := []int64{1, 2, 3, 4, 5, 6, 7, 8}
	arrFloat64s := []float64{1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8}
	arrStrings := []string{"a", "b", "c", "d", "e", "f", "g", "h"}

	b.ResetTimer()
	b.ReportAllocs()

	b.Run("bools", func(b *testing.B) {
		for b.Loop() {
			v := Bools("arr", arrBools)
			WriteFields(encoder.NewJSONEncoder(bytes.NewBuffer(nil)), []Field{v})
		}
	})

	b.Run("int64s", func(b *testing.B) {
		for b.Loop() {
			v := Int64s("arr", arrInt64s)
			WriteFields(encoder.NewJSONEncoder(bytes.NewBuffer(nil)), []Field{v})
		}
	})

	b.Run("float64s", func(b *testing.B) {
		for b.Loop() {
			v := Float64s("arr", arrFloat64s)
			WriteFields(encoder.NewJSONEncoder(bytes.NewBuffer(nil)), []Field{v})
		}
	})

	b.Run("strings", func(b *testing.B) {
		for b.Loop() {
			v := Strings("arr", arrStrings)
			WriteFields(encoder.NewJSONEncoder(bytes.NewBuffer(nil)), []Field{v})
		}
	})
}
