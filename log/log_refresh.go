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
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// global holds the global logger and appender registry protected by a mutex for thread safety.
var global struct {
	Mutex     sync.Mutex          // Mutex to ensure thread-safe updates
	Loggers   map[string]*Logger  // Registered loggers by name
	Appenders map[string]Appender // Registered appenders by name
}

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
	global.Mutex.Lock()
	defer global.Mutex.Unlock()

	var rootNode *Node
	{
		r, ok := readers[ext]
		if !ok {
			return fmt.Errorf("unsupported file type %s", ext)
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
		return errors.New("the Configuration root not found")
	}

	var (
		cRoot      *Logger
		cLoggers   = make(map[string]*Logger)
		cAppenders = make(map[string]Appender)
		cMarkers   = make(map[string]*Logger)
	)

	// Parse <Appenders> section
	if node := rootNode.getChild("Appenders"); node != nil {
		for _, c := range node.Children {
			p, ok := plugins[c.Label]
			if !ok {
				return fmt.Errorf("plugin %s not found", c.Label)
			}
			name, ok := c.Attributes["name"]
			if !ok {
				return errors.New("attribute 'name' not found")
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
					return errors.New("found more than one root loggers")
				}
				c.Attributes["name"] = ""
			}

			p, ok := plugins[c.Label]
			if !ok || p == nil {
				return fmt.Errorf("plugin %s not found", c.Label)
			}
			name, ok := c.Attributes["name"]
			if !ok {
				return errors.New("attribute 'name' not found")
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
					return fmt.Errorf("appender %s not found", r.Ref)
				}
				r.appender = appender
			}

			if isRootLogger {
				if base.Marker != "" {
					return fmt.Errorf("root logger can not have marker")
				}
			} else {
				if base.Marker == "" {
					return fmt.Errorf("logger must have marker except root logger")
				}
				ss := strings.Split(base.Marker, ",")
				for _, s := range ss {
					cMarkers[s] = logger
				}
			}
		}
	}

	if cRoot == nil {
		return errors.New("found no root logger")
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

	for s, marker := range markers {
		logger, ok := cMarkers[s]
		if !ok {
			marker.SetLogger(cRoot)
			continue
		}
		marker.SetLogger(logger)
	}

	Stop(context.Background())

	global.Loggers = cLoggers
	global.Appenders = cAppenders
	return nil
}

// Stop stops all currently active loggers and appenders.
// This ensures a clean shutdown before applying new configurations.
func Stop(ctx context.Context) {
	for _, l := range global.Loggers {
		l.Stop(ctx)
	}
	for _, a := range global.Appenders {
		a.Stop(ctx)
	}
}
