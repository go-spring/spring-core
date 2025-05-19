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
	AppendEncoderBegin() error
	AppendEncoderEnd() error
	AppendObjectBegin() error
	AppendObjectEnd() error
	AppendArrayBegin() error
	AppendArrayEnd() error
	AppendKey(key string) error
	AppendBool(v bool) error
	AppendInt64(v int64) error
	AppendUint64(v uint64) error
	AppendFloat64(v float64) error
	AppendString(v string) error
	AppendReflect(v interface{}) error
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
func (enc *JSONEncoder) AppendEncoderBegin() error {
	enc.last = jsonTokenObjectBegin
	enc.buf.WriteByte('{')
	return nil
}

// AppendEncoderEnd writes the end of an encoder section (closes a JSON object).
func (enc *JSONEncoder) AppendEncoderEnd() error {
	enc.last = jsonTokenObjectEnd
	enc.buf.WriteByte('}')
	return nil
}

// AppendObjectBegin starts a new JSON object.
func (enc *JSONEncoder) AppendObjectBegin() error {
	enc.last = jsonTokenObjectBegin
	enc.buf.WriteByte('{')
	return nil
}

// AppendObjectEnd ends a JSON object.
func (enc *JSONEncoder) AppendObjectEnd() error {
	enc.last = jsonTokenObjectEnd
	enc.buf.WriteByte('}')
	return nil
}

// AppendArrayBegin starts a new JSON array.
func (enc *JSONEncoder) AppendArrayBegin() error {
	enc.last = jsonTokenArrayBegin
	enc.buf.WriteByte('[')
	return nil
}

// AppendArrayEnd ends a JSON array.
func (enc *JSONEncoder) AppendArrayEnd() error {
	enc.last = jsonTokenArrayEnd
	enc.buf.WriteByte(']')
	return nil
}

// appendSeparator inserts a comma if necessary before a key or value.
func (enc *JSONEncoder) appendSeparator(curr jsonToken) {
	switch curr {
	case jsonTokenKey:
		// Insert a comma between key-value pairs or elements
		if enc.last == jsonTokenObjectEnd || enc.last == jsonTokenArrayEnd || enc.last == jsonTokenValue {
			enc.buf.WriteByte(',')
		}
	case jsonTokenValue:
		if enc.last == jsonTokenValue {
			enc.buf.WriteByte(',')
		}
	default: // for linter
	}
}

// AppendKey writes a JSON key (as a string followed by a colon).
func (enc *JSONEncoder) AppendKey(key string) error {
	enc.appendSeparator(jsonTokenKey)
	enc.last = jsonTokenKey
	enc.buf.WriteByte('"')
	enc.safeAddString(key)
	enc.buf.WriteByte('"')
	enc.buf.WriteByte(':')
	return nil
}

// AppendBool writes a boolean value.
func (enc *JSONEncoder) AppendBool(v bool) error {
	enc.appendSeparator(jsonTokenValue)
	enc.last = jsonTokenValue
	enc.buf.WriteString(strconv.FormatBool(v))
	return nil
}

// AppendInt64 writes a signed 64-bit integer.
func (enc *JSONEncoder) AppendInt64(v int64) error {
	enc.appendSeparator(jsonTokenValue)
	enc.last = jsonTokenValue
	enc.buf.WriteString(strconv.FormatInt(v, 10))
	return nil
}

// AppendUint64 writes an unsigned 64-bit integer.
func (enc *JSONEncoder) AppendUint64(u uint64) error {
	enc.appendSeparator(jsonTokenValue)
	enc.last = jsonTokenValue
	enc.buf.WriteString(strconv.FormatUint(u, 10))
	return nil
}

// AppendFloat64 writes a floating-point number.
func (enc *JSONEncoder) AppendFloat64(v float64) error {
	enc.appendSeparator(jsonTokenValue)
	enc.last = jsonTokenValue
	enc.buf.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
	return nil
}

// AppendString writes a string value (properly escaped).
func (enc *JSONEncoder) AppendString(v string) error {
	enc.appendSeparator(jsonTokenValue)
	enc.last = jsonTokenValue
	enc.buf.WriteByte('"')
	enc.safeAddString(v)
	enc.buf.WriteByte('"')
	return nil
}

