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

type nodeType int

const (
	nodeTypeNil nodeType = iota
	nodeTypeValue
	nodeTypeMap
	nodeTypeArray
)

type treeNode struct {
	Type nodeType
	Data map[string]*treeNode
}

// Storage is a key-value store that verifies the format of the key.
type Storage struct {
	tree *treeNode
	data map[string]string
}

func NewStorage() *Storage {
	return &Storage{
		tree: &treeNode{
			Type: nodeTypeNil,
			Data: make(map[string]*treeNode),
		},
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
	tree := s.tree
	for i, pathNode := range path {
		m := tree.Data
		v, ok := m[pathNode.Elem]
		if !ok {
			return nil, nil
		}
		switch v.Type {
		case nodeTypeNil:
			return nil, nil
		case nodeTypeValue:
			return nil, fmt.Errorf("property '%s' is value", JoinPath(path[:i+1]))
		case nodeTypeArray, nodeTypeMap:
			tree = v
		default:
			return nil, fmt.Errorf("invalid node type %d", v.Type)
		}
	}
	m := tree.Data
	keys := util.OrderedMapKeys(m)
	return keys, nil
}

// Has returns whether the key exists.
func (s *Storage) Has(key string) bool {
	if key == "" {
		return false
	}
	if _, ok := s.data[key]; ok {
		return true
	}
	path, err := SplitPath(key)
	if err != nil {
		return false
	}
	tree := s.tree
	for i, node := range path {
		m := tree.Data
		switch tree.Type {
		case nodeTypeArray:
			if node.Type != PathTypeIndex {
				return false
			}
		case nodeTypeMap:
			if node.Type != PathTypeKey {
				return false
			}
		default: // for linter
		}
		v, ok := m[node.Elem]
		if !ok {
			return false
		}
		if v.Type == nodeTypeNil || v.Type == nodeTypeValue {
			return i == len(path)-1
		}
		tree = v
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
	if err := s.checkNode(key); err != nil {
		return err
	}
	s.data[key] = val
	return nil
}

func (s *Storage) checkNode(key string) error {
	path, err := SplitPath(key)
	if err != nil {
		return err
	}
	tree := s.tree
	for i, pathNode := range path {
		if tree.Type == nodeTypeMap {
			if pathNode.Type != PathTypeKey {
				return fmt.Errorf("property '%s' is a map but '%s' wants other type", JoinPath(path[:i]), key)
			}
		}
		m := tree.Data
		v, ok := m[pathNode.Elem]
		if v != nil && v.Type == nodeTypeNil {
			delete(s.data, JoinPath(path[:i+1]))
		}
		if !ok || v.Type == nodeTypeNil {
			if i < len(path)-1 {
				n := &treeNode{
					Data: make(map[string]*treeNode),
				}
				if path[i+1].Type == PathTypeIndex {
					n.Type = nodeTypeArray
				} else {
					n.Type = nodeTypeMap
				}
				m[pathNode.Elem] = n
				tree = n
				continue
			}
			tree = &treeNode{Type: nodeTypeValue}
			m[pathNode.Elem] = tree
			break // break for 100% test
		}
		switch v.Type {
		case nodeTypeMap:
			if i < len(path)-1 {
				tree = v
				continue
			}
			return fmt.Errorf("property '%s' is a map but '%s' wants other type", JoinPath(path[:i+1]), key)
		case nodeTypeArray:
			if pathNode.Type != PathTypeIndex {
				if i < len(path)-1 && path[i+1].Type != PathTypeIndex {
					return fmt.Errorf("property '%s' is an array but '%s' wants other type", JoinPath(path[:i+1]), key)
				}
			}
			if i < len(path)-1 {
				tree = v
				continue
			}
			return fmt.Errorf("property '%s' is an array but '%s' wants other type", JoinPath(path[:i+1]), key)
		case nodeTypeValue:
			if i == len(path)-1 {
				return nil
			}
			return fmt.Errorf("property '%s' is a value but '%s' wants other type", JoinPath(path[:i+1]), key)
		default:
			return fmt.Errorf("invalid node type %d", v.Type)
		}
	}
	return nil
}
