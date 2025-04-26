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
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/util/sysconf"
	"github.com/lvan100/go-assert"
	"go.uber.org/mock/gomock"
)

var logBuf = &bytes.Buffer{}

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(logBuf, nil)))
}

func clean() {
	logBuf.Reset()
	os.Args = nil
	os.Clearenv()
	sysconf.Clear()
}

func TestApp(t *testing.T) {

	t.Run("os signals", func(t *testing.T) {
		t.Skip()
		t.Cleanup(clean)
		app := NewApp()
		go func() {
			time.Sleep(50 * time.Millisecond)
			p, err := os.FindProcess(os.Getpid())
			assert.Nil(t, err)
			err = p.Signal(os.Interrupt)
			assert.Nil(t, err)
			time.Sleep(50 * time.Millisecond)
		}()
		err := app.Run()
		assert.Nil(t, err)
		time.Sleep(50 * time.Millisecond)
		assert.ThatString(t, logBuf.String()).Contains("Received signal: interrupt")
	})

	t.Run("config refresh error", func(t *testing.T) {
		t.Cleanup(clean)
		sysconf.Set("a", "123")
		_ = os.Setenv("GS_A_B", "456")
		app := NewApp()
		err := app.Run()
		assert.ThatError(t, err).Matches("property conflict at path a.b")
	})

	t.Run("container refresh error", func(t *testing.T) {
		t.Cleanup(clean)
		app := NewApp()
		app.C.Provide(func() (*http.Server, error) {
			return nil, errors.New("fail to create bean")
		})
		err := app.Run()
		assert.ThatError(t, err).Matches("fail to create bean")
	})

	t.Run("runner return error", func(t *testing.T) {
		t.Cleanup(clean)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		r := NewMockRunner(ctrl)
		r.EXPECT().Run().Return(errors.New("runner return error"))
		app := NewApp()
		app.C.Object(r).AsRunner()
		err := app.Run()
		assert.ThatError(t, err).Matches("runner return error")
	})

	t.Run("disable jobs & servers", func(t *testing.T) {
		t.Cleanup(clean)
		sysconf.Set("spring.app.enable-jobs", "false")
		sysconf.Set("spring.app.enable-servers", "false")
		app := NewApp()
		go func() {
			time.Sleep(50 * time.Millisecond)
			assert.False(t, app.EnableJobs)
			assert.False(t, app.EnableServers)
			assert.That(t, len(app.Jobs)).Equal(0)
			assert.That(t, len(app.Servers)).Equal(0)
			assert.That(t, len(app.Runners)).Equal(0)
			app.ShutDown()
		}()
		err := app.Run()
		assert.Nil(t, err)
		time.Sleep(50 * time.Millisecond)
		assert.ThatString(t, logBuf.String()).Contains("shutdown complete")
	})

	t.Run("job return error", func(t *testing.T) {
		t.Cleanup(clean)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		r := NewMockJob(ctrl)
		r.EXPECT().Run(gomock.Any()).Return(errors.New("job return error"))
		app := NewApp()
		app.C.Object(r).AsJob()
		err := app.Run()
		assert.Nil(t, err)
		time.Sleep(50 * time.Millisecond)
		assert.ThatString(t, logBuf.String()).Contains("job run error: job return error")
	})

	t.Run("job panic", func(t *testing.T) {
		t.Cleanup(clean)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		r := NewMockJob(ctrl)
		r.EXPECT().Run(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
			panic("job panic")
		})
		app := NewApp()
		app.C.Object(r).AsJob()
		err := app.Run()
		assert.Nil(t, err)
		time.Sleep(50 * time.Millisecond)
		assert.ThatString(t, logBuf.String()).Contains("panic: job panic")
	})

	t.Run("server return error", func(t *testing.T) {
		t.Cleanup(clean)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		r := NewMockServer(ctrl)
		r.EXPECT().Shutdown(gomock.Any()).Return(nil)
		r.EXPECT().ListenAndServe(gomock.Any()).Return(errors.New("server return error"))
		app := NewApp()
		app.C.Object(r).AsServer()
		err := app.Run()
		assert.Nil(t, err)
		time.Sleep(50 * time.Millisecond)
		assert.ThatString(t, logBuf.String()).Contains("server serve error: server return error")
	})

	t.Run("server panic", func(t *testing.T) {
		t.Cleanup(clean)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		r := NewMockServer(ctrl)
		r.EXPECT().Shutdown(gomock.Any()).Return(nil)
		r.EXPECT().ListenAndServe(gomock.Any()).DoAndReturn(func(sig gs.ReadySignal) error {
			panic("server panic")
		})
		app := NewApp()
		app.C.Object(r).AsServer()
		err := app.Run()
		assert.Nil(t, err)
		time.Sleep(50 * time.Millisecond)
		assert.ThatString(t, logBuf.String()).Contains("panic: server panic")
	})

	t.Run("success", func(t *testing.T) {
		t.Cleanup(clean)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		app := NewApp()
		{
			r1 := NewMockRunner(ctrl)
			r1.EXPECT().Run().Return(nil)
			app.C.Object(r1).AsRunner().Name("r1")
		}
		{
			r2 := NewMockRunner(ctrl)
			r2.EXPECT().Run().Return(nil)
			app.C.Object(r2).AsRunner().Name("r2")
		}
		{
			j1 := NewMockJob(ctrl)
			j1.EXPECT().Run(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			})
			app.C.Object(j1).AsJob().Name("j1")
		}
		j2Wait := make(chan struct{})
		{
			j2 := NewMockJob(ctrl)
			j2.EXPECT().Run(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
				for {
					time.Sleep(time.Millisecond)
					if app.Exiting() {
						j2Wait <- struct{}{}
						return nil
					}
				}
			})
			app.C.Object(j2).AsJob().Name("j2")
		}
		{
			s1 := NewMockServer(ctrl)
			s1.EXPECT().Shutdown(gomock.Any()).Return(nil)
			s1.EXPECT().ListenAndServe(gomock.Any()).DoAndReturn(func(sig gs.ReadySignal) error {
				<-sig.TriggerAndWait()
				return nil
			})
			app.C.Object(s1).AsServer().Name("s1")
		}
		{
			s2 := NewMockServer(ctrl)
			s2.EXPECT().Shutdown(gomock.Any()).Return(nil)
			s2.EXPECT().ListenAndServe(gomock.Any()).DoAndReturn(func(sig gs.ReadySignal) error {
				<-sig.TriggerAndWait()
				return nil
			})
			app.C.Object(s2).AsServer().Name("s2")
		}
		go func() {
			time.Sleep(50 * time.Millisecond)
			assert.That(t, app.ShutDownTimeout).Equal(time.Second * 15)
			assert.True(t, app.EnableJobs)
			assert.True(t, app.EnableServers)
			assert.That(t, len(app.Jobs)).Equal(2)
			assert.That(t, len(app.Servers)).Equal(2)
			assert.That(t, len(app.Runners)).Equal(2)
			app.ShutDown()
		}()
		err := app.Run()
		assert.Nil(t, err)
		time.Sleep(50 * time.Millisecond)
		<-j2Wait
		assert.ThatString(t, logBuf.String()).Contains("shutdown complete")
	})

	t.Run("shutdown timeout", func(t *testing.T) {
		t.Cleanup(clean)
		sysconf.Set("spring.app.shutdown-timeout", "10ms")
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		app := NewApp()
		s := NewMockServer(ctrl)
		s.EXPECT().Shutdown(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
			return nil
		})
		s.EXPECT().ListenAndServe(gomock.Any()).DoAndReturn(func(sig gs.ReadySignal) error {
			<-sig.TriggerAndWait()
			time.Sleep(time.Second)
			return nil
		})
		app.C.Object(s).AsServer()
		go func() {
			time.Sleep(50 * time.Millisecond)
			app.ShutDown()
		}()
		err := app.Run()
		assert.Nil(t, err)
		time.Sleep(50 * time.Millisecond)
		assert.ThatString(t, logBuf.String()).Contains("shutdown timeout")
	})

	t.Run("shutdown error", func(t *testing.T) {
		t.Cleanup(clean)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		app := NewApp()
		s := NewMockServer(ctrl)
		s.EXPECT().Shutdown(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
			return errors.New("server shutdown error")
		})
		s.EXPECT().ListenAndServe(gomock.Any()).DoAndReturn(func(sig gs.ReadySignal) error {
			<-sig.TriggerAndWait()
			return nil
		})
		app.C.Object(s).AsServer()
		go func() {
			time.Sleep(50 * time.Millisecond)
			app.ShutDown()
		}()
		err := app.Run()
		assert.Nil(t, err)
		time.Sleep(50 * time.Millisecond)
		assert.ThatString(t, logBuf.String()).Contains("shutdown server failed: server shutdown error")
	})
}
