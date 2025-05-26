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
	"strconv"
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
	defer func() {
		_ = file.Close()
	}()
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
		return errors.New("RefreshReader: the Configuration root not found")
	}

	var (
		cRoot      *Logger
		cLoggers   = make(map[string]*Logger)
		cAppenders = make(map[string]Appender)
		cTags      = make(map[string]*Logger)
	)

	// Parse <Appenders> section
	if node := rootNode.getChild("Appenders"); node != nil {
		for _, c := range node.Children {
			p, ok := plugins[c.Label]
			if !ok {
				return fmt.Errorf("RefreshReader: plugin %s not found", c.Label)
			}
			name, ok := c.Attributes["name"]
			if !ok {
				return errors.New("RefreshReader: attribute 'name' not found")
			}
			v, err := NewPlugin(p.Class, c)
			if err != nil {
				return err
			}
			cAppenders[name] = v.Interface().(Appender)
		}
	}

	// Parse <Loggers> section
	if node := rootNode.getChild("Loggers"); node != nil {
		for _, c := range node.Children {
			isRootLogger := c.Label == "Root" || c.Label == "AsyncRoot"
			if isRootLogger {
				if cRoot != nil {
					return errors.New("RefreshReader: found more than one root loggers")
				}
				c.Attributes["name"] = ""
			}

			p, ok := plugins[c.Label]
			if !ok || p == nil {
				return fmt.Errorf("RefreshReader: plugin %s not found", c.Label)
			}
			name, ok := c.Attributes["name"]
			if !ok {
				return errors.New("RefreshReader: attribute 'name' not found")
			}
			v, err := NewPlugin(p.Class, c)
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
					return fmt.Errorf("RefreshReader: root logger can not have tags attribute")
				}
			} else {
				if base.Tags == "" {
					return fmt.Errorf("RefreshReader: logger must have tags attribute except root logger")
				}
				ss := strings.Split(base.Tags, ",")
				for _, s := range ss {
					cTags[s] = logger
				}
			}
		}
	}

	if cRoot == nil {
		return errors.New("found no root logger")
	}

	var (
		cfgMaxBufferSize int
	)

	if node := rootNode.getChild("Properties"); node != nil {
		strMaxBufferSize, ok := node.Attributes["MaxBufferSize"]
		if ok {
			i, err := strconv.Atoi(strMaxBufferSize)
			if err != nil {
				return err
			}
			if i <= 0 {
				return errors.New("RefreshReader: MaxBufferSize must be greater than 0")
			}
			cfgMaxBufferSize = i
		}
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
			return err
		}
	}
	for _, l := range cLoggers {
		if err := l.Start(); err != nil {
			return err
		}
	}

	for s, tag := range tags {
		logger := cRoot
		for i, r := range tagArray {
			if r.MatchString(s) {
				logger = logArray[i]
				break
			}
		}
		tag.SetLogger(logger)
	}

	maxBufferSize.Store(int32(cfgMaxBufferSize))

	return nil
}
