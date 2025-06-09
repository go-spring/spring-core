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
	"github.com/go-spring/spring-core/util"
)

// TimeNow is a function that can be overridden to provide custom timestamp behavior (e.g., for testing).
var TimeNow func(ctx context.Context) time.Time

// StringFromContext can be set to extract a string from the context.
var StringFromContext func(ctx context.Context) string

// FieldsFromContext can be set to extract structured fields from the context (e.g., trace IDs, user IDs).
var FieldsFromContext func(ctx context.Context) []Field

// Fields converts a map of string keys to any values to a slice of Field.
func Fields(fields map[string]any) []Field {
	var ret []Field
	for _, k := range util.OrderedMapKeys(fields) {
		ret = append(ret, Any(k, fields[k]))
	}
	return ret
}

// Trace logs a message at TraceLevel using a tag and a lazy field-generating function.
func Trace(ctx context.Context, tag *Tag, fn func() []Field) {
	if tag.GetLogger().EnableLevel(TraceLevel) {
		Record(ctx, TraceLevel, tag, fn()...)
	}
}

// Tracef logs a message at TraceLevel using a tag and a formatted message.
func Tracef(ctx context.Context, tag *Tag, format string, args ...any) {
	if tag.GetLogger().EnableLevel(TraceLevel) {
		Record(ctx, TraceLevel, tag, Msgf(format, args...))
	}
}

// Debug logs a message at DebugLevel using a tag and a lazy field-generating function.
func Debug(ctx context.Context, tag *Tag, fn func() []Field) {
	if tag.GetLogger().EnableLevel(DebugLevel) {
		Record(ctx, DebugLevel, tag, fn()...)
	}
}

// Debugf logs a message at DebugLevel using a tag and a formatted message.
func Debugf(ctx context.Context, tag *Tag, format string, args ...any) {
	if tag.GetLogger().EnableLevel(DebugLevel) {
		Record(ctx, DebugLevel, tag, Msgf(format, args...))
	}
}

// Info logs a message at InfoLevel using structured fields.
func Info(ctx context.Context, tag *Tag, fields ...Field) {
	Record(ctx, InfoLevel, tag, fields...)
}

// Infof logs a message at InfoLevel using a formatted message.
func Infof(ctx context.Context, tag *Tag, format string, args ...any) {
	Record(ctx, InfoLevel, tag, Msgf(format, args...))
}

// Warn logs a message at WarnLevel using structured fields.
func Warn(ctx context.Context, tag *Tag, fields ...Field) {
	Record(ctx, WarnLevel, tag, fields...)
}

// Warnf logs a message at WarnLevel using a formatted message.
func Warnf(ctx context.Context, tag *Tag, format string, args ...any) {
	Record(ctx, WarnLevel, tag, Msgf(format, args...))
}

// Error logs a message at ErrorLevel using structured fields.
func Error(ctx context.Context, tag *Tag, fields ...Field) {
	Record(ctx, ErrorLevel, tag, fields...)
}

// Errorf logs a message at ErrorLevel using a formatted message.
func Errorf(ctx context.Context, tag *Tag, format string, args ...any) {
	Record(ctx, ErrorLevel, tag, Msgf(format, args...))
}

// Panic logs a message at PanicLevel using structured fields.
func Panic(ctx context.Context, tag *Tag, fields ...Field) {
	Record(ctx, PanicLevel, tag, fields...)
}

// Panicf logs a message at PanicLevel using a formatted message.
func Panicf(ctx context.Context, tag *Tag, format string, args ...any) {
	Record(ctx, PanicLevel, tag, Msgf(format, args...))
}

// Fatal logs a message at FatalLevel using structured fields.
func Fatal(ctx context.Context, tag *Tag, fields ...Field) {
	Record(ctx, FatalLevel, tag, fields...)
}

// Fatalf logs a message at FatalLevel using a formatted message.
func Fatalf(ctx context.Context, tag *Tag, format string, args ...any) {
	Record(ctx, FatalLevel, tag, Msgf(format, args...))
}

// Record is the core function that handles publishing log events.
// It checks the logger level, captures caller information, gathers context fields,
// and sends the log event to the logger.
func Record(ctx context.Context, level Level, tag *Tag, fields ...Field) {
	logger := tag.GetLogger()
	if !logger.EnableLevel(level) {
		return // Skip if the logger doesn't allow this level
	}

	file, line := internal.Caller(2, true)

	// Determine the log timestamp
	now := time.Now()
	if TimeNow != nil {
		now = TimeNow(ctx)
	}

	// Extract a string from the context
	var ctxString string
	if StringFromContext != nil {
		ctxString = StringFromContext(ctx)
	}

	// Extract contextual fields from the context
	var ctxFields []Field
	if FieldsFromContext != nil {
		ctxFields = FieldsFromContext(ctx)
	}

	e := GetEvent()
	e.Level = level
	e.Time = now
	e.File = file
	e.Line = line
	e.Tag = tag.GetName()
	e.Fields = fields
	e.CtxString = ctxString
	e.CtxFields = ctxFields

	logger.Publish(e)
}
