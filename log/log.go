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

// TimeNow is a function that can be overridden to provide custom timestamp behavior (e.g., for testing).
var TimeNow func(ctx context.Context) time.Time

// StringFromContext can be set to extract a string from the context.
var StringFromContext func(ctx context.Context) string

// FieldsFromContext can be set to extract structured fields from the context (e.g., trace IDs, user IDs).
var FieldsFromContext func(ctx context.Context) []Field

// Trace logs a message at TraceLevel using a tag and a lazy field-generating function.
func Trace(ctx context.Context, tag *Tag, fn func() []Field) {
	if tag.GetLogger().enableLevel(TraceLevel) {
		Record(ctx, TraceLevel, tag, fn()...)
	}
}

// Debug logs a message at DebugLevel using a tag and a lazy field-generating function.
func Debug(ctx context.Context, tag *Tag, fn func() []Field) {
	if tag.GetLogger().enableLevel(DebugLevel) {
		Record(ctx, DebugLevel, tag, fn()...)
	}
}

// Info logs a message at InfoLevel using structured fields.
func Info(ctx context.Context, tag *Tag, fields ...Field) {
	Record(ctx, InfoLevel, tag, fields...)
}

// Warn logs a message at WarnLevel using structured fields.
func Warn(ctx context.Context, tag *Tag, fields ...Field) {
	Record(ctx, WarnLevel, tag, fields...)
}

// Error logs a message at ErrorLevel using structured fields.
func Error(ctx context.Context, tag *Tag, fields ...Field) {
	Record(ctx, ErrorLevel, tag, fields...)
}

// Panic logs a message at PanicLevel using structured fields.
func Panic(ctx context.Context, tag *Tag, fields ...Field) {
	Record(ctx, PanicLevel, tag, fields...)
}

// Fatal logs a message at FatalLevel using structured fields.
func Fatal(ctx context.Context, tag *Tag, fields ...Field) {
	Record(ctx, FatalLevel, tag, fields...)
}

// Record is the core function that handles publishing log events.
// It checks the logger level, captures caller information, gathers context fields,
// and sends the log event to the logger.
func Record(ctx context.Context, level Level, tag *Tag, fields ...Field) {
	logger := tag.GetLogger()
	if !logger.enableLevel(level) {
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

	logger.publish(e)
}
