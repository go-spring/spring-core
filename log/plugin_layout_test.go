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

package log

import (
	"testing"
	"time"

	"github.com/lvan100/go-assert"
)

func TestTextLayout(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		layout := &TextLayout{}
		b, err := layout.ToBytes(&Event{
			Level:     InfoLevel,
			Time:      time.Time{},
			File:      "file.go",
			Line:      100,
			Tag:       "_def",
			Fields:    []Field{Msg("hello world")},
			CtxFields: nil,
		})
		assert.Nil(t, err)
		assert.ThatString(t, string(b)).Equal("[INFO][0001-01-01T00:00:00.000][file.go:100] _def||msg=hello world\n")
	})
}

func TestJSONLayout(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		layout := &JSONLayout{}
		b, err := layout.ToBytes(&Event{
			Level:     InfoLevel,
			Time:      time.Time{},
			File:      "file.go",
			Line:      100,
			Tag:       "_def",
			Fields:    []Field{Msg("hello world")},
			CtxFields: nil,
		})
		assert.Nil(t, err)
		assert.ThatString(t, string(b)).Equal(`{"level":"info","time":"0001-01-01T00:00:00.000","fileLine":"file.go:100","tag":"_def","msg":"hello world"}` + "\n")
	})
}
