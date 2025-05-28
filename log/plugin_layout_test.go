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
	"errors"
	"testing"
	"time"

	"github.com/lvan100/go-assert"
)

func TestParseHumanizeBytes(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    HumanizeBytes
		wantErr error
	}{
		{
			name:  "basic bytes",
			input: "1024B",
			want:  1024,
		},
		{
			name:  "kilobytes",
			input: "1KB",
			want:  1024,
		},
		{
			name:  "megabytes",
			input: "2MB",
			want:  2 * 1024 * 1024,
		},
		{
			name:  "case insensitive",
			input: "1kb",
			want:  1024,
		},
		{
			name:  "space before unit",
			input: "1 KB",
			want:  1024,
		},
		{
			name:  "space after unit",
			input: "1KB ",
			want:  1024,
		},
		{
			name:    "invalid number",
			input:   "abcKB",
			wantErr: errors.New(`strconv.ParseInt: parsing "": invalid syntax`),
		},
		{
			name:    "missing unit",
			input:   "1024",
			wantErr: errors.New(`unhandled size name: ""`),
		},
		{
			name:    "unknown unit",
			input:   "1GB",
			wantErr: errors.New(`unhandled size name: "GB"`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHumanizeBytes(tt.input)
			if err != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("ParseHumanizeBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseHumanizeBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTextLayout(t *testing.T) {

	//t.Run("error - encoder begin", func(t *testing.T) {
	//	layout := &TextLayout{
	//		BaseLayout{
	//			FileLineLength: 48,
	//		},
	//	}
	//
	//})

	t.Run("success", func(t *testing.T) {
		layout := &TextLayout{
			BaseLayout{
				FileLineLength: 48,
			},
		}
		b := layout.ToBytes(&Event{
			Level:     InfoLevel,
			Time:      time.Time{},
			File:      "gs/examples/bookman/src/biz/service/book_service/book_service_test.go",
			Line:      100,
			Tag:       "_def",
			Fields:    []Field{Msg("hello world")},
			CtxString: "trace_id=0a882193682db71edd48044db54cae88||span_id=50ef0724418c0a66",
			CtxFields: nil,
		})
		assert.ThatString(t, string(b)).Equal("[INFO][0001-01-01T00:00:00.000][...iz/service/book_service/book_service_test.go:100] _def||trace_id=0a882193682db71edd48044db54cae88||span_id=50ef0724418c0a66||msg=hello world\n")
	})
}

func TestJSONLayout(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		layout := &JSONLayout{
			BaseLayout{
				FileLineLength: 48,
			},
		}
		b := layout.ToBytes(&Event{
			Level:     InfoLevel,
			Time:      time.Time{},
			File:      "gs/examples/bookman/src/biz/service/book_service/book_service_test.go",
			Line:      100,
			Tag:       "_def",
			Fields:    []Field{Msg("hello world")},
			CtxString: "trace_id=0a882193682db71edd48044db54cae88||span_id=50ef0724418c0a66",
			CtxFields: nil,
		})
		assert.ThatString(t, string(b)).Equal(`{"level":"info","time":"0001-01-01T00:00:00.000","fileLine":"...iz/service/book_service/book_service_test.go:100","tag":"_def","ctxString":"trace_id=0a882193682db71edd48044db54cae88||span_id=50ef0724418c0a66","msg":"hello world"}` + "\n")
	})
}
