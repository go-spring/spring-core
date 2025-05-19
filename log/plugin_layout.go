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
	"strconv"
	"strings"
)

func init() {
	RegisterPlugin[TextLayout]("TextLayout", PluginTypeLayout)
	RegisterPlugin[JSONLayout]("JSONLayout", PluginTypeLayout)
}

// Layout is the interface that defines how a log event is converted to bytes.
type Layout interface {
	ToBytes(e *Event) ([]byte, error)
}

// TextLayout formats the log event as a human-readable text string.
type TextLayout struct{}

// ToBytes converts a log event to a formatted plain-text line.
func (c *TextLayout) ToBytes(e *Event) ([]byte, error) {
	const separator = "||"
	const maxLength = 48

	fileLine := e.File + ":" + strconv.Itoa(e.Line)
	if n := len(fileLine); n > maxLength-3 {
		fileLine = "..." + fileLine[n-maxLength:]
	}

	buf := GetBuffer()
	buf.WriteString("[")
	buf.WriteString(strings.ToUpper(e.Level.String()))
	buf.WriteString("][")
	buf.WriteString(e.Time.Format("2006-01-02T15:04:05.000"))
	buf.WriteString("][")
	buf.WriteString(fileLine)
	buf.WriteString("] ")
	buf.WriteString(e.Tag)
	buf.WriteString(separator)

	enc := NewTextEncoder(buf, separator)
	if err := enc.AppendEncoderBegin(); err != nil {
		return nil, err
	}
	if err := writeFields(enc, e.CtxFields); err != nil {
		return nil, err
	}
	if err := writeFields(enc, e.Fields); err != nil {
		return nil, err
	}
	if err := enc.AppendEncoderEnd(); err != nil {
		return nil, err
	}

	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

// JSONLayout formats the log event as a structured JSON object.
type JSONLayout struct{}

// ToBytes converts a log event to a JSON-formatted byte slice.
func (c *JSONLayout) ToBytes(e *Event) ([]byte, error) {
	const maxLength = 48
	fileLine := e.File + ":" + strconv.Itoa(e.Line)
	if n := len(fileLine); n > maxLength-3 {
		fileLine = "..." + fileLine[n-maxLength:]
	}

	fields := []Field{
		String("level", strings.ToLower(e.Level.String())),
		String("time", e.Time.Format("2006-01-02T15:04:05.000")),
		String("fileLine", fileLine),
		String("tag", e.Tag),
	}

	buf := GetBuffer()
	enc := NewJSONEncoder(buf)
	if err := enc.AppendEncoderBegin(); err != nil {
		return nil, err
	}
	if err := writeFields(enc, fields); err != nil {
		return nil, err
	}
	if err := writeFields(enc, e.CtxFields); err != nil {
		return nil, err
	}
	if err := writeFields(enc, e.Fields); err != nil {
		return nil, err
	}
	if err := enc.AppendEncoderEnd(); err != nil {
		return nil, err
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

// writeFields writes a slice of Field objects to the encoder.
func writeFields(enc Encoder, fields []Field) error {
	for _, f := range fields {
		if err := enc.AppendKey(f.Key); err != nil {
			return err
		}
		if err := f.Val.Encode(enc); err != nil {
			return err
		}
	}
	return nil
}
