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
	"fmt"
	"strings"
	"testing"

	"github.com/lvan100/go-assert"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		str     string
		want    Level
		wantErr error
	}{
		{
			str:  "none",
			want: NoneLevel,
		},
		{
			str:  "trace",
			want: TraceLevel,
		},
		{
			str:  "debug",
			want: DebugLevel,
		},
		{
			str:  "info",
			want: InfoLevel,
		},
		{
			str:  "warn",
			want: WarnLevel,
		},
		{
			str:  "error",
			want: ErrorLevel,
		},
		{
			str:  "panic",
			want: PanicLevel,
		},
		{
			str:  "fatal",
			want: FatalLevel,
		},
		{
			str:     "unknown",
			want:    Level(-1),
			wantErr: fmt.Errorf("invalid level unknown"),
		},
	}
	for _, tt := range tests {
		got, err := ParseLevel(tt.str)
		assert.That(t, got).Equal(tt.want)
		assert.That(t, err).Equal(tt.wantErr)
		if tt.str == "unknown" {
			assert.ThatString(t, got.String()).Equal("INVALID")
		} else {
			assert.ThatString(t, got.String()).Equal(strings.ToUpper(tt.str))
		}
	}
}
