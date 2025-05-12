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
	"context"
	"os"
)

func init() {
	RegisterPlugin[DiscardAppender]("Discard", PluginTypeAppender)
	RegisterPlugin[ConsoleAppender]("Console", PluginTypeAppender)
	RegisterPlugin[FileAppender]("File", PluginTypeAppender)
}

// Appender is an Appender that writes log events.
type Appender interface {
	LifeCycle
	Append(e *Event)
}

var (
	_ Appender = (*DiscardAppender)(nil)
	_ Appender = (*ConsoleAppender)(nil)
	_ Appender = (*FileAppender)(nil)
)

// BaseAppender is an Appender that writes log events.
type BaseAppender struct {
	Name   string `PluginAttribute:"name"`
	Layout Layout `PluginElement:"Layout"`
}

func (c *BaseAppender) Start() error             { return nil }
func (c *BaseAppender) Stop(ctx context.Context) {}
func (c *BaseAppender) Append(e *Event)          {}

// DiscardAppender is an Appender that ignores log events.
type DiscardAppender struct {
	BaseAppender
}

// ConsoleAppender is an Appender that writing messages to os.Stdout.
type ConsoleAppender struct {
	BaseAppender
}

func (c *ConsoleAppender) Append(e *Event) {
	data, err := c.Layout.ToBytes(e)
	if err != nil {
		return
	}
	_, _ = os.Stdout.Write(data)
}

// FileAppender is an Appender writing messages to *os.File.
type FileAppender struct {
	BaseAppender
	FileName string `PluginAttribute:"fileName"`
	file     *os.File
}

func (c *FileAppender) Init() error {
	f, err := os.OpenFile(c.FileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	c.file = f
	return nil
}

func (c *FileAppender) Destroy() {
	_ = c.file.Close()
}

func (c *FileAppender) Append(e *Event) {
	data, err := c.Layout.ToBytes(e)
	if err != nil {
		return
	}
	_, _ = c.file.Write(data)
}
