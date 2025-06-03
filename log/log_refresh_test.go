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
	"strings"
	"testing"

	"github.com/lvan100/go-assert"
)

type funcReader func(p []byte) (n int, err error)

func (r funcReader) Read(p []byte) (n int, err error) {
	return r(p)
}

func TestRefresh(t *testing.T) {

	t.Run("file not exist", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshFile("testdata/file-not-exist.xml")
		assert.ThatError(t, err).Matches("open testdata/file-not-exist.xml")
	})

	t.Run("already refresh", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshFile("testdata/log.xml")
		assert.Nil(t, err)
		// ...
		err = RefreshFile("testdata/log.xml")
		assert.ThatError(t, err).Matches("RefreshReader: log refresh already done")
	})

	t.Run("unsupported file", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(nil, ".json")
		assert.ThatError(t, err).Matches("RefreshReader: unsupported file type .json")
	})

	t.Run("read file error", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(funcReader(func(p []byte) (n int, err error) {
			return 0, errors.New("read error")
		}), ".xml")
		assert.ThatError(t, err).Matches("read error")
	})

	t.Run("read node error - 1", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(""), ".xml")
		assert.ThatError(t, err).Matches("invalid XML structure: missing root element")
	})

	t.Run("read node error - 2", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Map></Map>
		`), ".xml")
		assert.ThatError(t, err).Matches("RefreshReader: Configuration root not found")
	})

	t.Run("more Properties", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
				<Properties></Properties>
				<Properties></Properties>
			</Configuration>
		`), ".xml")
		assert.ThatError(t, err).Matches("RefreshReader: Properties section must be unique")
	})

	t.Run("error Property", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
				<Properties>
					<Property>abc</Property>
				</Properties>
			</Configuration>
		`), ".xml")
		assert.ThatError(t, err).Matches("RefreshReader: attribute 'name' not found for node Property")
	})

	t.Run("no Appenders", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
			</Configuration>
		`), ".xml")
		assert.ThatError(t, err).Matches("RefreshReader: Appenders section not found")
	})

	t.Run("more Appenders", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
				<Appenders></Appenders>
				<Appenders></Appenders>
			</Configuration>
		`), ".xml")
		assert.ThatError(t, err).Matches("RefreshReader: Appenders section must be unique")
	})

	t.Run("unfound Appender", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
				<Appenders>
					<NotExistAppender/>
				</Appenders>
			</Configuration>
		`), ".xml")
		assert.ThatError(t, err).Matches("RefreshReader: plugin NotExistAppender not found")
	})

	t.Run("Appender error - 1", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
				<Appenders>
					<File/>
				</Appenders>
			</Configuration>
		`), ".xml")
		assert.ThatError(t, err).Matches("RefreshReader: attribute 'name' not found")
	})

	t.Run("Appender error - 2", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
				<Appenders>
					<File name="file"/>
				</Appenders>
			</Configuration>
		`), ".xml")
		assert.ThatError(t, err).Matches("create plugin log.FileAppender error << found no plugin elements for struct field Layout")
	})

	t.Run("Appender error - 3", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
				<Appenders>
					<File name="file">
						<TextLayout/>
					</File>
				</Appenders>
			</Configuration>
		`), ".xml")
		assert.ThatError(t, err).Matches("create plugin log.FileAppender error << found no attribute for struct field FileName")
	})

	t.Run("no Loggers", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
				<Appenders>
					<File name="file" fileName="access.log">
						<TextLayout/>
					</File>
				</Appenders>
			</Configuration>
		`), ".xml")
		assert.ThatError(t, err).Matches("RefreshReader: Loggers section not found")
	})

	t.Run("more Loggers", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
				<Appenders>
					<File name="file" fileName="access.log">
						<TextLayout/>
					</File>
				</Appenders>
				<Loggers></Loggers>
				<Loggers></Loggers>
			</Configuration>
		`), ".xml")
		assert.ThatError(t, err).Matches("RefreshReader: Loggers section must be unique")
	})

	t.Run("no Logger", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
				<Appenders>
					<File name="file" fileName="access.log">
						<TextLayout/>
					</File>
				</Appenders>
				<Loggers></Loggers>
			</Configuration>
		`), ".xml")
		assert.ThatError(t, err).Matches("RefreshReader: found no root logger")
	})

	t.Run("Logger error - 1", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
				<Appenders>
					<File name="file" fileName="access.log">
						<TextLayout/>
					</File>
				</Appenders>
				<Loggers>
					<NotExistLogger/>
				</Loggers>
			</Configuration>
		`), ".xml")
		assert.ThatError(t, err).Matches("RefreshReader: plugin NotExistLogger not found")
	})

	t.Run("Logger error - 2", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
				<Appenders>
					<File name="file" fileName="access.log">
						<TextLayout/>
					</File>
				</Appenders>
				<Loggers>
					<Logger/>
				</Loggers>
			</Configuration>
		`), ".xml")
		assert.ThatError(t, err).Matches("RefreshReader: attribute 'name' not found for node Logger")
	})

	t.Run("Logger error - 3", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
				<Appenders>
					<File name="file" fileName="access.log">
						<TextLayout/>
					</File>
				</Appenders>
				<Loggers>
					<Logger name="biz"/>
				</Loggers>
			</Configuration>
		`), ".xml")
		assert.ThatError(t, err).Matches("create plugin log.LoggerConfig error << found no attribute for struct field Level")
	})

	t.Run("Logger error - 4", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
				<Appenders>
					<File name="file" fileName="access.log">
						<TextLayout/>
					</File>
				</Appenders>
				<Loggers>
					<Logger name="biz" level="info">
						<AppenderRef ref="not-exist-appender"/>
					</Logger>
				</Loggers>
			</Configuration>
		`), ".xml")
		assert.ThatError(t, err).Matches("RefreshReader: appender not-exist-appender not found")
	})

	t.Run("Logger error - 5", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
				<Appenders>
					<File name="file" fileName="access.log">
						<TextLayout/>
					</File>
				</Appenders>
				<Loggers>
					<Logger name="biz" level="info">
						<AppenderRef ref="file"/>
					</Logger>
				</Loggers>
			</Configuration>
		`), ".xml")
		assert.ThatError(t, err).Matches("RefreshReader: logger must have attribute 'tags' except root logger")
	})

	t.Run("Root error - 1", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
				<Appenders>
					<File name="file" fileName="access.log">
						<TextLayout/>
					</File>
				</Appenders>
				<Loggers>
					<Logger name="biz" level="info" tags="a,,b">
						<AppenderRef ref="file"/>
					</Logger>
				</Loggers>
			</Configuration>
		`), ".xml")
		assert.ThatError(t, err).Matches("RefreshReader: found no root logger")
	})

	t.Run("Root error - 2", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
				<Appenders>
					<File name="file" fileName="access.log">
						<TextLayout/>
					</File>
				</Appenders>
				<Loggers>
					<Root/>
				</Loggers>
			</Configuration>
		`), ".xml")
		assert.ThatError(t, err).Matches("create plugin log.LoggerConfig error << found no attribute for struct field Level")
	})

	t.Run("Root error - 3", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
				<Appenders>
					<File name="file" fileName="access.log">
						<TextLayout/>
					</File>
				</Appenders>
				<Loggers>
					<Root level="info"/>
				</Loggers>
			</Configuration>
		`), ".xml")
		assert.ThatError(t, err).Matches("create plugin log.LoggerConfig error << found no plugin elements for struct field AppenderRefs")
	})

	t.Run("Root error - 4", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
				<Appenders>
					<File name="file" fileName="access.log">
						<TextLayout/>
					</File>
				</Appenders>
				<Loggers>
					<Root level="info" tags="a,b,">
						<AppenderRef ref="file"/>
					</Root>
				</Loggers>
			</Configuration>
		`), ".xml")
		assert.ThatError(t, err).Matches("RefreshReader: root logger can not have attribute 'tags'")
	})

	t.Run("Root error - 5", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
				<Appenders>
					<File name="file" fileName="access.log">
						<TextLayout/>
					</File>
				</Appenders>
				<Loggers>
					<Root level="info">
						<AppenderRef ref="file"/>
					</Root>
					<AsyncRoot level="info">
						<AppenderRef ref="file"/>
					</AsyncRoot>
				</Loggers>
			</Configuration>
		`), ".xml")
		assert.ThatError(t, err).Matches("RefreshReader: found more than one root loggers")
	})

	t.Run("tag error", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
				<Appenders>
					<Console name="console">
						<TextLayout/>
					</Console>
					<File name="file" fileName="access.log">
						<TextLayout/>
					</File>
				</Appenders>
				<Loggers>
					<Root level="trace">
						<AppenderRef ref="console"/>
					</Root>
					<Logger name="biz" level="info" tags="a[">
						<AppenderRef ref="file"/>
					</Logger>
				</Loggers>
			</Configuration>
		`), ".xml")
		assert.ThatError(t, err).Matches("RefreshReader: `a\\[` regexp compile error << error parsing regexp: missing closing \\]: `\\[`")
	})

	t.Run("appender start error", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
				<Appenders>
					<Console name="console">
						<TextLayout/>
					</Console>
					<File name="file" fileName="/not-exist-dir/access.log">
						<TextLayout/>
					</File>
				</Appenders>
				<Loggers>
					<Root level="trace">
						<AppenderRef ref="console"/>
					</Root>
					<AsyncLogger name="biz" level="info" tags="a">
						<AppenderRef ref="file"/>
					</AsyncLogger>
				</Loggers>
			</Configuration>
		`), ".xml")
		assert.ThatError(t, err).Matches("RefreshReader: appender file start error << open /not-exist-dir/access.log: no such file or directory")
	})

	t.Run("logger start error", func(t *testing.T) {
		defer func() { initOnce.Store(false) }()
		err := RefreshReader(strings.NewReader(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
				<Appenders>
					<Console name="console">
						<TextLayout/>
					</Console>
					<File name="file" fileName="access.log">
						<TextLayout/>
					</File>
				</Appenders>
				<Loggers>
					<AsyncRoot level="trace" bufferSize="10">
						<AppenderRef ref="console"/>
					</AsyncRoot>
				</Loggers>
			</Configuration>
		`), ".xml")
		assert.ThatError(t, err).Matches("RefreshReader: logger ::root:: start error << bufferSize is too small")
	})
}
