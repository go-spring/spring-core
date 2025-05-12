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
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

///////////////////////////////////////////////////////////////////////////////

var readers = map[string]Reader{}

func init() {
	RegisterReader(new(XMLReader), ".xml")
}

// Node 配置项节点。
type Node struct {
	Label      string
	Children   []*Node
	Attributes map[string]string
}

// getChild 获取子节点。
func (node *Node) getChild(label string) *Node {
	for _, c := range node.Children {
		if c.Label == label {
			return c
		}
	}
	return nil
}

// Reader 配置项解析器。
type Reader interface {
	Read(b []byte) (*Node, error)
}

// RegisterReader 注册配置项解析器。
func RegisterReader(r Reader, ext ...string) {
	for _, s := range ext {
		readers[s] = r
	}
}

// XMLReader XML配置项解析器。
type XMLReader struct{}

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

///////////////////////////////////////////////////////////////////////////////

var globalConfig struct {
	Loggers   map[string]*Logger
	Appenders map[string]Appender
}

// RefreshFile 加载日志配置文件。
func RefreshFile(fileName string) error {
	ext := filepath.Ext(fileName)
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	return RefreshReader(file, ext)
}

// RefreshReader 加载日志配置文件。
func RefreshReader(input io.Reader, ext string) error {

	var rootNode *Node
	{
		r, ok := readers[ext]
		if !ok {
			return fmt.Errorf("unsupported file type %s", ext)
		}
		data, err := io.ReadAll(input)
		if err != nil {
			return err
		}
		rootNode, err = r.Read(data)
		if err != nil {
			return err
		}
	}

	if rootNode.Label != "Configuration" {
		return errors.New("the Configuration root not found")
	}

	var (
		cRoot      *Logger
		cLoggers   = make(map[string]*Logger)
		cAppenders = make(map[string]Appender)
		cMarkers   = make(map[string]*Logger)
	)

	// 获取 Appenders 节点配置
	if node := rootNode.getChild("Appenders"); node != nil {
		for _, c := range node.Children {
			p, ok := plugins[c.Label]
			if !ok {
				return fmt.Errorf("plugin %s not found", c.Label)
			}
			name, ok := c.Attributes["name"]
			if !ok {
				return errors.New("attribute 'name' not found")
			}
			v, err := NewPlugin(p.Class, c)
			if err != nil {
				return err
			}
			cAppenders[name] = v.Interface().(Appender)
		}
	}

	// 获取 Loggers 节点配置
	if node := rootNode.getChild("Loggers"); node != nil {
		for _, c := range node.Children {

			// 判断是否是 Root 或 AsyncRoot 节点
			isRootLogger := c.Label == "Root" || c.Label == "AsyncRoot"
			if isRootLogger {
				if cRoot != nil {
					return errors.New("found more than one root loggers")
				}
				c.Attributes["name"] = "<<ROOT>>"
			}

			p, ok := plugins[c.Label]
			if !ok || p == nil {
				return fmt.Errorf("plugin %s not found", c.Label)
			}
			name, ok := c.Attributes["name"]
			if !ok {
				return errors.New("attribute 'name' not found")
			}
			v, err := NewPlugin(p.Class, c)
			if err != nil {
				return err
			}

			logger := &Logger{
				privateConfig: v.Interface().(privateConfig),
			}
			if isRootLogger {
				cRoot = logger
			}
			cLoggers[name] = logger

			// 根据 AppenderRef 初始化对应的 Appender
			var base *baseLoggerConfig
			switch config := v.Interface().(type) {
			case *LoggerConfig:
				base = &config.baseLoggerConfig
			case *AsyncLoggerConfig:
				base = &config.baseLoggerConfig
			}
			for _, r := range base.AppenderRefs {
				appender, ok := cAppenders[r.Ref]
				if !ok {
					return fmt.Errorf("appender %s not found", r.Ref)
				}
				r.appender = appender
			}

			// 记录所有的 Marker
			if isRootLogger {
				if base.Marker != "" {
					return fmt.Errorf("root logger can not have marker")
				}
				cMarkers[""] = logger
			} else {
				if base.Marker == "" {
					return fmt.Errorf("logger must have marker except root logger")
				}
				ss := strings.Split(base.Marker, ",")
				for _, s := range ss {
					cMarkers[s] = logger
				}
			}
		}
	}

	if cRoot == nil {
		return errors.New("found no root logger")
	}

	for _, a := range cAppenders {
		if err := a.Start(); err != nil {
			return err
		}
	}
	for _, l := range cLoggers {
		if err := l.Start(); err != nil {
			return err
		}
	}

	for s, marker := range markers {
		logger, ok := cMarkers[s]
		if !ok {
			marker.SetLogger(cRoot)
			continue
		}
		marker.SetLogger(logger)
	}

	// 停止之前的日志资源
	Stop(context.Background())

	globalConfig.Loggers = cLoggers
	globalConfig.Appenders = cAppenders
	return nil
}

// Stop 停止日志系统。
func Stop(ctx context.Context) {
	for _, l := range globalConfig.Loggers {
		l.Stop(ctx)
	}
	for _, l := range globalConfig.Appenders {
		l.Stop(ctx)
	}
}
