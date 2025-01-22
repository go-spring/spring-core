/*
 * Copyright 2012-2024 the original author or authors.
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

package syslog

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"runtime"
	"time"
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

// Debugf logs at [slog.LevelDebug].
func Debugf(format string, a ...any) {
	logMsg(slog.LevelDebug, format, a...)
}

// Infof logs at [slog.LevelInfo].
func Infof(format string, a ...any) {
	logMsg(slog.LevelInfo, format, a...)
}

// Warnf logs at [slog.LevelWarn].
func Warnf(format string, a ...any) {
	logMsg(slog.LevelWarn, format, a...)
}

// Errorf logs at [slog.LevelError].
func Errorf(format string, a ...any) {
	logMsg(slog.LevelError, format, a...)
}

func logMsg(level slog.Level, format string, a ...any) {
	ctx := context.Background()
	if !slog.Default().Enabled(ctx, level) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	msg := fmt.Sprintf(format, a...)
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	_ = slog.Default().Handler().Handle(ctx, r)
}
