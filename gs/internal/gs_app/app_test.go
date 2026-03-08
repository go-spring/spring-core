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

package gs_app

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-spring/gs-mock/gsmock"
	"github.com/go-spring/log"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/stdlib/errutil"
	"github.com/go-spring/stdlib/goutil"
	"github.com/go-spring/stdlib/testing/assert"
)

var logBuf = &bytes.Buffer{}

func init() {
	goutil.OnPanic = func(ctx context.Context, info goutil.PanicInfo) {
		log.Panicf(ctx, log.TagAppDef, "panic: %v\n%s\n", info.Panic, info.Stack)
	}
}

func Reset() {
	logBuf.Reset()
	log.Stdout = logBuf
	os.Args = nil
	os.Clearenv()
}

type funcRunner struct {
	fn func(ctx context.Context) error
}

func (f *funcRunner) Run(ctx context.Context) error {
	return f.fn(ctx)
}

func TestApp(t *testing.T) {

	t.Run("property conflict", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		app := NewApp()
		app.Property("a", "123")
		_ = os.Setenv("GS_A_B", "456")
		err := app.Start()
		assert.Error(t, err).Nil() // .Matches("path a.b conflicts with existing structure")
	})

	t.Run("bean creation failure", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		app := NewApp()
		app.c.Provide(func() (Runner, error) {
			return nil, errutil.Explain(nil, "fail to create bean")
		})
		err := app.Start()
		assert.Error(t, err).Matches("fail to create bean")
	})

	t.Run("runner panic", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		r := &funcRunner{fn: func(ctx context.Context) error {
			panic("runner panic")
		}}

		app := NewApp()
		app.c.Provide(r).Export(gs.As[Runner]())

		assert.Panic(t, func() {
			_ = app.Start()
		}, "runner panic")
	})

	t.Run("runner return error", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		r := &funcRunner{fn: func(ctx context.Context) error {
			return errutil.Explain(nil, "runner return error")
		}}

		app := NewApp()
		app.c.Provide(r).Export(gs.As[Runner]())
		err := app.Start()
		assert.Error(t, err).Matches("runner return error")
	})

	t.Run("multiple runners with error", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		app := NewApp()

		// success
		r1 := &funcRunner{fn: func(ctx context.Context) error {
			return nil
		}}
		app.c.Provide(r1).Export(gs.As[Runner]()).Name("r1")

		// error
		r2 := &funcRunner{fn: func(ctx context.Context) error {
			return errutil.Explain(nil, "runner error")
		}}
		app.c.Provide(r2).Export(gs.As[Runner]()).Name("r2")

		err := app.Start()
		assert.Error(t, err).Matches("runner error")
	})

	t.Run("server return error", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		m := gsmock.NewManager()
		r := NewServerMockImpl(m)
		r.MockStop().ReturnDefault()
		r.MockRun().ReturnValue(errutil.Explain(nil, "server return error"))

		app := NewApp()
		app.c.Provide(r).Export(gs.As[Server]())
		err := app.Start()
		assert.Error(t, err).String("server intercepted")
		time.Sleep(50 * time.Millisecond)
		assert.String(t, logBuf.String()).Contains("server serve error: server return error")
	})

	t.Run("server panic", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		m := gsmock.NewManager()
		r := NewServerMockImpl(m)
		r.MockStop().ReturnDefault()
		r.MockRun().Handle(func(ctx context.Context, sig ReadySignal) error {
			panic("server panic")
		})

		app := NewApp()
		app.c.Provide(r).Export(gs.As[Server]())
		err := app.Start()
		assert.Error(t, err).String("server intercepted")
		time.Sleep(50 * time.Millisecond)
		assert.String(t, logBuf.String()).Contains("panic: server panic")
	})

	t.Run("success", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		app := NewApp()
		{
			r := &funcRunner{fn: func(ctx context.Context) error {
				return nil
			}}
			app.c.Provide(r).Export(gs.As[Runner]()).Name("r1")
		}
		{
			r := &funcRunner{fn: func(ctx context.Context) error {
				return nil
			}}
			app.c.Provide(r).Export(gs.As[Runner]()).Name("r2")
		}
		{
			m := gsmock.NewManager()
			r := NewServerMockImpl(m)
			r.MockStop().ReturnDefault()
			r.MockRun().Handle(func(ctx context.Context, sig ReadySignal) error {
				<-sig.TriggerAndWait()
				return nil
			})

			app.c.Provide(r).Export(gs.As[Server]()).Name("s1")
		}
		{
			m := gsmock.NewManager()
			r := NewServerMockImpl(m)
			r.MockStop().ReturnDefault()
			r.MockRun().Handle(func(ctx context.Context, sig ReadySignal) error {
				<-sig.TriggerAndWait()
				return nil
			})

			app.c.Provide(r).Export(gs.As[Server]()).Name("s2")
		}
		go func() {
			time.Sleep(50 * time.Millisecond)
			assert.That(t, len(app.Servers)).Equal(2)
			assert.That(t, len(app.Runners)).Equal(2)
			app.ShutDown()
		}()
		err := app.Start()
		assert.That(t, err).Nil()
		app.WaitForShutdown()
		time.Sleep(50 * time.Millisecond)
		assert.String(t, logBuf.String()).Contains("shutdown complete")
	})

	t.Run("shutdown error", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		app := NewApp()

		m := gsmock.NewManager()
		r := NewServerMockImpl(m)
		r.MockStop().Handle(func() error {
			return errutil.Explain(nil, "server shutdown error")
		})
		r.MockRun().Handle(func(ctx context.Context, sig ReadySignal) error {
			<-sig.TriggerAndWait()
			return nil
		})

		app.c.Provide(r).Export(gs.As[Server]())
		go func() {
			time.Sleep(50 * time.Millisecond)
			app.ShutDown()
		}()
		err := app.Start()
		assert.That(t, err).Nil()
		app.WaitForShutdown()
		time.Sleep(50 * time.Millisecond)
		assert.String(t, logBuf.String()).Contains("shutdown server failed: server shutdown error")
	})
}
