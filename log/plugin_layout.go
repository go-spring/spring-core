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
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

var bytesSizeTable = map[string]int64{
	"B":  1,
	"KB": 1024,
	"MB": 1024 * 1024,
}

func init() {
	RegisterConverter[HumanizeBytes](ParseHumanizeBytes)
	RegisterPlugin[*TextLayout]("TextLayout", PluginTypeLayout)
	RegisterPlugin[*JSONLayout]("JSONLayout", PluginTypeLayout)
}

type HumanizeBytes int

// ParseHumanizeBytes converts a human-readable byte string to an integer.
func ParseHumanizeBytes(s string) (HumanizeBytes, error) {
	lastDigit := 0
	for _, r := range s {
		if !unicode.IsDigit(r) {
			break
		}
		lastDigit++
	}
	num := s[:lastDigit]
	f, err := strconv.ParseInt(num, 10, 64)
	if err != nil {
		return 0, err
	}
	extra := strings.ToUpper(strings.TrimSpace(s[lastDigit:]))
	if m, ok := bytesSizeTable[extra]; ok {
		f *= m
		return HumanizeBytes(f), nil
	}
	return 0, fmt.Errorf("unhandled size name: %q", extra)
}

// Layout is the interface that defines how a log event is converted to bytes.
type Layout interface {
	Lifecycle
	ToBytes(e *Event) []byte
}

// BaseLayout is the base class for Layout.
type BaseLayout struct {
	BufferSize     HumanizeBytes `PluginAttribute:"bufferSize,default=1MB"`
	FileLineLength int           `PluginAttribute:"fileLineLength,default=48"`

	buffer *bytes.Buffer
}

func (c *BaseLayout) Start() error { return nil }
func (c *BaseLayout) Stop()        {}

// GetBuffer returns a buffer that can be used to format the log event.
func (c *BaseLayout) GetBuffer() *bytes.Buffer {
	if c.buffer == nil {
		c.buffer = &bytes.Buffer{}
		c.buffer.Grow(int(c.BufferSize))
	}
	return c.buffer
}

// PutBuffer puts a buffer back into the pool.
func (c *BaseLayout) PutBuffer(buf *bytes.Buffer) {
	if buf.Cap() > int(c.BufferSize) {
		c.buffer = nil
		return
	}
	c.buffer = buf
	c.buffer.Reset()
}

// GetFileLine returns the file name and line number of the log event.
func (c *BaseLayout) GetFileLine(e *Event) string {
	fileLine := e.File + ":" + strconv.Itoa(e.Line)
	if n := len(fileLine); n > c.FileLineLength-3 {
		fileLine = "..." + fileLine[n-c.FileLineLength:]
	}
	return fileLine
}

// TextLayout formats the log event as a human-readable text string.
type TextLayout struct {
	BaseLayout
}

// ToBytes converts a log event to a formatted plain-text line.
func (c *TextLayout) ToBytes(e *Event) []byte {
	const separator = "||"

	buf := c.GetBuffer()
	defer c.PutBuffer(buf)

	buf.WriteString("[")
	buf.WriteString(strings.ToUpper(e.Level.String()))
	buf.WriteString("][")
	buf.WriteString(e.Time.Format("2006-01-02T15:04:05.000"))
	buf.WriteString("][")
	buf.WriteString(c.GetFileLine(e))
	buf.WriteString("] ")
	buf.WriteString(e.Tag)
	buf.WriteString(separator)

	if e.CtxString != "" {
		buf.WriteString(e.CtxString)
		buf.WriteString(separator)
	}

	enc := NewTextEncoder(buf, separator)
	enc.AppendEncoderBegin()
	WriteFields(enc, e.CtxFields)
	WriteFields(enc, e.Fields)
	enc.AppendEncoderEnd()

	buf.WriteByte('\n')
	return buf.Bytes()
}

// JSONLayout formats the log event as a structured JSON object.
type JSONLayout struct {
	BaseLayout
}

// ToBytes converts a log event to a JSON-formatted byte slice.
func (c *JSONLayout) ToBytes(e *Event) []byte {
	buf := c.GetBuffer()
	defer c.PutBuffer(buf)

	fields := make([]Field, 0, 5)
	fields = append(fields, String("level", strings.ToLower(e.Level.String())))
	fields = append(fields, String("time", e.Time.Format("2006-01-02T15:04:05.000")))
	fields = append(fields, String("fileLine", c.GetFileLine(e)))
	fields = append(fields, String("tag", e.Tag))

	if e.CtxString != "" {
		fields = append(fields, String("ctxString", e.CtxString))
	}

	enc := NewJSONEncoder(buf)
	enc.AppendEncoderBegin()
	WriteFields(enc, fields)
	WriteFields(enc, e.CtxFields)
	WriteFields(enc, e.Fields)
	enc.AppendEncoderEnd()

	buf.WriteByte('\n')
	return buf.Bytes()
}
