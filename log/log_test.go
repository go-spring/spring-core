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

package log_test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-spring/spring-core/log"
	"github.com/lvan100/go-assert"
)

var (
	keyTraceID int
	keySpanID  int
)

var TagDefault = log.GetTag("_def")
var TagRequestIn = log.GetTag("_com_request_in")
var TagRequestOut = log.GetTag("_com_request_out")

func TestLog(t *testing.T) {
	ctx := t.Context()

	logBuf := bytes.NewBuffer(nil)
	log.Stdout = logBuf
	defer func() {
		log.Stdout = os.Stdout
	}()

	log.TimeNow = func(ctx context.Context) time.Time {
		return time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	}

	log.StringFromContext = func(ctx context.Context) string {
		return ""
	}

	log.FieldsFromContext = func(ctx context.Context) []log.Field {
		traceID, _ := ctx.Value(&keyTraceID).(string)
		spanID, _ := ctx.Value(&keySpanID).(string)
		return []log.Field{
			log.String("trace_id", traceID),
			log.String("span_id", spanID),
		}
	}

	// not print
	log.Debug(ctx, TagRequestOut, func() []log.Field {
		return []log.Field{
			log.Msgf("hello %s", "world"),
		}
	})

	// print
	log.Info(ctx, TagDefault, log.Msgf("hello %s", "world"))
	log.Info(ctx, TagRequestIn, log.Msgf("hello %s", "world"))

	err := log.RefreshFile("testdata/log.xml")
	assert.Nil(t, err)

	ctx = context.WithValue(ctx, &keyTraceID, "0a882193682db71edd48044db54cae88")
	ctx = context.WithValue(ctx, &keySpanID, "50ef0724418c0a66")

	// print
	log.Trace(ctx, TagRequestOut, func() []log.Field {
		return []log.Field{
			log.Msgf("hello %s", "world"),
		}
	})

	// print
	log.Debug(ctx, TagRequestOut, func() []log.Field {
		return []log.Field{
			log.Msgf("hello %s", "world"),
		}
	})

	// print
	log.Info(ctx, TagRequestIn, log.Msgf("hello %s", "world"))
	log.Warn(ctx, TagRequestIn, log.Msgf("hello %s", "world"))
	log.Error(ctx, TagRequestIn, log.Msgf("hello %s", "world"))
	log.Panic(ctx, TagRequestIn, log.Msgf("hello %s", "world"))
	log.Fatal(ctx, TagRequestIn, log.Msgf("hello %s", "world"))

	// print
	log.Infof(ctx, TagRequestIn, "hello %s", "world")
	log.Warnf(ctx, TagRequestIn, "hello %s", "world")
	log.Errorf(ctx, TagRequestIn, "hello %s", "world")
	log.Panicf(ctx, TagRequestIn, "hello %s", "world")
	log.Fatalf(ctx, TagRequestIn, "hello %s", "world")

	// not print
	log.Info(ctx, TagDefault, log.Msgf("hello %s", "world"))

	// print
	log.Warn(ctx, TagDefault, log.Msgf("hello %s", "world"))
	log.Error(ctx, TagDefault, log.Msgf("hello %s", "world"))
	log.Panic(ctx, TagDefault, log.Msgf("hello %s", "world"))

	// print
	log.Error(ctx, TagDefault, log.Fields(map[string]any{
		"key1": "value1",
		"key2": "value2",
	})...)

	expectLog := `
[INFO][2025-06-01T00:00:00.000][/Users/didi/spring-core/log/log_test.go:74] _def||trace_id=||span_id=||msg=hello world
[INFO][2025-06-01T00:00:00.000][/Users/didi/spring-core/log/log_test.go:75] _com_request_in||trace_id=||span_id=||msg=hello world
{"level":"trace","time":"2025-06-01T00:00:00.000","fileLine":"/Users/didi/spring-core/log/log_test.go:84","tag":"_com_request_out","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"debug","time":"2025-06-01T00:00:00.000","fileLine":"/Users/didi/spring-core/log/log_test.go:91","tag":"_com_request_out","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"info","time":"2025-06-01T00:00:00.000","fileLine":"/Users/didi/spring-core/log/log_test.go:98","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"warn","time":"2025-06-01T00:00:00.000","fileLine":"/Users/didi/spring-core/log/log_test.go:99","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"error","time":"2025-06-01T00:00:00.000","fileLine":"/Users/didi/spring-core/log/log_test.go:100","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"panic","time":"2025-06-01T00:00:00.000","fileLine":"/Users/didi/spring-core/log/log_test.go:101","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"fatal","time":"2025-06-01T00:00:00.000","fileLine":"/Users/didi/spring-core/log/log_test.go:102","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"info","time":"2025-06-01T00:00:00.000","fileLine":"/Users/didi/spring-core/log/log_test.go:105","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"warn","time":"2025-06-01T00:00:00.000","fileLine":"/Users/didi/spring-core/log/log_test.go:106","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"error","time":"2025-06-01T00:00:00.000","fileLine":"/Users/didi/spring-core/log/log_test.go:107","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"panic","time":"2025-06-01T00:00:00.000","fileLine":"/Users/didi/spring-core/log/log_test.go:108","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
{"level":"fatal","time":"2025-06-01T00:00:00.000","fileLine":"/Users/didi/spring-core/log/log_test.go:109","tag":"_com_request_in","trace_id":"0a882193682db71edd48044db54cae88","span_id":"50ef0724418c0a66","msg":"hello world"}
[WARN][2025-06-01T00:00:00.000][/Users/didi/spring-core/log/log_test.go:115] _def||trace_id=0a882193682db71edd48044db54cae88||span_id=50ef0724418c0a66||msg=hello world
[ERROR][2025-06-01T00:00:00.000][/Users/didi/spring-core/log/log_test.go:116] _def||trace_id=0a882193682db71edd48044db54cae88||span_id=50ef0724418c0a66||msg=hello world
[PANIC][2025-06-01T00:00:00.000][/Users/didi/spring-core/log/log_test.go:117] _def||trace_id=0a882193682db71edd48044db54cae88||span_id=50ef0724418c0a66||msg=hello world
[ERROR][2025-06-01T00:00:00.000][/Users/didi/spring-core/log/log_test.go:120] _def||trace_id=0a882193682db71edd48044db54cae88||span_id=50ef0724418c0a66||key1=value1||key2=value2
`

	assert.ThatString(t, logBuf.String()).Equal(strings.TrimLeft(expectLog, "\n"))
}
