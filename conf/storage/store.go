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
Package storage provides hierarchical configuration storage and path parsing utilities.

Features:
- Storage manages key-value pairs with support for nested paths, subkey lookup, and conflict detection.
- Path represents structured access paths with support for parsing (SplitPath) and construction (JoinPath).
- Supports two path types:
  - Key (e.g., "user.name") for map access
  - Index (e.g., "[0]") for array access

- Maintains a tree structure (treeNode) for consistent and type-safe hierarchy management.

Use cases:
- Accessing values in JSON/YAML/TOML-like configs
- Managing nested config data (CRUD)
- Validating structure and detecting conflicts

Notes:
- Path syntax follows common config patterns (e.g., "users[0].profile.age")
- Type-safe path handling (keys vs. indices)
*/
package storage

import (
	"errors"
	"fmt"

	"github.com/go-spring/spring-core/util"
)

// treeNode represents a node in the hierarchical key path tree.
// Each node tracks the type of its path segment and its child nodes.
type treeNode struct {
	Type PathType
	Data map[string]*treeNode
}

// Storage stores hierarchical key-value pairs and tracks their structure using a tree.
// It supports nested paths and detects structural conflicts when paths differ in type.
type Storage struct {
	root *treeNode         // Root of the hierarchical key path tree
	data map[string]string // Flat key-value storage for exact key matches
}

// NewStorage creates and initializes a new Storage instance.
func NewStorage() *Storage {
	return &Storage{
		data: make(map[string]string),
	}
}

// RawData returns the underlying flat key-value map.
// Note: This exposes internal state; use with caution.
func (s *Storage) RawData() map[string]string {
	return s.data
}

// SubKeys returns the immediate subkeys under the given key path.
// It walks the tree structure and returns child elements if the path exists.
// Returns an error if there's a type conflict along the path.
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

// Has returns true if the given key exists in the storage,
// either as a direct value or as a valid path in the hierarchical tree structure.
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

// Set inserts a key-value pair into the storage.
// It also constructs or extends the corresponding hierarchical path in the tree.
// Returns an error if there is a type conflict or if the key is empty.
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
