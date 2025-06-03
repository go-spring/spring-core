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
	"os"
	"testing"
	"time"

	"github.com/lvan100/go-assert"
)

func TestDiscardAppender(t *testing.T) {
	a := &DiscardAppender{}
	err := a.Start()
	assert.Nil(t, err)
	a.Append(&Event{})
	a.Stop()
}

func TestConsoleAppender(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		file, err := os.CreateTemp(os.TempDir(), "")
		assert.Nil(t, err)

		oldStdout := os.Stdout
		os.Stdout = file
		defer func() {
			os.Stdout = oldStdout
		}()

		a := &ConsoleAppender{
			BaseAppender: BaseAppender{
				Layout: &TextLayout{
					BaseLayout{
						FileLineLength: 48,
					},
				},
			},
		}
		a.Append(&Event{
			Level:     InfoLevel,
			Time:      time.Time{},
			File:      "file.go",
			Line:      100,
			Tag:       "_def",
			Fields:    []Field{Msg("hello world")},
			CtxFields: nil,
		})

		err = file.Close()
		assert.Nil(t, err)

		b, err := os.ReadFile(file.Name())
		assert.Nil(t, err)
		assert.ThatString(t, string(b)).Equal("[INFO][0001-01-01T00:00:00.000][file.go:100] _def||msg=hello world\n")
	})
}

func TestFileAppender(t *testing.T) {

	t.Run("Start error", func(t *testing.T) {
		a := &FileAppender{
			BaseAppender: BaseAppender{
				Layout: &TextLayout{
					BaseLayout{
						FileLineLength: 48,
					},
				},
			},
			FileName: "/not-exist-dir/file.log",
		}
		err := a.Start()
		assert.ThatError(t, err).Matches("open /not-exist-dir/file.log: no such file or directory")
	})

	t.Run("success", func(t *testing.T) {
		file, err := os.CreateTemp(os.TempDir(), "")
		assert.Nil(t, err)
		err = file.Close()
		assert.Nil(t, err)

		a := &FileAppender{
			BaseAppender: BaseAppender{
				Layout: &TextLayout{
					BaseLayout{
						FileLineLength: 48,
					},
				},
			},
			FileName: file.Name(),
		}
		err = a.Start()
		assert.Nil(t, err)

		a.Append(&Event{
			Level:     InfoLevel,
			Time:      time.Time{},
			File:      "file.go",
			Line:      100,
			Tag:       "_def",
			Fields:    []Field{Msg("hello world")},
			CtxFields: nil,
		})

		a.Stop()

		b, err := os.ReadFile(a.file.Name())
		assert.Nil(t, err)
		assert.ThatString(t, string(b)).Equal("[INFO][0001-01-01T00:00:00.000][file.go:100] _def||msg=hello world\n")
	})
}
