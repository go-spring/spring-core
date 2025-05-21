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
	"testing"
	"time"

	"github.com/lvan100/go-assert"
)

type CountAppender struct {
	Appender
	count int
}

func (c *CountAppender) Append(e *Event) {
	c.count++
	c.Appender.Append(e)
}

func TestLoggerConfig(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		a := &CountAppender{
			Appender: &DiscardAppender{},
		}

		err := a.Start()
		assert.Nil(t, err)

		defer func() {
			a.Stop()
		}()

		l := &LoggerConfig{baseLoggerConfig{
			Level: InfoLevel,
			Tags:  "_com_*",
			AppenderRefs: []*AppenderRef{
				{appender: a},
			},
		}}

		err = l.Start()
		assert.Nil(t, err)

		defer func() {
			l.Stop()
		}()

		assert.False(t, l.enableLevel(TraceLevel))
		assert.False(t, l.enableLevel(DebugLevel))
		assert.True(t, l.enableLevel(InfoLevel))
		assert.True(t, l.enableLevel(WarnLevel))
		assert.True(t, l.enableLevel(ErrorLevel))
		assert.True(t, l.enableLevel(PanicLevel))
		assert.True(t, l.enableLevel(FatalLevel))

		for i := 0; i < 5; i++ {
			l.publish(&Event{})
		}

		assert.That(t, a.count).Equal(5)
	})
}

func TestAsyncLoggerConfig(t *testing.T) {

	t.Run("enable level", func(t *testing.T) {
		l := &AsyncLoggerConfig{
			baseLoggerConfig: baseLoggerConfig{
				Level: InfoLevel,
			},
		}

		assert.False(t, l.enableLevel(TraceLevel))
		assert.False(t, l.enableLevel(DebugLevel))
		assert.True(t, l.enableLevel(InfoLevel))
		assert.True(t, l.enableLevel(WarnLevel))
		assert.True(t, l.enableLevel(ErrorLevel))
		assert.True(t, l.enableLevel(PanicLevel))
		assert.True(t, l.enableLevel(FatalLevel))
	})

	t.Run("error BufferSize", func(t *testing.T) {
		l := &AsyncLoggerConfig{
			baseLoggerConfig: baseLoggerConfig{
				Name: "file",
			},
			BufferSize: 10,
		}

		err := l.Start()
		assert.ThatError(t, err).Matches("bufferSize is too small")
	})

	t.Run("drop events", func(t *testing.T) {
		a := &CountAppender{
			Appender: &DiscardAppender{},
		}

		err := a.Start()
		assert.Nil(t, err)

		defer func() {
			a.Stop()
		}()

		dropCount := 0
		OnDropEvent = func(*Event) {
			dropCount++
		}
		defer func() {
			OnDropEvent = nil
		}()

		l := &AsyncLoggerConfig{
			baseLoggerConfig: baseLoggerConfig{
				Level: InfoLevel,
				Tags:  "_com_*",
				AppenderRefs: []*AppenderRef{
					{appender: a},
				},
			},
			BufferSize: 100,
		}

		err = l.Start()
		assert.Nil(t, err)

		defer func() {
			l.Stop()
			assert.True(t, dropCount > 0)
		}()

		for i := 0; i < 5000; i++ {
			l.publish(GetEvent())
		}

		time.Sleep(200 * time.Millisecond)
	})

	t.Run("success", func(t *testing.T) {
		a := &CountAppender{
			Appender: &DiscardAppender{},
		}

		err := a.Start()
		assert.Nil(t, err)

		defer func() {
			a.Stop()
		}()

		l := &AsyncLoggerConfig{
			baseLoggerConfig: baseLoggerConfig{
				Level: InfoLevel,
				Tags:  "_com_*",
				AppenderRefs: []*AppenderRef{
					{appender: a},
				},
			},
			BufferSize: 100,
		}

		err = l.Start()
		assert.Nil(t, err)

		defer func() {
			l.Stop()
		}()

		for i := 0; i < 5; i++ {
			l.publish(GetEvent())
		}

		time.Sleep(100 * time.Millisecond)
		assert.That(t, a.count).Equal(5)
	})
}
