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

// jsonToken represents the type of the last JSON token written.
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

// JSONEncoder encodes Fields in json format.
type JSONEncoder struct {
	buf  *bytes.Buffer // Buffer to write JSON output
	last jsonToken     // The last token type written
}

// NewJSONEncoder creates a new JSONEncoder.
func NewJSONEncoder(buf *bytes.Buffer) *JSONEncoder {
	return &JSONEncoder{
		buf:  buf,
		last: jsonTokenUnknown,
	}
}

// Reset clears the encoder's state.
func (enc *JSONEncoder) Reset() {
	enc.last = jsonTokenUnknown
}

// AppendEncoderBegin writes the start of an encoder section, represented as a JSON object.
func (enc *JSONEncoder) AppendEncoderBegin() {
	enc.AppendObjectBegin()
}

// AppendEncoderEnd writes the end of an encoder section (closes a JSON object).
func (enc *JSONEncoder) AppendEncoderEnd() {
	enc.AppendObjectEnd()
}

// AppendObjectBegin starts a new JSON object.
func (enc *JSONEncoder) AppendObjectBegin() {
	enc.last = jsonTokenObjectBegin
	enc.buf.WriteByte('{')
}

// AppendObjectEnd ends a JSON object.
func (enc *JSONEncoder) AppendObjectEnd() {
	enc.last = jsonTokenObjectEnd
	enc.buf.WriteByte('}')
}

// AppendArrayBegin starts a new JSON array.
func (enc *JSONEncoder) AppendArrayBegin() {
	enc.last = jsonTokenArrayBegin
	enc.buf.WriteByte('[')
}

// AppendArrayEnd ends a JSON array.
func (enc *JSONEncoder) AppendArrayEnd() {
	enc.last = jsonTokenArrayEnd
	enc.buf.WriteByte(']')
}

// appendSeparator inserts a comma if necessary before a key or value.
func (enc *JSONEncoder) appendSeparator() {
	if enc.last == jsonTokenObjectEnd || enc.last == jsonTokenArrayEnd || enc.last == jsonTokenValue {
		enc.buf.WriteByte(',')
	}
}

// AppendKey writes a JSON key (as a string followed by a colon).
func (enc *JSONEncoder) AppendKey(key string) {
	enc.appendSeparator()
	enc.last = jsonTokenKey
	enc.buf.WriteByte('"')
	enc.safeAddString(key)
	enc.buf.WriteByte('"')
	enc.buf.WriteByte(':')
}

// AppendBool writes a boolean value.
func (enc *JSONEncoder) AppendBool(v bool) {
	enc.appendSeparator()
	enc.last = jsonTokenValue
	enc.buf.WriteString(strconv.FormatBool(v))
}

// AppendInt64 writes a signed 64-bit integer.
func (enc *JSONEncoder) AppendInt64(v int64) {
	enc.appendSeparator()
	enc.last = jsonTokenValue
	enc.buf.WriteString(strconv.FormatInt(v, 10))
}

// AppendUint64 writes an unsigned 64-bit integer.
func (enc *JSONEncoder) AppendUint64(u uint64) {
	enc.appendSeparator()
	enc.last = jsonTokenValue
	enc.buf.WriteString(strconv.FormatUint(u, 10))
}

// AppendFloat64 writes a floating-point number.
func (enc *JSONEncoder) AppendFloat64(v float64) {
	enc.appendSeparator()
	enc.last = jsonTokenValue
	enc.buf.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
}

// AppendString writes a string value (properly escaped).
func (enc *JSONEncoder) AppendString(v string) {
	enc.appendSeparator()
	enc.last = jsonTokenValue
	enc.buf.WriteByte('"')
	enc.safeAddString(v)
	enc.buf.WriteByte('"')
}

// AppendReflect marshals any Go value into JSON and appends it.
func (enc *JSONEncoder) AppendReflect(v interface{}) {
	enc.appendSeparator()
	enc.last = jsonTokenValue
	b, err := json.Marshal(v)
	if err != nil {
		b = []byte(err.Error())
	}
	enc.buf.Write(b)
}

// safeAddString escapes and writes a string according to JSON rules.
func (enc *JSONEncoder) safeAddString(s string) {
	for i := 0; i < len(s); {
		// Try to add a single-byte (ASCII) character directly
		if enc.tryAddRuneSelf(s[i]) {
			i++
			continue
		}
		// Decode multi-byte UTF-8 character
		r, size := utf8.DecodeRuneInString(s[i:])
		// Handle invalid UTF-8 encoding
		if enc.tryAddRuneError(r, size) {
			i++
			continue
		}
		// Valid multi-byte rune; add as is
		enc.buf.WriteString(s[i : i+size])
		i += size
	}
}

// tryAddRuneSelf handles ASCII characters and escapes control/quote characters.
func (enc *JSONEncoder) tryAddRuneSelf(b byte) bool {
	const _hex = "0123456789abcdef"
	if b >= utf8.RuneSelf {
		return false // not a single-byte rune
	}
	if 0x20 <= b && b != '\\' && b != '"' {
		enc.buf.WriteByte(b)
		return true
	}
	// Handle escaping
	switch b {
	case '\\', '"':
		enc.buf.WriteByte('\\')
		enc.buf.WriteByte(b)
	case '\n':
		enc.buf.WriteByte('\\')
		enc.buf.WriteByte('n')
	case '\r':
		enc.buf.WriteByte('\\')
		enc.buf.WriteByte('r')
	case '\t':
		enc.buf.WriteByte('\\')
		enc.buf.WriteByte('t')
	default:
		// Encode bytes < 0x20, except for the escape sequences above.
		enc.buf.WriteString(`\u00`)
		enc.buf.WriteByte(_hex[b>>4])
		enc.buf.WriteByte(_hex[b&0xF])
	}
	return true
}

