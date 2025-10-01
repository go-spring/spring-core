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
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-spring/gs-mock/gsmock"
	"github.com/go-spring/log"
	"github.com/go-spring/spring-base/testing/assert"
	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_conf"
	"github.com/go-spring/spring-core/util/goutil"
)

var logBuf = &bytes.Buffer{}

func init() {
	goutil.OnPanic = func(ctx context.Context, r any, stack []byte) {
		log.Panicf(ctx, log.TagAppDef, "panic: %v\n%s\n", r, stack)
	}
}

func Reset() {
	logBuf.Reset()
	log.Stdout = logBuf
	os.Args = nil
	os.Clearenv()
	gs_conf.SysConf = conf.New()
}

func TestApp(t *testing.T) {

	t.Run("property conflict", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		fileID := gs_conf.SysConf.AddFile("app_test.go")
		_ = gs_conf.SysConf.Set("a", "123", fileID)
		_ = os.Setenv("GS_A_B", "456")
		app := NewApp()
		err := app.Start()
		assert.Error(t, err).Matches("property conflict at path a.b")
	})

	t.Run("bean creation failure", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		app := NewApp()
		app.C.Root(app.C.Provide(func() (*http.Server, error) {
			return nil, errors.New("fail to create bean")
		}))
		err := app.Start()
		assert.Error(t, err).Matches("fail to create bean")
	})

	t.Run("runner panic", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		r := gs.FuncRunner(func() error {
			panic("runner panic")
		})

		app := NewApp()
		app.C.Object(r).AsRunner()

		assert.Panic(t, func() {
			_ = app.Start()
		}, "runner panic")
	})

	t.Run("runner return error", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		r := gs.FuncRunner(func() error {
			return errors.New("runner return error")
		})

		app := NewApp()
		app.C.Object(r).AsRunner()
		err := app.Start()
		assert.Error(t, err).Matches("runner return error")
	})

	t.Run("multiple runners with error", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		app := NewApp()

		// success
		r1 := gs.FuncRunner(func() error {
			return nil
		})
		app.C.Object(r1).AsRunner().Name("r1")

		// error
		r2 := gs.FuncRunner(func() error {
			return errors.New("runner error")
		})
		app.C.Object(r2).AsRunner().Name("r2")

		err := app.Start()
		assert.Error(t, err).Matches("runner error")
	})

	t.Run("disable jobs & servers", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		fileID := gs_conf.SysConf.AddFile("app_test.go")
		_ = gs_conf.SysConf.Set("spring.app.enable-jobs", "false", fileID)
		_ = gs_conf.SysConf.Set("spring.app.enable-servers", "false", fileID)
		app := NewApp()
		go func() {
			time.Sleep(50 * time.Millisecond)
			assert.That(t, app.EnableJobs).False()
			assert.That(t, app.EnableServers).False()
			assert.That(t, len(app.Jobs)).Equal(0)
			assert.That(t, len(app.Servers)).Equal(0)
			assert.That(t, len(app.Runners)).Equal(0)
			app.ShutDown()
		}()
		err := app.Start()
		assert.That(t, err).Nil()
		app.WaitForShutdown()
		time.Sleep(50 * time.Millisecond)
		assert.String(t, logBuf.String()).Contains("shutdown complete")
	})

	t.Run("job panic", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		r := gs.FuncJob(func(ctx context.Context) error {
			panic("job panic")
		})

		app := NewApp()
		app.C.Object(r).AsJob()
		err := app.Start()
		assert.That(t, err).Nil()
		time.Sleep(50 * time.Millisecond)
		assert.String(t, logBuf.String()).Contains("panic: job panic")
	})

	t.Run("job return error", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		r := gs.FuncJob(func(ctx context.Context) error {
			return errors.New("job return error")
		})

		app := NewApp()
		app.C.Object(r).AsJob()
		err := app.Start()
		assert.That(t, err).Nil()
		time.Sleep(50 * time.Millisecond)
		assert.String(t, logBuf.String()).Contains("job run error: job return error")
	})

	t.Run("job context cancel", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		jobFinished := make(chan bool, 1)
		r := gs.FuncJob(func(ctx context.Context) error {
			<-ctx.Done()
			jobFinished <- true
			return ctx.Err()
		})

		app := NewApp()
		app.C.Object(r).AsJob()

		go func() {
			time.Sleep(50 * time.Millisecond)
			app.ShutDown()
		}()

		err := app.Start()
		assert.That(t, err).Nil()
		app.WaitForShutdown()
		<-jobFinished
		assert.String(t, logBuf.String()).Contains("job run error: context canceled")
	})

	t.Run("server return error", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		m := gsmock.NewManager()
		r := gs.NewServerMockImpl(m)
		r.MockShutdown().ReturnDefault()
		r.MockListenAndServe().ReturnValue(errors.New("server return error"))

		app := NewApp()
		app.C.Object(r).AsServer()
		err := app.Start()
		assert.Error(t, err).String("server intercepted")
		time.Sleep(50 * time.Millisecond)
		assert.String(t, logBuf.String()).Contains("server serve error: server return error")
	})

	t.Run("server panic", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		m := gsmock.NewManager()
		r := gs.NewServerMockImpl(m)
		r.MockShutdown().ReturnDefault()
		r.MockListenAndServe().Handle(func(sig gs.ReadySignal) error {
			panic("server panic")
		})

		app := NewApp()
		app.C.Object(r).AsServer()
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
			r := gs.FuncRunner(func() error {
				return nil
			})
			app.C.Object(r).AsRunner().Name("r1")
		}
		{
			r := gs.FuncRunner(func() error {
				return nil
			})
			app.C.Object(r).AsRunner().Name("r2")
		}
		{
			r := gs.FuncJob(func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			})
			app.C.Object(r).AsJob().Name("j1")
		}
		j2Wait := make(chan struct{}, 1)
		{
			r := gs.FuncJob(func(ctx context.Context) error {
				for {
					time.Sleep(time.Millisecond)
					if app.Exiting() {
						j2Wait <- struct{}{}
						return nil
					}
				}
			})
			app.C.Object(r).AsJob().Name("j2")
		}
		{
			m := gsmock.NewManager()
			r := gs.NewServerMockImpl(m)
			r.MockShutdown().ReturnDefault()
			r.MockListenAndServe().Handle(func(sig gs.ReadySignal) error {
				<-sig.TriggerAndWait()
				return nil
			})

			app.C.Object(r).AsServer().Name("s1")
		}
		{
			m := gsmock.NewManager()
			r := gs.NewServerMockImpl(m)
			r.MockShutdown().ReturnDefault()
			r.MockListenAndServe().Handle(func(sig gs.ReadySignal) error {
				<-sig.TriggerAndWait()
				return nil
			})

			app.C.Object(r).AsServer().Name("s2")
		}
		go func() {
			time.Sleep(50 * time.Millisecond)
			assert.That(t, app.EnableJobs).True()
			assert.That(t, app.EnableServers).True()
			assert.That(t, len(app.Jobs)).Equal(2)
			assert.That(t, len(app.Servers)).Equal(2)
			assert.That(t, len(app.Runners)).Equal(2)
			app.ShutDown()
		}()
		err := app.Start()
		assert.That(t, err).Nil()
		app.WaitForShutdown()
		time.Sleep(50 * time.Millisecond)
		<-j2Wait
		assert.String(t, logBuf.String()).Contains("shutdown complete")
	})

	t.Run("shutdown error", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		app := NewApp()

		m := gsmock.NewManager()
		r := gs.NewServerMockImpl(m)
		r.MockShutdown().Handle(func(ctx context.Context) error {
			return errors.New("server shutdown error")
		})
		r.MockListenAndServe().Handle(func(sig gs.ReadySignal) error {
			<-sig.TriggerAndWait()
			return nil
		})

		app.C.Object(r).AsServer()
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
