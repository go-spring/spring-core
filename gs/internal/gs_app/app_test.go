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
	"runtime/debug"
	"testing"
	"time"

	"github.com/go-spring/gs-assert/assert"
	"github.com/go-spring/gs-mock/gsmock"
	"github.com/go-spring/log"
	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_conf"
	"github.com/go-spring/spring-core/util/goutil"
)

var logBuf = &bytes.Buffer{}

func init() {
	goutil.OnPanic = func(ctx context.Context, r any) {
		log.Panicf(ctx, log.TagAppDef, "panic: %v\n%s\n", r, debug.Stack())
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

	t.Run("os signals", func(t *testing.T) {
		t.Skip()

		Reset()
		t.Cleanup(Reset)

		app := NewApp()
		go func() {
			time.Sleep(50 * time.Millisecond)
			p, err := os.FindProcess(os.Getpid())
			assert.That(t, err).Nil()
			err = p.Signal(os.Interrupt)
			assert.That(t, err).Nil()
			time.Sleep(50 * time.Millisecond)
		}()
		err := app.Run()
		assert.That(t, err).Nil()
		time.Sleep(50 * time.Millisecond)
		assert.ThatString(t, logBuf.String()).Contains("Received signal: interrupt")
	})

	t.Run("config refresh error", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		fileID := gs_conf.SysConf.AddFile("app_test.go")
		_ = gs_conf.SysConf.Set("a", "123", fileID)
		_ = os.Setenv("GS_A_B", "456")
		app := NewApp()
		err := app.Run()
		assert.ThatError(t, err).Matches("property conflict at path a.b")
	})

	t.Run("container refresh error", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		app := NewApp()
		app.C.RootBean(app.C.Provide(func() (*http.Server, error) {
			return nil, errors.New("fail to create bean")
		}))
		err := app.Run()
		assert.ThatError(t, err).Matches("fail to create bean")
	})

	t.Run("runner return error", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		m := gsmock.NewManager()
		r := gs.NewRunnerMockImpl(m)
		r.MockRun().ReturnValue(errors.New("runner return error"))

		app := NewApp()
		app.C.Object(r).AsRunner()
		err := app.Run()
		assert.ThatError(t, err).Matches("runner return error")
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
		err := app.Run()
		assert.That(t, err).Nil()
		time.Sleep(50 * time.Millisecond)
		assert.ThatString(t, logBuf.String()).Contains("shutdown complete")
	})

	t.Run("job return error", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		m := gsmock.NewManager()
		r := gs.NewJobMockImpl(m)
		r.MockRun().ReturnValue(errors.New("job return error"))

		app := NewApp()
		app.C.Object(r).AsJob()
		err := app.Run()
		assert.That(t, err).Nil()
		time.Sleep(50 * time.Millisecond)
		assert.ThatString(t, logBuf.String()).Contains("job run error: job return error")
	})

	t.Run("job panic", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		m := gsmock.NewManager()
		r := gs.NewJobMockImpl(m)
		r.MockRun().Handle(func(ctx context.Context) error {
			panic("job panic")
		})

		app := NewApp()
		app.C.Object(r).AsJob()
		err := app.Run()
		assert.That(t, err).Nil()
		time.Sleep(50 * time.Millisecond)
		assert.ThatString(t, logBuf.String()).Contains("panic: job panic")
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
		err := app.Run()
		assert.That(t, err).Nil()
		time.Sleep(50 * time.Millisecond)
		assert.ThatString(t, logBuf.String()).Contains("server serve error: server return error")
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
		err := app.Run()
		assert.That(t, err).Nil()
		time.Sleep(50 * time.Millisecond)
		assert.ThatString(t, logBuf.String()).Contains("panic: server panic")
	})

	t.Run("success", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		app := NewApp()
		{
			m := gsmock.NewManager()
			r := gs.NewRunnerMockImpl(m)
			r.MockRun().ReturnDefault()

			app.C.Object(r).AsRunner().Name("r1")
		}
		{
			m := gsmock.NewManager()
			r := gs.NewRunnerMockImpl(m)
			r.MockRun().ReturnDefault()

			app.C.Object(r).AsRunner().Name("r2")
		}
		{
			m := gsmock.NewManager()
			r := gs.NewJobMockImpl(m)
			r.MockRun().Handle(func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			})

			app.C.Object(r).AsJob().Name("j1")
		}
		j2Wait := make(chan struct{}, 1)
		{
			m := gsmock.NewManager()
			r := gs.NewJobMockImpl(m)
			r.MockRun().Handle(func(ctx context.Context) error {
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
		err := app.Run()
		assert.That(t, err).Nil()
		time.Sleep(50 * time.Millisecond)
		<-j2Wait
		assert.ThatString(t, logBuf.String()).Contains("shutdown complete")
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
		err := app.Run()
		assert.That(t, err).Nil()
		time.Sleep(50 * time.Millisecond)
		assert.ThatString(t, logBuf.String()).Contains("shutdown server failed: server shutdown error")
	})
}
