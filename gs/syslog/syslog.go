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
	"log/slog"
	"os"
)

func init() {
	slog.Default().Enabled(context.Background(), slog.LevelInfo)
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
}

// Debug calls [Logger.Debug] on the default logger.
func Debug(msg string, args ...any) {
	slog.Default().Log(context.Background(), slog.LevelDebug, msg, args...)
}

// Info calls [Logger.Info] on the default logger.
func Info(msg string, args ...any) {
	slog.Default().Log(context.Background(), slog.LevelInfo, msg, args...)
}

// Warn calls [Logger.Warn] on the default logger.
func Warn(msg string, args ...any) {
	slog.Default().Log(context.Background(), slog.LevelWarn, msg, args...)
}

// Error calls [Logger.Error] on the default logger.
func Error(msg string, args ...any) {
	slog.Default().Log(context.Background(), slog.LevelError, msg, args...)
}
