/*
 * Copyright 2012-2019 the original author or authors.
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

package util

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
)

var frameMap sync.Map

// FileLine returns the file name and line of the call point.
// In reality FileLine here costs less time than debug.Stack.
func FileLine() string {
	rpc := make([]uintptr, 1)
	runtime.Callers(3, rpc[:])
	pc := rpc[0]
	if v, ok := frameMap.Load(pc); ok {
		e := v.(*runtime.Frame)
		return fmt.Sprintf("%s:%d", e.File, e.Line)
	}
	e, _ := runtime.CallersFrames(rpc).Next()
	frameMap.Store(pc, &e)
	return fmt.Sprintf("%s:%d", e.File, e.Line)
}

// ForbiddenMethod throws this error when calling a method is prohibited.
var ForbiddenMethod = errors.New("forbidden method")

// UnimplementedMethod throws this error when calling an unimplemented method.
var UnimplementedMethod = errors.New("unimplemented method")

var WrapFormat = func(err error, fileline string, format string, a ...interface{}) error {
	if err == nil {
		if format != "" {
			return fmt.Errorf(fileline+" "+format, a...)
		}
		return errors.New(fileline + " " + fmt.Sprint(a...))
	}
	if format == "" {
		return fmt.Errorf("%s %s; %w", fileline, fmt.Sprint(a...), err)
	}
	return fmt.Errorf("%s %s; %w", fileline, fmt.Sprintf(format, a...), err)
}

// Error returns an error with the file and line.
// The file and line may be calculated at the compile time in the future.
func Error(fileline string, text string) error {
	return WrapFormat(nil, fileline, "", text)
}

// Errorf returns an error with the file and line.
// The file and line may be calculated at the compile time in the future.
func Errorf(fileline string, format string, a ...interface{}) error {
	return WrapFormat(nil, fileline, format, a...)
}

// Wrap returns an error with the file and line.
// The file and line may be calculated at the compile time in the future.
func Wrap(err error, fileline string, text string) error {
	return WrapFormat(err, fileline, "", text)
}

// Wrapf returns an error with the file and line.
// The file and line may be calculated at the compile time in the future.
func Wrapf(err error, fileline string, format string, a ...interface{}) error {
	return WrapFormat(err, fileline, format, a...)
}
