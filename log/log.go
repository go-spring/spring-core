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
	"context"
	"time"

	"github.com/go-spring/spring-core/log/internal"
)

// DefaultMarker is the default logging marker used when none is specified.
var DefaultMarker = RegisterMarker("_def")

func init() {
	DefaultMarker.SetLogger(&Logger{
		privateConfig: &LoggerConfig{
			baseLoggerConfig: baseLoggerConfig{
				Level: InfoLevel,
				AppenderRefs: []*AppenderRef{
					{
						appender: &ConsoleAppender{
							BaseAppender: BaseAppender{
								Layout: &TextLayout{},
							},
						},
					},
				},
			},
		},
	})
}

// TimeNow is a function that can be overridden to provide custom timestamp behavior (e.g., for testing).
var TimeNow func(ctx context.Context) time.Time

// FieldsFromContext can be set to extract structured fields from the context (e.g., trace IDs, user IDs).
var FieldsFromContext func(ctx context.Context) []Field

// Trace logs a message at TraceLevel using a marker and a lazy field-generating function.
func Trace(ctx context.Context, marker *Marker, fn func() []Field) {
	if marker.GetLogger().enableLevel(TraceLevel) {
		Record(ctx, TraceLevel, marker, fn()...)
	}
}

// Debug logs a message at DebugLevel using a marker and a lazy field-generating function.
func Debug(ctx context.Context, marker *Marker, fn func() []Field) {
	if marker.GetLogger().enableLevel(DebugLevel) {
		Record(ctx, DebugLevel, marker, fn()...)
	}
}

// Info logs a message at InfoLevel using structured fields.
func Info(ctx context.Context, marker *Marker, fields ...Field) {
	Record(ctx, InfoLevel, marker, fields...)
}

// Infof logs a formatted message at InfoLevel using the default marker.
func Infof(format string, args ...interface{}) {
	Record(context.Background(), InfoLevel, DefaultMarker, Msgf(format, args...))
}

// Warn logs a message at WarnLevel using structured fields.
func Warn(ctx context.Context, marker *Marker, fields ...Field) {
	Record(ctx, WarnLevel, marker, fields...)
}

// Warnf logs a formatted message at WarnLevel using the default marker.
func Warnf(format string, args ...interface{}) {
	Record(context.Background(), WarnLevel, DefaultMarker, Msgf(format, args...))
}

// Error logs a message at ErrorLevel using structured fields.
func Error(ctx context.Context, marker *Marker, fields ...Field) {
	Record(ctx, ErrorLevel, marker, fields...)
}

// Errorf logs a formatted message at ErrorLevel using the default marker.
func Errorf(format string, args ...interface{}) {
	Record(context.Background(), ErrorLevel, DefaultMarker, Msgf(format, args...))
}

// Panic logs a message at PanicLevel using structured fields.
func Panic(ctx context.Context, marker *Marker, fields ...Field) {
	Record(ctx, PanicLevel, marker, fields...)
}

// Panicf logs a formatted message at PanicLevel using the default marker.
func Panicf(format string, args ...interface{}) {
	Record(context.Background(), PanicLevel, DefaultMarker, Msgf(format, args...))
}

// Fatal logs a message at FatalLevel using structured fields.
func Fatal(ctx context.Context, marker *Marker, fields ...Field) {
	Record(ctx, FatalLevel, marker, fields...)
}

// Fatalf logs a formatted message at FatalLevel using the default marker.
func Fatalf(format string, args ...interface{}) {
	Record(context.Background(), FatalLevel, DefaultMarker, Msgf(format, args...))
}

// Record is the core function that handles publishing log events.
// It checks the logger level, captures caller information, gathers context fields,
// and sends the log event to the logger.
func Record(ctx context.Context, level Level, marker *Marker, fields ...Field) {
	logger := marker.GetLogger()
	if !logger.enableLevel(level) {
		return // Skip if the logger doesn't allow this level
	}

	file, line, _ := internal.Caller(2, true)

	// Determine the log timestamp
	now := time.Now()
	if TimeNow != nil {
		now = TimeNow(ctx)
	}

	// Extract contextual fields from the context
	var ctxFields []Field
	if FieldsFromContext != nil {
		ctxFields = FieldsFromContext(ctx)
	}

	logger.publish(&Event{
		Level:     level,
		Time:      now,
		File:      file,
		Line:      line,
		Marker:    marker.GetName(),
		Fields:    fields,
		CtxFields: ctxFields,
	})
}
