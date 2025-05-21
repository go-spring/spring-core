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
	"os"
)

func init() {
	RegisterPlugin[DiscardAppender]("Discard", PluginTypeAppender)
	RegisterPlugin[ConsoleAppender]("Console", PluginTypeAppender)
	RegisterPlugin[FileAppender]("File", PluginTypeAppender)
}

// Appender is an interface that defines components that handle log output.
type Appender interface {
	Lifecycle        // Appenders must be startable and stoppable
	Append(e *Event) // Handles writing a log event
}

var (
	_ Appender = (*DiscardAppender)(nil)
	_ Appender = (*ConsoleAppender)(nil)
	_ Appender = (*FileAppender)(nil)
)

// BaseAppender provides shared configuration and behavior for appenders.
type BaseAppender struct {
	Name   string `PluginAttribute:"name"` // Appender name from config
	Layout Layout `PluginElement:"Layout"` // Layout defines how logs are formatted
}

func (c *BaseAppender) Start() error    { return nil }
func (c *BaseAppender) Stop()           {}
func (c *BaseAppender) Append(e *Event) {}

// DiscardAppender ignores all log events (no output).
type DiscardAppender struct {
	BaseAppender
}

// ConsoleAppender writes formatted log events to stdout.
type ConsoleAppender struct {
	BaseAppender
}

// Append formats the event and writes it to standard output.
func (c *ConsoleAppender) Append(e *Event) {
	data, err := c.Layout.ToBytes(e)
	if err != nil {
		return
	}
	_, _ = os.Stdout.Write(data)
}

// FileAppender writes formatted log events to a specified file.
type FileAppender struct {
	BaseAppender
	FileName string `PluginAttribute:"fileName"`
	openFile *os.File
}

// Init opens or creates the log file for appending.
func (c *FileAppender) Init() error {
	f, err := os.OpenFile(c.FileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	c.openFile = f
	return nil
}

// Append formats the log event and writes it to the file.
func (c *FileAppender) Append(e *Event) {
	data, err := c.Layout.ToBytes(e)
	if err != nil {
		return
	}
	_, _ = c.openFile.Write(data)
}

// Stop closes the file.
func (c *FileAppender) Stop() {
	if c.openFile != nil {
		_ = c.openFile.Close()
	}
}
