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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"

	"github.com/go-spring/spring-core/util"
	"github.com/go-spring/spring-core/util/errutil"
)

var initOnce atomic.Bool

// RefreshFile loads a logging configuration from a file by its name.
func RefreshFile(fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	ext := filepath.Ext(fileName)
	return RefreshReader(file, ext)
}

// RefreshReader reads the configuration from an io.Reader using the reader for the given extension.
func RefreshReader(input io.Reader, ext string) error {
	if !initOnce.CompareAndSwap(false, true) {
		return errors.New("RefreshReader: log refresh already done")
	}

	var rootNode *Node
	{
		r, ok := readers[ext]
		if !ok {
			return fmt.Errorf("RefreshReader: unsupported file type %s", ext)
		}
		data, err := io.ReadAll(input)
		if err != nil {
			return err
		}
		rootNode, err = r.Read(data)
		if err != nil {
			return err
		}
	}

	if rootNode.Label != "Configuration" {
		return errors.New("RefreshReader: Configuration root not found")
	}

	var (
		cRoot      *Logger
		cLoggers   = make(map[string]*Logger)
		cAppenders = make(map[string]Appender)
		cTags      = make(map[string]*Logger)
		properties = make(map[string]string)
	)

	// Parse <Properties> section
	nodes := rootNode.getChildren("Properties")
	if len(nodes) > 1 {
		return errors.New("RefreshReader: Properties section must be unique")
	}
	if len(nodes) == 1 {
		for _, c := range nodes[0].Children {
			if c.Label != "Property" {
				continue
			}
			name, ok := c.Attributes["name"]
			if !ok {
				return fmt.Errorf("RefreshReader: attribute 'name' not found for node %s", c.Label)
			}
			properties[name] = c.Text
		}
	}

	// Parse <Appenders> section
	nodes = rootNode.getChildren("Appenders")
	if len(nodes) == 0 {
		return errors.New("RefreshReader: Appenders section not found")
	}
	if len(nodes) > 1 {
		return errors.New("RefreshReader: Appenders section must be unique")
	}
	for _, c := range nodes[0].Children {
		p, ok := plugins[c.Label]
		if !ok {
			return fmt.Errorf("RefreshReader: plugin %s not found", c.Label)
		}
		name, ok := c.Attributes["name"]
		if !ok {
			return errors.New("RefreshReader: attribute 'name' not found")
		}
		v, err := NewPlugin(p.Class, c, properties)
		if err != nil {
			return err
		}
		cAppenders[name] = v.Interface().(Appender)
	}

	// Parse <Loggers> section
	nodes = rootNode.getChildren("Loggers")
	if len(nodes) == 0 {
		return errors.New("RefreshReader: Loggers section not found")
	}
	if len(nodes) > 1 {
		return errors.New("RefreshReader: Loggers section must be unique")
	}
	for _, c := range nodes[0].Children {
		isRootLogger := c.Label == "Root" || c.Label == "AsyncRoot"
		if isRootLogger {
			if cRoot != nil {
				return errors.New("RefreshReader: found more than one root loggers")
			}
			c.Attributes["name"] = "::root::"
		}

		p, ok := plugins[c.Label]
		if !ok || p == nil {
			return fmt.Errorf("RefreshReader: plugin %s not found", c.Label)
		}
		name, ok := c.Attributes["name"]
		if !ok {
			return fmt.Errorf("RefreshReader: attribute 'name' not found for node %s", c.Label)
		}
		v, err := NewPlugin(p.Class, c, properties)
		if err != nil {
			return err
		}

		logger := &Logger{v.Interface().(privateConfig)}
		if isRootLogger {
			cRoot = logger
		}
		cLoggers[name] = logger

		var base *baseLoggerConfig
		switch config := v.Interface().(type) {
		case *LoggerConfig:
			base = &config.baseLoggerConfig
		case *AsyncLoggerConfig:
			base = &config.baseLoggerConfig
		}

		for _, r := range base.AppenderRefs {
			appender, ok := cAppenders[r.Ref]
			if !ok {
				return fmt.Errorf("RefreshReader: appender %s not found", r.Ref)
			}
			r.appender = appender
		}

		if isRootLogger {
			if base.Tags != "" {
				return fmt.Errorf("RefreshReader: root logger can not have attribute 'tags'")
			}
		} else {
			var ss []string
			for _, s := range strings.Split(base.Tags, ",") {
				if s = strings.TrimSpace(s); s == "" {
					continue
				}
				ss = append(ss, s)
			}
			if len(ss) == 0 {
				return fmt.Errorf("RefreshReader: logger must have attribute 'tags' except root logger")
			}
			for _, s := range ss {
				cTags[s] = logger
			}
		}
	}

	if cRoot == nil {
		return errors.New("RefreshReader: found no root logger")
	}

	var (
		logArray []*Logger
		tagArray []*regexp.Regexp
	)

	for _, s := range util.OrderedMapKeys(cTags) {
		r, err := regexp.Compile(s)
		if err != nil {
			return errutil.WrapError(err, "RefreshReader: `%s` regexp compile error", s)
		}
		tagArray = append(tagArray, r)
		logArray = append(logArray, cTags[s])
	}

	for _, a := range cAppenders {
		if err := a.Start(); err != nil {
			return errutil.WrapError(err, "RefreshReader: appender %s start error", a.GetName())
		}
	}
	for _, l := range cLoggers {
		if err := l.Start(); err != nil {
			return errutil.WrapError(err, "RefreshReader: logger %s start error", l.GetName())
		}
	}

	for s, tag := range tagMap {
		logger := cRoot
		for i, r := range tagArray {
			if r.MatchString(s) {
				logger = logArray[i]
				break
			}
		}
		tag.SetLogger(logger)
	}

	return nil
}
