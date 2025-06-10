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
	"encoding/json"
	"strconv"
	"unicode/utf8"
)

// Encoder is an interface that defines methods for appending structured data elements.
type Encoder interface {
	AppendEncoderBegin()
	AppendEncoderEnd()
	AppendObjectBegin()
	AppendObjectEnd()
	AppendArrayBegin()
	AppendArrayEnd()
	AppendKey(key string)
	AppendBool(v bool)
	AppendInt64(v int64)
	AppendUint64(v uint64)
	AppendFloat64(v float64)
	AppendString(v string)
	AppendReflect(v interface{})
}

var (
	_ Encoder = (*JSONEncoder)(nil)
	_ Encoder = (*TextEncoder)(nil)
)

// jsonToken represents the current state of the encoder while building a JSON structure.
type jsonToken int

const (
	jsonTokenUnknown jsonToken = iota
	jsonTokenObjectBegin
	jsonTokenObjectEnd
	jsonTokenArrayBegin
	jsonTokenArrayEnd
	jsonTokenKey
	jsonTokenValue
)

// JSONEncoder is a simple JSON encoder.
type JSONEncoder struct {
	buf  *bytes.Buffer // Buffer to write JSON output.
	last jsonToken     // The last token type written.
}

// NewJSONEncoder creates a new JSONEncoder.
func NewJSONEncoder(buf *bytes.Buffer) *JSONEncoder {
	return &JSONEncoder{
		buf:  buf,
		last: jsonTokenUnknown,
	}
}

// Reset resets the encoder's state.
func (enc *JSONEncoder) Reset() {
	enc.last = jsonTokenUnknown
}

// AppendEncoderBegin writes the start of an encoder section.
func (enc *JSONEncoder) AppendEncoderBegin() {
	enc.AppendObjectBegin()
}

// AppendEncoderEnd writes the end of an encoder section.
func (enc *JSONEncoder) AppendEncoderEnd() {
	enc.AppendObjectEnd()
}

// AppendObjectBegin writes the beginning of a JSON object.
func (enc *JSONEncoder) AppendObjectBegin() {
	enc.last = jsonTokenObjectBegin
	enc.buf.WriteByte('{')
}

// AppendObjectEnd writes the end of a JSON object.
func (enc *JSONEncoder) AppendObjectEnd() {
	enc.last = jsonTokenObjectEnd
	enc.buf.WriteByte('}')
}

// AppendArrayBegin writes the beginning of a JSON array.
func (enc *JSONEncoder) AppendArrayBegin() {
	enc.last = jsonTokenArrayBegin
	enc.buf.WriteByte('[')
}

// AppendArrayEnd writes the end of a JSON array.
func (enc *JSONEncoder) AppendArrayEnd() {
	enc.last = jsonTokenArrayEnd
	enc.buf.WriteByte(']')
}

// appendSeparator writes a comma if the previous token requires separation (e.g., between values).
func (enc *JSONEncoder) appendSeparator() {
	if enc.last == jsonTokenObjectEnd || enc.last == jsonTokenArrayEnd || enc.last == jsonTokenValue {
		enc.buf.WriteByte(',')
	}
}

// AppendKey writes a JSON key.
func (enc *JSONEncoder) AppendKey(key string) {
	enc.appendSeparator()
	enc.last = jsonTokenKey
	enc.buf.WriteByte('"')
	WriteLogString(enc.buf, key)
	enc.buf.WriteByte('"')
	enc.buf.WriteByte(':')
}

// AppendBool writes a boolean value.
func (enc *JSONEncoder) AppendBool(v bool) {
	enc.appendSeparator()
	enc.last = jsonTokenValue
	enc.buf.WriteString(strconv.FormatBool(v))
}

// AppendInt64 writes an int64 value.
func (enc *JSONEncoder) AppendInt64(v int64) {
	enc.appendSeparator()
	enc.last = jsonTokenValue
	enc.buf.WriteString(strconv.FormatInt(v, 10))
}

