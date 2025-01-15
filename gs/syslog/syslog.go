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
	"log/slog"
	"os"
)

func init() {
	slog.SetLogLoggerLevel(slog.LevelInfo)
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
}

// Debugf logs at [slog.LevelDebug].
func Debugf(msg string, args ...any) {
	slog.Default().Log(context.Background(), slog.LevelDebug, fmt.Sprintf(msg, args...))
}

// Infof logs at [slog.LevelInfo].
func Infof(msg string, args ...any) {
	slog.Default().Log(context.Background(), slog.LevelInfo, fmt.Sprintf(msg, args...))
}

// Warnf logs at [slog.LevelWarn].
func Warnf(msg string, args ...any) {
	slog.Default().Log(context.Background(), slog.LevelWarn, fmt.Sprintf(msg, args...))
}

// Errorf logs at [slog.LevelError].
func Errorf(msg string, args ...any) {
	slog.Default().Log(context.Background(), slog.LevelError, fmt.Sprintf(msg, args...))
}
