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
	"strings"
	"sync/atomic"

	"github.com/go-spring/spring-core/util"
)

var tagMap = map[string]*Tag{}

var initLogger = &Logger{
	privateConfig: &LoggerConfig{
		baseLoggerConfig: baseLoggerConfig{
			Level: InfoLevel,
			AppenderRefs: []*AppenderRef{
				{
					appender: &ConsoleAppender{
						BaseAppender: BaseAppender{
							Layout: &TextLayout{
								BaseLayout: BaseLayout{
									BufferSize:     500 * 1024,
									FileLineLength: 48,
								},
							},
						},
					},
				},
			},
		},
	},
}

// Tag is a struct representing a named logging tag.
// It holds a pointer to a Logger and a string identifier.
type Tag struct {
	v atomic.Pointer[Logger]
	s string
}

// GetName returns the name of the tag.
func (m *Tag) GetName() string {
	return m.s
}

// GetLogger returns the Logger associated with this tag.
// It uses atomic loading to ensure safe concurrent access.
func (m *Tag) GetLogger() *Logger {
	return m.v.Load()
}

// SetLogger sets or replaces the Logger associated with this tag.
// Uses atomic storing to ensure thread safety.
func (m *Tag) SetLogger(logger *Logger) {
	m.v.Store(logger)
}

// isValidTag checks whether the tag is valid according to the following rules:
// 1. The length must be between 3 and 36 characters.
// 2. Only lowercase letters (a-z), digits (0-9), and underscores (_) are allowed.
// 3. The tag can start with an underscore.
// 4. Underscores separate the tag into 1 to 4 non-empty segments.
// 5. No empty segments are allowed (i.e., no consecutive or trailing underscores).
func isValidTag(tag string) bool {
	if len(tag) < 3 || len(tag) > 36 {
		return false
	}
	for i := 0; i < len(tag); i++ {
		c := tag[i]
		if !(c >= 'a' && c <= 'z') && !(c >= '0' && c <= '9') && c != '_' {
			return false
		}
	}
	ss := strings.Split(strings.TrimPrefix(tag, "_"), "_")
	if len(ss) < 1 || len(ss) > 4 {
		return false
	}
	for _, s := range ss {
		if s == "" {
			return false
		}
	}
	return true
}

// GetTag creates or retrieves a Tag by name.
// If the tag does not exist, it is created and added to the global registry.
func GetTag(tag string) *Tag {
	if !isValidTag(tag) {
		panic("invalid tag name")
	}
	m, ok := tagMap[tag]
	if !ok {
		m = &Tag{s: tag}
		m.v.Store(initLogger)
		tagMap[tag] = m
	}
	return m
}

// GetAllTags returns all registered tags.
func GetAllTags() []string {
	return util.OrderedMapKeys(tagMap)
}
