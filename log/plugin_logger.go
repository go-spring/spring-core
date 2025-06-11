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
	"errors"
)

// OnDropEvent is a callback function that is called when an event is dropped.
var OnDropEvent func(logger string, e *Event)

func init() {
	RegisterPlugin[AppenderRef]("AppenderRef", PluginTypeAppenderRef)
	RegisterPlugin[LoggerConfig]("Root", PluginTypeRoot)
	RegisterPlugin[AsyncLoggerConfig]("AsyncRoot", PluginTypeAsyncRoot)
	RegisterPlugin[LoggerConfig]("Logger", PluginTypeLogger)
	RegisterPlugin[AsyncLoggerConfig]("AsyncLogger", PluginTypeAsyncLogger)
}

// Logger is the primary logging structure used to emit log events.
type Logger struct {
	privateConfig
}

// privateConfig is the interface implemented by all logger configs.
type privateConfig interface {
	Lifecycle                     // Start/Stop methods
	GetName() string              // Get the name of the logger
	Publish(e *Event)             // Logic for sending events to appenders
	EnableLevel(level Level) bool // Whether a log level is enabled
}

// AppenderRef represents a reference to an appender by name,
// which will be resolved and bound later.
type AppenderRef struct {
	Ref      string `PluginAttribute:"ref"`
	appender Appender
}

// baseLoggerConfig contains shared fields for all logger configurations.
type baseLoggerConfig struct {
	Name         string         `PluginAttribute:"name"`
	Level        Level          `PluginAttribute:"level"`
	Tags         string         `PluginAttribute:"tags,default="`
	AppenderRefs []*AppenderRef `PluginElement:"AppenderRef"`
}

// GetName returns the name of the logger.
func (c *baseLoggerConfig) GetName() string {
	return c.Name
}

// callAppenders sends the event to all configured appenders.
func (c *baseLoggerConfig) callAppenders(e *Event) {
	for _, r := range c.AppenderRefs {
		r.appender.Append(e)
	}
}

// EnableLevel returns true if the specified log level is enabled.
func (c *baseLoggerConfig) EnableLevel(level Level) bool {
	return level >= c.Level
}

// LoggerConfig is a synchronous logger configuration.
type LoggerConfig struct {
	baseLoggerConfig
}

func (c *LoggerConfig) Start() error { return nil }
func (c *LoggerConfig) Stop()        {}

// Publish sends the event directly to the appenders.
func (c *LoggerConfig) Publish(e *Event) {
	c.callAppenders(e)
	PutEvent(e)
}

// AsyncLoggerConfig is an asynchronous logger configuration.
// It buffers log events and processes them in a separate goroutine.
type AsyncLoggerConfig struct {
	baseLoggerConfig
	BufferSize int `PluginAttribute:"bufferSize,default=10000"`

	buf  chan *Event // Channel buffer for log events
	wait chan struct{}
}

// Start initializes the asynchronous logger and starts its worker goroutine.
func (c *AsyncLoggerConfig) Start() error {
	if c.BufferSize < 100 {
		return errors.New("bufferSize is too small")
	}
	c.buf = make(chan *Event, c.BufferSize)
	c.wait = make(chan struct{})

	// Launch a background goroutine to process events
	go func() {
		for e := range c.buf {
			c.callAppenders(e)
			PutEvent(e)
		}
		close(c.wait)
	}()
	return nil
}

// Publish places the event in the buffer if there's space; drops it otherwise.
func (c *AsyncLoggerConfig) Publish(e *Event) {
	select {
	case c.buf <- e:
	default:
		// Drop the event if the buffer is full
		if OnDropEvent != nil {
			OnDropEvent(c.Name, e)
		}
		// Return the event to the pool
		PutEvent(e)
	}
}

// Stop shuts down the asynchronous logger and waits for the worker goroutine to finish.
func (c *AsyncLoggerConfig) Stop() {
	close(c.buf)
	<-c.wait
}
