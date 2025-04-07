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

type PathType int

const (
	PathTypeKey   PathType = iota // PathTypeKey is map key like a/b in a[0][1].b
	PathTypeIndex                 // PathTypeIndex is array index like 0/1 in a[0][1].b
)

type Path struct {
	Type PathType
	Elem string
}

// JoinPath joins all path elements into a single path.
func JoinPath(path []Path) string {
	var s strings.Builder
	for i, p := range path {
		switch p.Type {
		case PathTypeKey:
			if i > 0 {
				s.WriteString(".")
			}
			s.WriteString(p.Elem)
		case PathTypeIndex:
			s.WriteString("[")
			s.WriteString(p.Elem)
			s.WriteString("]")
		}
	}
	return s.String()
}

// SplitPath splits key into individual path elements.
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

// appendKey appends a key to the path.
func appendKey(path []Path, s string) ([]Path, error) {
	_, err := strconv.ParseUint(s, 10, 64)
	if err == nil {
		return nil, errors.New("invalid key")
	}
	path = append(path, Path{PathTypeKey, s})
	return path, nil
}

// appendIndex appends an index to the path.
func appendIndex(path []Path, s string) ([]Path, error) {
	_, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return nil, errors.New("invalid key")
	}
	path = append(path, Path{PathTypeIndex, s})
	return path, nil
}
