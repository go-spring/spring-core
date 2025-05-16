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

package log_test

import (
	"strings"
	"testing"

	"github.com/go-spring/spring-core/log"
)

var TagRequestIn = log.GetTag("_com_request_in")
var TagRequestOut = log.GetTag("_com_request_out")

func TestLog(t *testing.T) {
	ctx := t.Context()

	log.Debug(ctx, TagRequestOut, func() []log.Field {
		return []log.Field{
			log.Msgf("hello %s", "world"),
		}
	})

	log.Infof("hello %s", "world")
	log.Info(ctx, TagRequestIn, log.Msgf("hello %s", "world"))

	xml := `
		<?xml version="1.0" encoding="UTF-8"?>
		<Configuration>
			<Appenders>
				<Console name="Console_JSON">
					<JSONLayout/>
				</Console>
				<Console name="Console_Pattern">
					<TextLayout/>
				</Console>
			</Appenders>
			<Loggers>
				<Root level="trace">
					<AppenderRef ref="Console_JSON"/>
				</Root>
				<Logger name="file" level="trace" tags="_com_request_in,_com_request_out">
					<AppenderRef ref="Console_Pattern"/>
				</Logger>
			</Loggers>
		</Configuration>
	`
	err := log.RefreshReader(strings.NewReader(xml), ".xml")
	if err != nil {
		t.Fatal(err)
	}

	log.Debug(ctx, TagRequestOut, func() []log.Field {
		return []log.Field{
			log.Msgf("hello %s", "world"),
		}
	})

	log.Infof("hello %s", "world")
	log.Info(ctx, TagRequestIn, log.Msgf("hello %s", "world"))
}