// tryAddRuneError checks and escapes invalid UTF-8 runes.
func (enc *JSONEncoder) tryAddRuneError(r rune, size int) bool {
	if r == utf8.RuneError && size == 1 {
		enc.buf.WriteString(`\ufffd`)
		return true
	}
	return false
}

// TextEncoder encodes key-value pairs in a plain text format,
// optionally using JSON when inside objects/arrays.
type TextEncoder struct {
	buf         *bytes.Buffer // Buffer to write the encoded output
	separator   string        // Separator used between top-level key-value pairs
	jsonEncoder *JSONEncoder  // Embedded JSON encoder for nested objects/arrays
	jsonDepth   int8          // Tracks depth of nested JSON structures
	firstField  bool          // Tracks if the first key-value has been written
}

// NewTextEncoder creates a new TextEncoder writing to the given buffer, using the specified separator.
func NewTextEncoder(buf *bytes.Buffer, separator string) *TextEncoder {
	return &TextEncoder{
		buf:         buf,
		separator:   separator,
		jsonEncoder: &JSONEncoder{buf: buf},
	}
}

// AppendEncoderBegin is a no-op for TextEncoder (no special start token).
func (enc *TextEncoder) AppendEncoderBegin() {}

// AppendEncoderEnd is a no-op for TextEncoder (no special end token).
func (enc *TextEncoder) AppendEncoderEnd() {}

// AppendObjectBegin delegates to JSONEncoder and increases JSON depth.
func (enc *TextEncoder) AppendObjectBegin() {
	enc.jsonDepth++
	enc.jsonEncoder.AppendObjectBegin()
}

// AppendObjectEnd delegates to JSONEncoder, decreases depth, and resets JSON encoder when top-level ends.
func (enc *TextEncoder) AppendObjectEnd() {
	enc.jsonDepth--
	enc.jsonEncoder.AppendObjectEnd()
	if enc.jsonDepth == 0 {
		enc.jsonEncoder.Reset()
	}
}

// AppendArrayBegin delegates to JSONEncoder and increases JSON depth.
func (enc *TextEncoder) AppendArrayBegin() {
	enc.jsonDepth++
	enc.jsonEncoder.AppendArrayBegin()
}

// AppendArrayEnd delegates to JSONEncoder, decreases depth, and resets when array ends at top-level.
func (enc *TextEncoder) AppendArrayEnd() {
	enc.jsonDepth--
	enc.jsonEncoder.AppendArrayEnd()
	if enc.jsonDepth == 0 {
		enc.jsonEncoder.Reset()
	}
}

// AppendKey writes a key for a key-value pair.
// If inside a JSON object, it delegates to JSONEncoder.
// Otherwise, it writes key= format and handles separator.
func (enc *TextEncoder) AppendKey(key string) {
	if enc.jsonDepth > 0 {
		enc.jsonEncoder.AppendKey(key)
		return
	}
	if enc.firstField {
		enc.buf.WriteString(enc.separator)
	} else {
		enc.firstField = true
	}
	enc.buf.WriteString(key)
	enc.buf.WriteByte('=')
}

// AppendBool writes a boolean value, delegating to JSONEncoder if inside nested structure.
func (enc *TextEncoder) AppendBool(v bool) {
	if enc.jsonDepth > 0 {
		enc.jsonEncoder.AppendBool(v)
		return
	}
	enc.buf.WriteString(strconv.FormatBool(v))
}

// AppendInt64 writes an int64 value, or delegates to JSONEncoder if in nested structure.
func (enc *TextEncoder) AppendInt64(v int64) {
	if enc.jsonDepth > 0 {
		enc.jsonEncoder.AppendInt64(v)
		return
	}
	enc.buf.WriteString(strconv.FormatInt(v, 10))
}

// AppendUint64 writes a uint64 value, or delegates to JSONEncoder if in nested structure.
func (enc *TextEncoder) AppendUint64(v uint64) {
	if enc.jsonDepth > 0 {
		enc.jsonEncoder.AppendUint64(v)
		return
	}
	enc.buf.WriteString(strconv.FormatUint(v, 10))
}

// AppendFloat64 writes a float64 value, or delegates to JSONEncoder if in nested structure.
func (enc *TextEncoder) AppendFloat64(v float64) {
	if enc.jsonDepth > 0 {
		enc.jsonEncoder.AppendFloat64(v)
		return
	}
	enc.buf.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
}

// AppendString writes a raw string, or delegates to JSONEncoder if in nested structure.
func (enc *TextEncoder) AppendString(v string) {
	if enc.jsonDepth > 0 {
		enc.jsonEncoder.AppendString(v)
		return
	}
	enc.buf.WriteString(v)
}

// AppendReflect marshals and writes a value using JSON if not nested; otherwise, uses JSONEncoder.
func (enc *TextEncoder) AppendReflect(v interface{}) {
	if enc.jsonDepth > 0 {
		enc.jsonEncoder.AppendReflect(v)
		return
	}
	b, err := json.Marshal(v)
	if err != nil {
		b = []byte(err.Error())
	}
	enc.buf.Write(b)
}
