/*
 * Copyright 2024 The Go-Spring Authors.
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

/*
Package syslog provides simplified logging utilities for tracking the execution flow
of the go-spring framework. It is designed to offer a more convenient interface than
the standard library's slog package, whose Info, Warn, and related methods can be
cumbersome to use. Logs produced by this package are typically output to the console.
*/
package syslog

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime"
	"time"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Flags() | log.Lshortfile)
}

// Debugf logs a debug-level message using slog.
func Debugf(format string, a ...any) {
	logMsg(slog.LevelDebug, format, a...)
}

// Infof logs an info-level message using slog.
func Infof(format string, a ...any) {
	logMsg(slog.LevelInfo, format, a...)
}

// Warnf logs a warning-level message using slog.
func Warnf(format string, a ...any) {
	logMsg(slog.LevelWarn, format, a...)
}

// Errorf logs an error-level message using slog.
func Errorf(format string, a ...any) {
	logMsg(slog.LevelError, format, a...)
}

// logMsg constructs and logs a message at the specified log level.
func logMsg(level slog.Level, format string, a ...any) {
	ctx := context.Background()
	if !slog.Default().Enabled(ctx, level) {
		return
	}

	// skip [runtime.Callers, syslog.logMsg, syslog.*f]
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])

	msg := fmt.Sprintf(format, a...)
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	err := slog.Default().Handler().Handle(ctx, r)
	_ = err // ignore error
}
