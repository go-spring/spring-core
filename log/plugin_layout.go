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
	"strings"

	"github.com/go-spring/spring-core/util/color"
)

func init() {
	RegisterPlugin[PatternLayout]("PatternLayout", PluginTypeLayout)
	RegisterPlugin[JSONLayout]("JSONLayout", PluginTypeLayout)
}

// Layout lays out an Event in []byte format.
type Layout interface {
	ToBytes(e *Event) ([]byte, error)
}

type ColorStyle int

const (
	ColorStyleNone = ColorStyle(iota)
	ColorStyleNormal
	ColorStyleBright
)

// ParseColorStyle parses `s` to a ColorStyle value.
func ParseColorStyle(s string) (ColorStyle, error) {
	switch strings.ToLower(s) {
	case "none":
		return ColorStyleNone, nil
	case "normal":
		return ColorStyleNormal, nil
	case "bright":
		return ColorStyleBright, nil
	default:
		return -1, fmt.Errorf("invalid color style '%s'", s)
	}
}

type FormatFunc func(e *Event) string

// A PatternLayout is a flexible layout configurable with pattern string.
type PatternLayout struct {
	ColorStyle ColorStyle `PluginAttribute:"colorStyle,default=none"`
	Pattern    string     `PluginAttribute:"pattern,default=[:level][:time][:fileline][:msg]"`
	steps      []FormatFunc
}

func (c *PatternLayout) Init() error {
	return c.parse(c.Pattern)
}

// ToBytes lays out an Event in []byte format.
func (c *PatternLayout) ToBytes(e *Event) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	for _, step := range c.steps {
		buf.WriteString(step(e))
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

func (c *PatternLayout) parse(pattern string) error {
	write := func(s string) FormatFunc {
		return func(e *Event) string {
			return s
		}
	}
	c.steps = append(c.steps, write("["))
	c.steps = append(c.steps, c.getLevel)
	c.steps = append(c.steps, write("]"))
	c.steps = append(c.steps, write("["))
	c.steps = append(c.steps, c.getTime)
	c.steps = append(c.steps, write("]"))
	c.steps = append(c.steps, write("["))
	c.steps = append(c.steps, c.getFileLine)
	c.steps = append(c.steps, write("]"))
	c.steps = append(c.steps, write(" "))
	c.steps = append(c.steps, c.getMsg)
	return nil
}

func (c *PatternLayout) getMsg(e *Event) string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(e.Marker)
	buf.WriteString("||")
	enc := NewFlatEncoder(buf, "||")
	err := enc.AppendEncoderBegin()
	if err != nil {
		return err.Error()
	}
	for _, f := range e.Fields {
		err = enc.AppendKey(f.Key)
		if err != nil {
			return err.Error()
		}
		err = f.Val.Encode(enc)
		if err != nil {
			return err.Error()
		}
	}
	err = enc.AppendEncoderEnd()
	if err != nil {
		return err.Error()
	}
	return buf.String()
}

func (c *PatternLayout) getLevel(e *Event) string {
	strLevel := strings.ToUpper(e.Level.String())
	switch c.ColorStyle {
	case ColorStyleNormal:
		if e.Level >= ErrorLevel {
			strLevel = color.Red.Sprint(strLevel)
		} else if e.Level == WarnLevel {
			strLevel = color.Yellow.Sprint(strLevel)
		} else if e.Level <= DebugLevel {
			strLevel = color.Green.Sprint(strLevel)
		}
	default:
	}
	return strLevel
}

func (c *PatternLayout) getTime(e *Event) string {
	return e.Time.Format("2006-01-02T15:04:05.000")
}

// Contract contracts `filename` and replace the excessive part using `...`.
func Contract(filename string, maxLength int) string {
	if n := len(filename); maxLength > 3 && n > maxLength-3 {
		return "..." + filename[n-maxLength+3:]
	}
	return filename
}

func (c *PatternLayout) getFileLine(e *Event) string {
	return Contract(fmt.Sprintf("%s:%d", e.File, e.Line), 48)
}

// A JSONLayout is a layout configurable with JSON encoding.
type JSONLayout struct{}

// ToBytes lays out an Event in []byte format.
func (c *JSONLayout) ToBytes(e *Event) ([]byte, error) {
	fields := []Field{
		String("level", strings.ToLower(e.Level.String())),
		String("time", e.Time.Format("2006-01-02T15:04:05.000")),
		String("fileLine", fmt.Sprintf("%s:%d", e.File, e.Line)),
		String("marker", e.Marker),
	}
	fields = append(fields, e.Fields...)

	buf := bytes.NewBuffer(nil)
	enc := NewJSONEncoder(buf)
	err := enc.AppendEncoderBegin()
	if err != nil {
		return nil, err
	}
	for _, f := range fields {
		err = enc.AppendKey(f.Key)
		if err != nil {
			return nil, err
		}
		err = f.Val.Encode(enc)
		if err != nil {
			return nil, err
		}
	}
	err = enc.AppendEncoderEnd()
	if err != nil {
		return nil, err
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}
