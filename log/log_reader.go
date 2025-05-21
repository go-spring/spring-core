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
	"bytes"
	"encoding/xml"
	"errors"
	"io"

	"github.com/go-spring/spring-core/util"
)

var readers = map[string]Reader{}

func init() {
	RegisterReader(new(XMLReader), ".xml")
}

// Node represents a parsed XML element with a label (tag name), child nodes,
// and a map of attributes.
type Node struct {
	Label      string            // Tag name of the XML element
	Children   []*Node           // Child elements (nested tags)
	Attributes map[string]string // Attributes of the XML element
}

// getChild returns the first child node with the specified label.
// Returns nil if no matching child is found.
func (node *Node) getChild(label string) *Node {
	for _, c := range node.Children {
		if c.Label == label {
			return c
		}
	}
	return nil
}

func DumpNode(node *Node, indent int, buf *bytes.Buffer) {
	for i := 0; i < indent; i++ {
		buf.WriteString("    ")
	}
	buf.WriteString(node.Label)
	if len(node.Attributes) > 0 {
		buf.WriteString(" {")
		for i, k := range util.OrderedMapKeys(node.Attributes) {
			if i > 0 {
				buf.WriteString(" ")
			}
			buf.WriteString(k)
			buf.WriteString("=")
			buf.WriteString(node.Attributes[k])
		}
		buf.WriteString("}")
	}
	for _, c := range node.Children {
		buf.WriteString("\n")
		DumpNode(c, indent+1, buf)
	}
}

// Reader is an interface for reading and parsing data into a Node structure.
type Reader interface {
	Read(b []byte) (*Node, error)
}

// RegisterReader registers a Reader for one or more file extensions.
// This allows dynamic selection of parsers based on file type.
func RegisterReader(r Reader, ext ...string) {
	for _, s := range ext {
		readers[s] = r
	}
}

// XMLReader is an implementation of the Reader interface that parses XML data.
type XMLReader struct{}

// Read parses XML bytes into a tree of Nodes.
// It uses a stack to track the current position in the XML hierarchy.
func (r *XMLReader) Read(b []byte) (*Node, error) {
	stack := []*Node{{Label: "<<STACK>>"}}
	d := xml.NewDecoder(bytes.NewReader(b))
	for {
		token, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := token.(type) {
		case xml.StartElement:
			curr := &Node{
				Label:      t.Name.Local,
				Attributes: make(map[string]string),
			}
			for _, attr := range t.Attr {
				curr.Attributes[attr.Name.Local] = attr.Value
			}
			stack = append(stack, curr)
		case xml.EndElement:
			curr := stack[len(stack)-1]
			parent := stack[len(stack)-2]
			parent.Children = append(parent.Children, curr)
			stack = stack[:len(stack)-1]
		default:
		}
	}
	if len(stack[0].Children) == 0 {
		return nil, errors.New("error xml config file")
	}
	return stack[0].Children[0], nil
}