// AppendUint64 writes an uint64 value.
func (enc *JSONEncoder) AppendUint64(u uint64) {
	enc.appendSeparator()
	enc.last = jsonTokenValue
	enc.buf.WriteString(strconv.FormatUint(u, 10))
}

// AppendFloat64 writes a float64 value.
func (enc *JSONEncoder) AppendFloat64(v float64) {
	enc.appendSeparator()
	enc.last = jsonTokenValue
	enc.buf.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
}

// AppendString writes a string value with proper escaping.
func (enc *JSONEncoder) AppendString(v string) {
	enc.appendSeparator()
	enc.last = jsonTokenValue
	enc.buf.WriteByte('"')
	WriteLogString(enc.buf, v)
	enc.buf.WriteByte('"')
}

// AppendReflect marshals any Go value into JSON and appends it.
func (enc *JSONEncoder) AppendReflect(v interface{}) {
	enc.appendSeparator()
	enc.last = jsonTokenValue
	b, err := json.Marshal(v)
	if err != nil {
		enc.buf.WriteByte('"')
		WriteLogString(enc.buf, err.Error())
		enc.buf.WriteByte('"')
		return
	}
	enc.buf.Write(b)
}

// TextEncoder encodes key-value pairs in a plain text format,
// optionally using JSON when inside objects/arrays.
type TextEncoder struct {
	buf         *bytes.Buffer // Buffer to write the encoded output
	separator   string        // Separator used between top-level key-value pairs
	jsonEncoder *JSONEncoder  // Embedded JSON encoder for nested objects/arrays
	jsonDepth   int8          // Tracks depth of nested JSON structures
	wroteField  bool          // Tracks if the first key-value has been written
}

// NewTextEncoder creates a new TextEncoder, using the specified separator.
func NewTextEncoder(buf *bytes.Buffer, separator string) *TextEncoder {
	return &TextEncoder{
		buf:         buf,
		separator:   separator,
		jsonEncoder: &JSONEncoder{buf: buf},
	}
}

// AppendEncoderBegin writes the start of an encoder section.
func (enc *TextEncoder) AppendEncoderBegin() {}

// AppendEncoderEnd writes the end of an encoder section.
func (enc *TextEncoder) AppendEncoderEnd() {}

// AppendObjectBegin signals the start of a JSON object.
// Increments the depth and delegates to the JSON encoder.
func (enc *TextEncoder) AppendObjectBegin() {
	enc.jsonDepth++
	enc.jsonEncoder.AppendObjectBegin()
}

// AppendObjectEnd signals the end of a JSON object.
// Decrements the depth and resets the JSON encoder if back to top level.
func (enc *TextEncoder) AppendObjectEnd() {
	enc.jsonDepth--
	enc.jsonEncoder.AppendObjectEnd()
	if enc.jsonDepth == 0 {
		enc.jsonEncoder.Reset()
	}
}

// AppendArrayBegin signals the start of a JSON array.
// Increments the depth and delegates to the JSON encoder.
func (enc *TextEncoder) AppendArrayBegin() {
	enc.jsonDepth++
	enc.jsonEncoder.AppendArrayBegin()
}

// AppendArrayEnd signals the end of a JSON array.
// Decrements the depth and resets the JSON encoder if back to top level.
func (enc *TextEncoder) AppendArrayEnd() {
	enc.jsonDepth--
	enc.jsonEncoder.AppendArrayEnd()
	if enc.jsonDepth == 0 {
		enc.jsonEncoder.Reset()
	}
}

// AppendKey appends a key for a key-value pair.
// If inside a JSON structure, the key is handled by the JSON encoder.
// Otherwise, it's written directly with proper separator handling.
func (enc *TextEncoder) AppendKey(key string) {
	if enc.jsonDepth > 0 {
		enc.jsonEncoder.AppendKey(key)
		return
	}
	if enc.wroteField {
		enc.buf.WriteString(enc.separator)
	} else {
		enc.wroteField = true
	}
	WriteLogString(enc.buf, key)
	enc.buf.WriteByte('=')
}

