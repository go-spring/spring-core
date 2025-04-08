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

	"github.com/go-spring/spring-core/util"
)

// treeNode represents a node in the hierarchical key structure.
// It contains the type of the path and a map of child nodes.
type treeNode struct {
	Type PathType
	Data map[string]*treeNode
}

// Storage is a key-value store that enforces hierarchical key structure validation.
// It uses a tree to manage key paths and a map to store the actual key-value pairs.
type Storage struct {
	root *treeNode
	data map[string]string
}

// NewStorage creates and initializes a new Storage instance.
func NewStorage() *Storage {
	return &Storage{
		data: make(map[string]string),
	}
}

// RawData returns the underlying key-value map.
// Note: This exposes internal state; use with caution.
func (s *Storage) RawData() map[string]string {
	return s.data
}

// SubKeys returns immediate child keys under the specified hierarchical key.
// It returns an error if the key format is invalid or if conflicts occur in the tree.
func (s *Storage) SubKeys(key string) (_ []string, err error) {
	var path []Path
	if key != "" {
		if path, err = SplitPath(key); err != nil {
			return nil, err
		}
	}

	if s.root == nil {
		return nil, nil
	}

	n := s.root
	for i, pathNode := range path {
		if n == nil || pathNode.Type != n.Type {
			return nil, fmt.Errorf("property conflict at path %s", JoinPath(path[:i+1]))
		}
		v, ok := n.Data[pathNode.Elem]
		if !ok {
			return nil, nil
		}
		n = v
	}

	if n == nil {
		return nil, fmt.Errorf("property conflict at path %s", key)
	}
	return util.OrderedMapKeys(n.Data), nil
}

// Has checks if a key exists in the storage.
// Returns false if the key format is invalid or the path doesn't exist.
func (s *Storage) Has(key string) bool {
	if key == "" || s.root == nil {
		return false
	}

	if _, ok := s.data[key]; ok {
		return true
	}

	path, err := SplitPath(key)
	if err != nil {
		return false
	}

	n := s.root
	for _, node := range path {
		if n == nil || node.Type != n.Type {
			return false
		}
		v, ok := n.Data[node.Elem]
		if !ok {
			return false
		}
		n = v
	}
	return true
}

// Set stores a key-value pair after validating the key's hierarchical structure.
// Returns an error for empty keys/values or path conflicts.
func (s *Storage) Set(key, val string) error {
	if key == "" {
		return errors.New("key is empty")
	}

	path, err := SplitPath(key)
	if err != nil {
		return err
	}

	// Initialize tree root if empty
	if s.root == nil {
		s.root = &treeNode{
			Type: path[0].Type,
			Data: make(map[string]*treeNode),
		}
	}

	n := s.root
	for i, pathNode := range path {
		if n == nil || pathNode.Type != n.Type {
			return fmt.Errorf("property conflict at path %s", JoinPath(path[:i+1]))
		}
		v, ok := n.Data[pathNode.Elem]
		if !ok {
			if i < len(path)-1 {
				v = &treeNode{
					Type: path[i+1].Type,
					Data: make(map[string]*treeNode),
				}
			}
			n.Data[pathNode.Elem] = v
		}
		n = v
	}
	if n != nil {
		return fmt.Errorf("property conflict at path %s", key)
	}

	s.data[key] = val
	return nil
}
