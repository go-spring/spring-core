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
	"sync/atomic"
)

func init() {
	RegisterPlugin[AppenderRef]("AppenderRef", PluginTypeAppenderRef)
	RegisterPlugin[LoggerConfig]("Root", PluginTypeRoot)
	RegisterPlugin[AsyncLoggerConfig]("AsyncRoot", PluginTypeAsyncRoot)
	RegisterPlugin[LoggerConfig]("Logger", PluginTypeLogger)
	RegisterPlugin[AsyncLoggerConfig]("AsyncLogger", PluginTypeAsyncLogger)
}

type Logger struct {
	privateConfig
}

// privateConfig is the inner Logger.
type privateConfig interface {
	LifeCycle
	publish(e *Event)
	enableLevel(level Level) bool
}

// AppenderRef is a reference to an Appender.
type AppenderRef struct {
	Ref      string `PluginAttribute:"ref"`
	appender Appender
}

// baseLoggerConfig is the base of LoggerConfig and AsyncLoggerConfig.
type baseLoggerConfig struct {
	Name         string         `PluginAttribute:"name"`
	Level        Level          `PluginAttribute:"level"`
	Marker       string         `PluginAttribute:"marker,default="`
	AppenderRefs []*AppenderRef `PluginElement:"AppenderRef"`
}

// filter returns whether the event should be logged.
func (c *baseLoggerConfig) enableLevel(level Level) bool {
	return level >= c.Level
}

// callAppenders calls all the appenders inherited from the hierarchy circumventing.
func (c *baseLoggerConfig) callAppenders(e *Event) {
	for _, r := range c.AppenderRefs {
		r.appender.Append(e)
	}
}

// LoggerConfig publishes events synchronously.
type LoggerConfig struct {
	baseLoggerConfig
}

func (c *LoggerConfig) Start() error {
	return nil
}

func (c *LoggerConfig) publish(e *Event) {
	c.callAppenders(e)
}

func (c *LoggerConfig) Stop(ctx context.Context) {}

// AsyncLoggerConfig publishes events synchronously.
type AsyncLoggerConfig struct {
	baseLoggerConfig

	BufferSize int `PluginAttribute:"bufferSize,default=10000"`

	buf  chan *Event
	exit atomic.Bool
	wait chan struct{}
}

func (c *AsyncLoggerConfig) Start() error {
	c.buf = make(chan *Event, c.BufferSize)
	c.wait = make(chan struct{})
	go func() {
		for {
			select {
			case e := <-c.buf:
				if e != nil {
					c.callAppenders(e)
				}
			default:
				if c.exit.Load() {
					close(c.wait)
					return
				}
			}
		}
	}()
	return nil
}

// publish pushes events into the queue and these events will consumed by other
// goroutine, so the current goroutine will not be blocked.
func (c *AsyncLoggerConfig) publish(e *Event) {
	select {
	case c.buf <- e:
	default:
	}
}

func (c *AsyncLoggerConfig) Stop(ctx context.Context) {
	c.exit.Store(true)
	<-c.wait
}
