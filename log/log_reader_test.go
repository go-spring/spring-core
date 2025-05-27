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
	"testing"

	"github.com/lvan100/go-assert"
)

func TestXMLReader(t *testing.T) {

	t.Run("empty", func(t *testing.T) {
		reader := XMLReader{}
		_, err := reader.Read([]byte(``))
		assert.ThatError(t, err).Matches("invalid XML structure: missing root element")
	})

	t.Run("invalid", func(t *testing.T) {
		reader := XMLReader{}
		_, err := reader.Read([]byte(`<>`))
		assert.ThatError(t, err).Matches("XML syntax error on line 1: .*")
	})

	t.Run("success", func(t *testing.T) {
		reader := XMLReader{}
		node, err := reader.Read([]byte(`
			<?xml version="1.0" encoding="UTF-8"?>
			<Configuration>
				<Properties>
					<Property name="MaxBufferSize">1048576</Property>
					<Property name="Dummy">foo,bar</Property>
				</Properties>
				<Appenders>
					<Console name="Console_JSON">
						<JSONLayout/>
					</Console>
					<Console name="Console_Text">
						<TextLayout/>
					</Console>
				</Appenders>
				<Loggers>
					<Root level="trace">
						<AppenderRef ref="Console_Text"/>
					</Root>
					<Logger name="file" level="trace" tags="_com_request_*">
						<AppenderRef ref="Console_JSON"/>
					</Logger>
				</Loggers>
			</Configuration>
		`))
		assert.Nil(t, err)

		buf := bytes.NewBuffer(nil)
		buf.WriteString("\n")
		DumpNode(node, 3, buf)
		assert.ThatString(t, buf.String()).Equal(`
			Configuration
				Properties
					Property {name=MaxBufferSize} : 1048576
					Property {name=Dummy} : foo,bar
				Appenders
					Console {name=Console_JSON}
						JSONLayout
					Console {name=Console_Text}
						TextLayout
				Loggers
					Root {level=trace}
						AppenderRef {ref=Console_Text}
					Logger {level=trace name=file tags=_com_request_*}
						AppenderRef {ref=Console_JSON}`)
	})
}