// AppendReflect marshals any Go value into JSON and appends it.
func (enc *JSONEncoder) AppendReflect(v interface{}) error {
	enc.appendSeparator(jsonTokenValue)
	enc.last = jsonTokenValue
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	enc.buf.Write(b)
	return nil
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
	init        bool          // Tracks if the first key-value has been written
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
func (enc *TextEncoder) AppendEncoderBegin() error {
	return nil
}

// AppendEncoderEnd is a no-op for TextEncoder (no special end token).
func (enc *TextEncoder) AppendEncoderEnd() error {
	return nil
}

// AppendObjectBegin delegates to JSONEncoder and increases JSON depth.
func (enc *TextEncoder) AppendObjectBegin() error {
	enc.jsonDepth++
	return enc.jsonEncoder.AppendObjectBegin()
}

// AppendObjectEnd delegates to JSONEncoder, decreases depth, and resets JSON encoder when top-level ends.
func (enc *TextEncoder) AppendObjectEnd() error {
	enc.jsonDepth--
	err := enc.jsonEncoder.AppendObjectEnd()
	if enc.jsonDepth == 0 {
		enc.jsonEncoder.Reset()
	}
	return err
}

// AppendArrayBegin delegates to JSONEncoder and increases JSON depth.
func (enc *TextEncoder) AppendArrayBegin() error {
	enc.jsonDepth++
	return enc.jsonEncoder.AppendArrayBegin()
}

// AppendArrayEnd delegates to JSONEncoder, decreases depth, and resets when array ends at top-level.
func (enc *TextEncoder) AppendArrayEnd() error {
	enc.jsonDepth--
	err := enc.jsonEncoder.AppendArrayEnd()
	if enc.jsonDepth == 0 {
		enc.jsonEncoder.Reset()
	}
	return err
}

// AppendKey writes a key for a key-value pair.
// If inside a JSON object, it delegates to JSONEncoder.
// Otherwise, it writes key= format and handles separator.
func (enc *TextEncoder) AppendKey(key string) error {
	if enc.jsonDepth > 0 {
		return enc.jsonEncoder.AppendKey(key)
	}
	if enc.init {
		enc.buf.WriteString(enc.separator)
	}
	enc.init = true
	enc.buf.WriteString(key)
	enc.buf.WriteByte('=')
	return nil
}

// AppendBool writes a boolean value, delegating to JSONEncoder if inside nested structure.
func (enc *TextEncoder) AppendBool(v bool) error {
	if enc.jsonDepth > 0 {
		return enc.jsonEncoder.AppendBool(v)
	}
	enc.buf.WriteString(strconv.FormatBool(v))
	return nil
}

// AppendInt64 writes an int64 value, or delegates to JSONEncoder if in nested structure.
func (enc *TextEncoder) AppendInt64(v int64) error {
	if enc.jsonDepth > 0 {
		return enc.jsonEncoder.AppendInt64(v)
	}
	enc.buf.WriteString(strconv.FormatInt(v, 10))
	return nil
}

// AppendUint64 writes a uint64 value, or delegates to JSONEncoder if in nested structure.
func (enc *TextEncoder) AppendUint64(v uint64) error {
	if enc.jsonDepth > 0 {
		return enc.jsonEncoder.AppendUint64(v)
	}
	enc.buf.WriteString(strconv.FormatUint(v, 10))
	return nil
}

// AppendFloat64 writes a float64 value, or delegates to JSONEncoder if in nested structure.
func (enc *TextEncoder) AppendFloat64(v float64) error {
	if enc.jsonDepth > 0 {
		return enc.jsonEncoder.AppendFloat64(v)
	}
	enc.buf.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
	return nil
}

// AppendString writes a raw string, or delegates to JSONEncoder if in nested structure.
func (enc *TextEncoder) AppendString(v string) error {
	if enc.jsonDepth > 0 {
		return enc.jsonEncoder.AppendString(v)
	}
	enc.buf.WriteString(v)
	return nil
}

// AppendReflect marshals and writes a value using JSON if not nested; otherwise, uses JSONEncoder.
func (enc *TextEncoder) AppendReflect(v interface{}) error {
	if enc.jsonDepth > 0 {
		return enc.jsonEncoder.AppendReflect(v)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	enc.buf.Write(b)
	return nil
}
