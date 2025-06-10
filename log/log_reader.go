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
	"strings"

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
	Text       string            // Text content of the XML element
}

// getChildren returns a slice of child nodes with a specific label.
func (node *Node) getChildren(label string) []*Node {
	var ret []*Node
	for _, c := range node.Children {
		if c.Label == label {
			ret = append(ret, c)
		}
	}
	return ret
}

// DumpNode prints the structure of a Node to a buffer.
func DumpNode(node *Node, indent int, buf *bytes.Buffer) {
	for i := 0; i < indent; i++ {
		buf.WriteString("\t")
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
	if node.Text != "" {
		buf.WriteString(" : ")
		buf.WriteString(node.Text)
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
		case xml.CharData:
			if text := strings.TrimSpace(string(t)); text != "" {
				curr := stack[len(stack)-1]
				curr.Text += text
			}
		case xml.EndElement:
			curr := stack[len(stack)-1]
			parent := stack[len(stack)-2]
			parent.Children = append(parent.Children, curr)
			stack = stack[:len(stack)-1]
		default: // for linter
		}
	}
	if len(stack[0].Children) == 0 {
		return nil, errors.New("invalid XML structure: missing root element")
	}
	return stack[0].Children[0], nil
}
