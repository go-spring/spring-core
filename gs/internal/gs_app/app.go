/*
 * Copyright 2024 The Go-Spring Authors.
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
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_conf"
	"github.com/go-spring/spring-core/gs/internal/gs_core"
	"github.com/go-spring/spring-core/util/goutil"
	"github.com/go-spring/spring-core/util/syslog"
)

// GS is the global application instance.
var GS = NewApp()

// App represents the core application, managing its lifecycle,
// configuration, and dependency injection.
type App struct {
	C *gs_core.Container
	P *gs_conf.AppConfig

	exiting atomic.Bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup

	Runners []gs.Runner `autowire:"${spring.app.runners:=*?}"`
	Jobs    []gs.Job    `autowire:"${spring.app.jobs:=*?}"`
	Servers []gs.Server `autowire:"${spring.app.servers:=*?}"`

	EnableJobs    bool `autowire:"${spring.enable.app-jobs:=true}"`
	EnableServers bool `autowire:"${spring.enable.app-servers:=true}"`
}

// NewApp creates and initializes a new application instance.
func NewApp() *App {
	ctx, cancel := context.WithCancel(context.Background())
	return &App{
		C:      gs_core.New(),
		P:      gs_conf.NewAppConfig(),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Run starts the application and listens for termination signals
// (e.g., SIGINT, SIGTERM). Upon receiving a signal, it initiates
// a graceful shutdown.
func (app *App) Run() error {
	app.C.Object(app)

	if err := app.Start(); err != nil {
		return err
	}

	// listens for OS termination signals
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		sig := <-ch
		syslog.Infof("Received signal: %v", sig)
		app.ShutDown()
	}()

	// waits for the shutdown signal
	<-app.ctx.Done()
	app.Stop()
	return nil
}

// Start initializes and starts the application. It performs configuration
// loading, IoC container refreshing, dependency injection, and runs
// runners, jobs and servers.
func (app *App) Start() error {
	// loads the layered app properties
	p, err := app.P.Refresh()
	if err != nil {
		return err
	}

	// refreshes the container properties
	err = app.C.RefreshProperties(p)
	if err != nil {
		return err
	}

	// refreshes the container
	err = app.C.Refresh()
	if err != nil {
		return err
	}

	// runs all runners
	for _, r := range app.Runners {
		if err := r.Run(); err != nil {
			return err
		}
	}

	// runs all jobs
	if app.EnableJobs {
		for _, job := range app.Jobs {
			goutil.GoFunc(func() {
				defer func() {
					if r := recover(); r != nil {
						app.ShutDown()
						panic(r)
					}
				}()
				if err := job.Run(app.ctx); err != nil {
					syslog.Errorf("job run error: %s", err.Error())
					app.ShutDown()
				}
			})
		}
	}

	// starts all servers
	if app.EnableServers {
		sig := NewReadySignal()
		for _, svr := range app.Servers {
			sig.Add()
			app.wg.Add(1)
			goutil.GoFunc(func() {
				defer app.wg.Done()
				defer func() {
					if r := recover(); r != nil {
						sig.Intercept()
						app.ShutDown()
						panic(r)
					}
				}()
				err := svr.ListenAndServe(sig)
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					syslog.Errorf("server serve error: %s", err.Error())
					sig.Intercept()
					app.ShutDown()
				}
			})
		}
		sig.Wait()
		if sig.Intercepted() {
			return nil
		}
		syslog.Infof("ready to serve requests")
		sig.Close()
	}
	return nil
}

// Stop gracefully shuts down the application, ensuring all servers and
// resources are properly closed.
func (app *App) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	waitChan := make(chan struct{})
	goutil.GoFunc(func() {
		for _, svr := range app.Servers {
			goutil.GoFunc(func() {
				if err := svr.Shutdown(ctx); err != nil {
					syslog.Errorf("shutdown server failed: %s", err.Error())
				}
			})
		}
		app.wg.Wait()
		app.C.Close()
		waitChan <- struct{}{}
	})

	select {
	case <-waitChan:
		syslog.Infof("shutdown complete")
	case <-ctx.Done():
		syslog.Infof("shutdown timeout")
	}
}

// Exiting returns a boolean indicating whether the application is exiting.
func (app *App) Exiting() bool {
	return app.exiting.Load()
}

// ShutDown gracefully terminates the application. This method should
// be called to trigger a proper shutdown process.
func (app *App) ShutDown() {
	app.exiting.Store(true)
	app.cancel()
}
