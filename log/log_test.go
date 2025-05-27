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
	"context"
	"strings"
	"testing"

	"github.com/go-spring/spring-core/log"
	"github.com/lvan100/go-assert"
)

var TagRequestIn = log.GetTag("_com_request_in")
var TagRequestOut = log.GetTag("_com_request_out")

func TestLog(t *testing.T) {
	ctx := t.Context()

	log.StringFromContext = func(ctx context.Context) string {
		return ""
	}

	log.FieldsFromContext = func(ctx context.Context) []log.Field {
		traceID, _ := ctx.Value("trace_id").(string)
		spanID, _ := ctx.Value("span_id").(string)
		return []log.Field{
			log.String("trace_id", traceID),
			log.String("span_id", spanID),
		}
	}

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
			<Properties>
				<Property name="LayoutBufferSize">100KB</Property>
			</Properties>
			<Appenders>
				<Console name="Console_JSON">
					<JSONLayout bufferSize="${LayoutBufferSize}"/>
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
	`
	err := log.RefreshReader(strings.NewReader(xml), ".xml")
	assert.Nil(t, err)

	ctx = context.WithValue(ctx, "trace_id", "0a882193682db71edd48044db54cae88")
	ctx = context.WithValue(ctx, "span_id", "50ef0724418c0a66")

	log.Debug(ctx, TagRequestOut, func() []log.Field {
		return []log.Field{
			log.Msgf("hello %s", "world"),
		}
	})

	log.Infof("hello %s", "world")
	log.Info(ctx, TagRequestIn, log.Msgf("hello %s", "world"))
}
