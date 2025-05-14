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
)

const (
	NoneLevel  Level = iota // No logging
	TraceLevel              // Very detailed logging, typically for debugging at a granular level
	DebugLevel              // Debugging information
	InfoLevel               // General informational messages
	WarnLevel               // Warnings that may indicate a potential problem
	ErrorLevel              // Errors that allow the application to continue running
	PanicLevel              // Severe issues that may lead to a panic
	FatalLevel              // Critical issues that will cause application termination
)

// Level is an enumeration used to identify the severity of a logging event.
type Level int32

func (level Level) String() string {
	switch level {
	case NoneLevel:
		return "NONE"
	case TraceLevel:
		return "TRACE"
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case PanicLevel:
		return "PANIC"
	case FatalLevel:
		return "FATAL"
	default:
		return "INVALID"
	}
}

// ParseLevel converts a string (case-insensitive) into a corresponding Level value.
// Returns an error if the input string does not match any valid level.
func ParseLevel(str string) (Level, error) {
	switch strings.ToUpper(str) {
	case "NONE":
		return NoneLevel, nil
	case "TRACE":
		return TraceLevel, nil
	case "DEBUG":
		return DebugLevel, nil
	case "INFO":
		return InfoLevel, nil
	case "WARN":
		return WarnLevel, nil
	case "ERROR":
		return ErrorLevel, nil
	case "PANIC":
		return PanicLevel, nil
	case "FATAL":
		return FatalLevel, nil
	default:
		return -1, fmt.Errorf("invalid level %s", str)
	}
}
