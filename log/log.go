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

var DefaultMarker = RegisterMarker("")

// Trace outputs log with level TraceLevel.
func Trace(ctx context.Context, marker *Marker, fn func() []Field) {
	if marker.GetLogger().enableLevel(TraceLevel) {
		Record(ctx, TraceLevel, marker, fn()...)
	}
}

// Debug outputs log with level DebugLevel.
func Debug(ctx context.Context, marker *Marker, fn func() []Field) {
	if marker.GetLogger().enableLevel(DebugLevel) {
		Record(ctx, DebugLevel, marker, fn()...)
	}
}

// Info outputs log with level InfoLevel.
func Info(ctx context.Context, marker *Marker, fields ...Field) {
	Record(ctx, InfoLevel, marker, fields...)
}

// Infof outputs log with level InfoLevel.
func Infof(format string, args ...interface{}) {
	Record(context.Background(), InfoLevel, DefaultMarker, Msgf(format, args...))
}

// Warn outputs log with level WarnLevel.
func Warn(ctx context.Context, marker *Marker, fields ...Field) {
	Record(ctx, WarnLevel, marker, fields...)
}

// Warnf outputs log with level WarnLevel.
func Warnf(format string, args ...interface{}) {
	Record(context.Background(), WarnLevel, DefaultMarker, Msgf(format, args...))
}

// Error outputs log with level ErrorLevel.
func Error(ctx context.Context, marker *Marker, fields ...Field) {
	Record(ctx, ErrorLevel, marker, fields...)
}

// Errorf outputs log with level ErrorLevel.
func Errorf(format string, args ...interface{}) {
	Record(context.Background(), ErrorLevel, DefaultMarker, Msgf(format, args...))
}

// Panic outputs log with level PanicLevel.
func Panic(ctx context.Context, marker *Marker, fields ...Field) {
	Record(ctx, PanicLevel, marker, fields...)
}

// Panicf outputs log with level PanicLevel.
func Panicf(format string, args ...interface{}) {
	Record(context.Background(), PanicLevel, DefaultMarker, Msgf(format, args...))
}

// Fatal outputs log with level FatalLevel.
func Fatal(ctx context.Context, marker *Marker, fields ...Field) {
	Record(ctx, FatalLevel, marker, fields...)
}

// Fatalf outputs log with level FatalLevel.
func Fatalf(format string, args ...interface{}) {
	Record(context.Background(), FatalLevel, DefaultMarker, Msgf(format, args...))
}

// TimeNow is the function used to get the current time.
var TimeNow func(ctx context.Context) time.Time

// Record outputs log.
func Record(ctx context.Context, level Level, marker *Marker, fields ...Field) {
	logger := marker.GetLogger()
	if !logger.enableLevel(level) {
		return
	}
	now := time.Now()
	if TimeNow != nil {
		now = TimeNow(ctx)
	}
	file, line, _ := internal.Caller(2, true)
	e := &Event{
		Marker:  marker.GetName(),
		Time:    now,
		Context: ctx,
		File:    file,
		Line:    line,
		Level:   level,
		Fields:  fields,
	}
	logger.publish(e)
}