// AppendBool appends a boolean value, using JSON encoder if nested.
func (enc *TextEncoder) AppendBool(v bool) {
	if enc.jsonDepth > 0 {
		enc.jsonEncoder.AppendBool(v)
		return
	}
	enc.buf.WriteString(strconv.FormatBool(v))
}

// AppendInt64 appends an int64 value, using JSON encoder if nested.
func (enc *TextEncoder) AppendInt64(v int64) {
	if enc.jsonDepth > 0 {
		enc.jsonEncoder.AppendInt64(v)
		return
	}
	enc.buf.WriteString(strconv.FormatInt(v, 10))
}

// AppendUint64 appends a uint64 value, using JSON encoder if nested.
func (enc *TextEncoder) AppendUint64(v uint64) {
	if enc.jsonDepth > 0 {
		enc.jsonEncoder.AppendUint64(v)
		return
	}
	enc.buf.WriteString(strconv.FormatUint(v, 10))
}

// AppendFloat64 appends a float64 value, using JSON encoder if nested.
func (enc *TextEncoder) AppendFloat64(v float64) {
	if enc.jsonDepth > 0 {
		enc.jsonEncoder.AppendFloat64(v)
		return
	}
	enc.buf.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
}

// AppendString appends a string value, using JSON encoder if nested.
func (enc *TextEncoder) AppendString(v string) {
	if enc.jsonDepth > 0 {
		enc.jsonEncoder.AppendString(v)
		return
	}
	WriteLogString(enc.buf, v)
}

// AppendReflect uses reflection to marshal any value as JSON.
// If nested, delegates to JSON encoder.
func (enc *TextEncoder) AppendReflect(v interface{}) {
	if enc.jsonDepth > 0 {
		enc.jsonEncoder.AppendReflect(v)
		return
	}
	b, err := json.Marshal(v)
	if err != nil {
		WriteLogString(enc.buf, err.Error())
		return
	}
	enc.buf.Write(b)
}

/************************************* string ********************************/

// WriteLogString escapes and writes a string according to JSON rules.
func WriteLogString(buf *bytes.Buffer, s string) {
	for i := 0; i < len(s); {
		// Try to add a single-byte (ASCII) character directly
		if tryAddRuneSelf(buf, s[i]) {
			i++
			continue
		}
		// Decode multi-byte UTF-8 character
		r, size := utf8.DecodeRuneInString(s[i:])
		// Handle invalid UTF-8 encoding
		if tryAddRuneError(buf, r, size) {
			i++
			continue
		}
		// Valid multi-byte rune; add as is
		buf.WriteString(s[i : i+size])
		i += size
	}
}

// tryAddRuneSelf handles ASCII characters and escapes control/quote characters.
func tryAddRuneSelf(buf *bytes.Buffer, b byte) bool {
	const _hex = "0123456789abcdef"
	if b >= utf8.RuneSelf {
		return false // not a single-byte rune
	}
	if 0x20 <= b && b != '\\' && b != '"' {
		buf.WriteByte(b)
		return true
	}
	// Handle escaping
	switch b {
	case '\\', '"':
		buf.WriteByte('\\')
		buf.WriteByte(b)
	case '\n':
		buf.WriteByte('\\')
		buf.WriteByte('n')
	case '\r':
		buf.WriteByte('\\')
		buf.WriteByte('r')
	case '\t':
		buf.WriteByte('\\')
		buf.WriteByte('t')
	default:
		// Encode bytes < 0x20, except for the escape sequences above.
		buf.WriteString(`\u00`)
		buf.WriteByte(_hex[b>>4])
		buf.WriteByte(_hex[b&0xF])
	}
	return true
}

// tryAddRuneError checks and escapes invalid UTF-8 runes.
func tryAddRuneError(buf *bytes.Buffer, r rune, size int) bool {
	if r == utf8.RuneError && size == 1 {
		buf.WriteString(`\ufffd`)
		return true
	}
	return false
}
