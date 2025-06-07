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

		l := &LoggerConfig{baseLoggerConfig{
			Level: InfoLevel,
			Tags:  "_com_*",
			AppenderRefs: []*AppenderRef{
				{appender: a},
			},
		}}

		err = l.Start()
		assert.Nil(t, err)

		assert.False(t, l.EnableLevel(TraceLevel))
		assert.False(t, l.EnableLevel(DebugLevel))
		assert.True(t, l.EnableLevel(InfoLevel))
		assert.True(t, l.EnableLevel(WarnLevel))
		assert.True(t, l.EnableLevel(ErrorLevel))
		assert.True(t, l.EnableLevel(PanicLevel))
		assert.True(t, l.EnableLevel(FatalLevel))

		for i := 0; i < 5; i++ {
			l.Publish(&Event{})
		}

		assert.That(t, a.count).Equal(5)

		l.Stop()
		a.Stop()
	})
}

func TestAsyncLoggerConfig(t *testing.T) {

	t.Run("enable level", func(t *testing.T) {
		l := &AsyncLoggerConfig{
			baseLoggerConfig: baseLoggerConfig{
				Level: InfoLevel,
			},
		}

		assert.False(t, l.EnableLevel(TraceLevel))
		assert.False(t, l.EnableLevel(DebugLevel))
		assert.True(t, l.EnableLevel(InfoLevel))
		assert.True(t, l.EnableLevel(WarnLevel))
		assert.True(t, l.EnableLevel(ErrorLevel))
		assert.True(t, l.EnableLevel(PanicLevel))
		assert.True(t, l.EnableLevel(FatalLevel))
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

		for i := 0; i < 5000; i++ {
			l.Publish(GetEvent())
		}

		time.Sleep(200 * time.Millisecond)

		l.Stop()
		a.Stop()

		assert.True(t, dropCount > 0)
	})

	t.Run("success", func(t *testing.T) {
		a := &CountAppender{
			Appender: &DiscardAppender{},
		}

		err := a.Start()
		assert.Nil(t, err)

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

		for i := 0; i < 5; i++ {
			l.Publish(GetEvent())
		}

		time.Sleep(100 * time.Millisecond)
		assert.That(t, a.count).Equal(5)

		l.Stop()
		a.Stop()
	})
}
