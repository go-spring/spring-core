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

package storage

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// PathType represents the type of a path segment.
type PathType int

const (
	PathTypeKey   PathType = iota // PathTypeKey indicates a named key in a map.
	PathTypeIndex                 // PathTypeIndex indicates a numeric index in a list.
)

// Path represents a segment of a hierarchical path.
// Each segment is either a key (e.g., "user") or an index (e.g., "0").
type Path struct {
	Type PathType // Type determines whether the segment is a key or index.
	Elem string   // Elem holds the actual key or index value as a string.
}

// JoinPath constructs a string representation from a slice of Path segments.
// Keys are joined with '.', and indices are represented as '[i]'.
func JoinPath(path []Path) string {
	var sb strings.Builder
	for i, p := range path {
		switch p.Type {
		case PathTypeKey:
			if i > 0 {
				sb.WriteString(".")
			}
			sb.WriteString(p.Elem)
		case PathTypeIndex:
			sb.WriteString("[")
			sb.WriteString(p.Elem)
			sb.WriteString("]")
		}
	}
	return sb.String()
}

// SplitPath parses a string path into a slice of Path segments.
// It supports keys separated by '.' and indices enclosed in brackets (e.g., "users[0].name").
func SplitPath(key string) (_ []Path, err error) {
	if key == "" {
		return nil, fmt.Errorf("invalid key '%s'", key)
	}
	var (
		path        []Path
		lastPos     int
		lastChar    int32
		openBracket bool
	)
	for i, c := range key {
		switch c {
		case ' ':
			return nil, fmt.Errorf("invalid key '%s'", key)
		case '.':
			if openBracket || lastChar == '.' {
				return nil, fmt.Errorf("invalid key '%s'", key)
			}
			if lastChar != ']' {
				path, err = appendKey(path, key[lastPos:i])
				if err != nil {
					return nil, fmt.Errorf("invalid key '%s'", key)
				}
			}
			lastPos = i + 1
			lastChar = c
		case '[':
			if openBracket || lastChar == '.' {
				return nil, fmt.Errorf("invalid key '%s'", key)
			}
			if i > 0 && lastChar != ']' {
				path, err = appendKey(path, key[lastPos:i])
				if err != nil {
					return nil, fmt.Errorf("invalid key '%s'", key)
				}
			}
			openBracket = true
			lastPos = i + 1
			lastChar = c
		case ']':
			if !openBracket {
				return nil, fmt.Errorf("invalid key '%s'", key)
			}
			path, err = appendIndex(path, key[lastPos:i])
			if err != nil {
				return nil, fmt.Errorf("invalid key '%s'", key)
			}
			openBracket = false
			lastPos = i + 1
			lastChar = c
		default:
			if lastChar == ']' {
				return nil, fmt.Errorf("invalid key '%s'", key)
			}
			lastChar = c
		}
	}
	if openBracket || lastChar == '.' {
		return nil, fmt.Errorf("invalid key '%s'", key)
	}
	if lastChar != ']' {
		path, err = appendKey(path, key[lastPos:])
		if err != nil {
			return nil, fmt.Errorf("invalid key '%s'", key)
		}
	}
	return path, nil
}

// appendKey appends a key segment to the path.
func appendKey(path []Path, s string) ([]Path, error) {
	_, err := strconv.ParseUint(s, 10, 64)
	if err == nil {
		return nil, errors.New("invalid key")
	}
	path = append(path, Path{PathTypeKey, s})
	return path, nil
}

// appendIndex appends an index segment to the path.
func appendIndex(path []Path, s string) ([]Path, error) {
	_, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return nil, errors.New("invalid key")
	}
	path = append(path, Path{PathTypeIndex, s})
	return path, nil
}
