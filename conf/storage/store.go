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
	node nodeType
	data interface{}
}

// Storage is a key-value store that verifies the format of the key.
type Storage struct {
	tree *treeNode
	data map[string]string
}

func NewStorage() *Storage {
	return &Storage{
		tree: &treeNode{
			node: nodeTypeNil,
			data: make(map[string]*treeNode),
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
		m := tree.data.(map[string]*treeNode)
		v, ok := m[pathNode.Elem]
		if !ok {
			return nil, nil
		}
		switch v.node {
		case nodeTypeNil:
			return nil, nil
		case nodeTypeValue:
			return nil, fmt.Errorf("property '%s' is value", JoinPath(path[:i+1]))
		case nodeTypeArray, nodeTypeMap:
			tree = v
		default:
			return nil, fmt.Errorf("invalid node type %d", v.node)
		}
	}
	m := tree.data.(map[string]*treeNode)
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
		m := tree.data.(map[string]*treeNode)
		switch tree.node {
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
		if v.node == nodeTypeNil || v.node == nodeTypeValue {
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
	if key == "" {
		return fmt.Errorf("key is empty")
	}
	tree, err := s.merge(key, val)
	if err != nil {
		return err
	}
	switch tree.node {
	case nodeTypeNil, nodeTypeValue:
		s.data[key] = val
	default:
		return fmt.Errorf("invalid node type %d", tree.node)
	}
	return nil
}

func (s *Storage) merge(key, val string) (*treeNode, error) {
	path, err := SplitPath(key)
	if err != nil {
		return nil, err
	}
	tree := s.tree
	for i, pathNode := range path {
		if tree.node == nodeTypeMap {
			if pathNode.Type != PathTypeKey {
				return nil, fmt.Errorf("property '%s' is a map but '%s' wants other type", JoinPath(path[:i]), key)
			}
		}
		m := tree.data.(map[string]*treeNode)
		v, ok := m[pathNode.Elem]
		if v != nil && v.node == nodeTypeNil {
			delete(s.data, JoinPath(path[:i+1]))
		}
		if !ok || v.node == nodeTypeNil {
			if i < len(path)-1 {
				n := &treeNode{
					data: make(map[string]*treeNode),
				}
				if path[i+1].Type == PathTypeIndex {
					n.node = nodeTypeArray
				} else {
					n.node = nodeTypeMap
				}
				m[pathNode.Elem] = n
				tree = n
				continue
			}
			if val == "" {
				tree = &treeNode{node: nodeTypeNil}
			} else {
				tree = &treeNode{node: nodeTypeValue}
			}
			m[pathNode.Elem] = tree
			break // break for 100% test
		}
		switch v.node {
		case nodeTypeMap:
			if i < len(path)-1 {
				tree = v
				continue
			}
			if val == "" {
				return v, nil
			}
			return nil, fmt.Errorf("property '%s' is a map but '%s' wants other type", JoinPath(path[:i+1]), key)
		case nodeTypeArray:
			if pathNode.Type != PathTypeIndex {
				if i < len(path)-1 && path[i+1].Type != PathTypeIndex {
					return nil, fmt.Errorf("property '%s' is an array but '%s' wants other type", JoinPath(path[:i+1]), key)
				}
			}
			if i < len(path)-1 {
				tree = v
				continue
			}
			if val == "" {
				return v, nil
			}
			return nil, fmt.Errorf("property '%s' is an array but '%s' wants other type", JoinPath(path[:i+1]), key)
		case nodeTypeValue:
			if i == len(path)-1 {
				return v, nil
			}
			return nil, fmt.Errorf("property '%s' is a value but '%s' wants other type", JoinPath(path[:i+1]), key)
		default:
			return nil, fmt.Errorf("invalid node type %d", v.node)
		}
	}
	return tree, nil
}
