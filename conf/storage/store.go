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

type treeNode struct {
	Type PathType
	Data map[string]*treeNode
}

// Storage is a key-value store that verifies the format of the key.
type Storage struct {
	tree *treeNode
	data map[string]string
}

func NewStorage() *Storage {
	return &Storage{
		data: make(map[string]string),
	}
}

// RawData returns the raw data of the storage.
func (s *Storage) RawData() map[string]string {
	return s.data
}

// Data returns the copied data of the storage.
func (s *Storage) Data() map[string]string {
	m := make(map[string]string)
	for k, v := range s.data {
		m[k] = v
	}
	return m
}

// Keys returns the sorted keys of the storage.
func (s *Storage) Keys() []string {
	return util.OrderedMapKeys(s.data)
}

// SubKeys returns the sorted sub keys of the key.
func (s *Storage) SubKeys(key string) (_ []string, err error) {
	var path []Path
	if key != "" {
		if path, err = SplitPath(key); err != nil {
			return nil, err
		}
	}
	if s.tree == nil {
		return nil, nil
	}
	n := s.tree
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

// Has returns whether the key exists.
func (s *Storage) Has(key string) bool {
	if key == "" || s.tree == nil {
		return false
	}
	if _, ok := s.data[key]; ok {
		return true
	}
	path, err := SplitPath(key)
	if err != nil {
		return false
	}
	n := s.tree
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

// Get returns the value of the key, and false if the key does not exist.
func (s *Storage) Get(key string) (string, bool) {
	val, ok := s.data[key]
	return val, ok
}

// Set stores the value of the key.
func (s *Storage) Set(key, val string) error {
	if key == "" || val == "" {
		return errors.New("key or value is empty")
	}
	path, err := SplitPath(key)
	if err != nil {
		return err
	}
	if s.tree == nil {
		s.tree = &treeNode{
			Type: path[0].Type,
			Data: make(map[string]*treeNode),
		}
	}
	n := s.tree
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
