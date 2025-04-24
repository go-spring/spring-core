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

package iterutil

import (
	"iter"
)

// Times executes the function 'fn' exactly 'count' times.
func Times(count int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := 0; i < count; i++ {
			yield(i)
		}
	}
}

// Ranges iterates from 'start' to 'end' (exclusive) and applies 'fn' to each index.
func Ranges(start, end int) iter.Seq[int] {
	if start < end {
		return stepRangesForward(start, end, 1)
	} else if start > end {
		return stepRangesBackward(start, end, -1)
	}
	return func(yield func(int) bool) {}
}

// StepRanges iterates from 'start' to 'end' using a step size and applies 'fn' to each index.
func StepRanges(start, end, step int) iter.Seq[int] {
	if step > 0 && start < end {
		return stepRangesForward(start, end, step)
	} else if step < 0 && start > end {
		return stepRangesBackward(start, end, step)
	}
	return func(yield func(int) bool) {}
}

// stepRangesForward helper function for forward step iteration.
func stepRangesForward(start, end, step int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := start; i < end; i += step {
			yield(i)
		}
	}
}

// stepRangesBackward helper function for backward step iteration.
func stepRangesBackward(start, end, step int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := start; i > end; i += step {
			yield(i)
		}
	}
}
