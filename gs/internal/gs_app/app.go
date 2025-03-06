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

// Package gs_app provides a framework for building and managing Go-Spring applications.
package gs_app

import (
	"context"
	"errors"
	"fmt"
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

var App = NewApplication()

// Application represents the core application, managing its lifecycle, configuration,
// and the injection of dependencies.
type Application struct {
	C *gs_core.Container
	P *gs_conf.AppConfig

	exiting atomic.Bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup

	Runners []gs.Runner `autowire:"${spring.app.runners:=*?}"`
	Jobs    []gs.Job    `autowire:"${spring.app.jobs:=*?}"`
	Servers []gs.Server `autowire:"${spring.app.servers:=*?}"`
}

// NewApplication creates and initializes a new application instance.
func NewApplication() *Application {
	ctx, cancel := context.WithCancel(context.Background())
	return &Application{
		C:      gs_core.New(),
		P:      gs_conf.NewAppConfig(),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Run starts the application and listens for termination signals
// (e.g., SIGINT, SIGTERM). When a signal is received, it shuts down
// the application gracefully. Use ShutDown but not Stop to end
// the application lifecycle.
func (app *Application) Run() error {
	app.C.Object(app)

	if err := app.Start(); err != nil {
		return err
	}

	// listens for OS termination signals
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		sig := <-ch
		app.ShutDown(fmt.Sprintf("Received signal: %v", sig))
	}()

	// waits for the shutdown signal
	<-app.ctx.Done()
	app.Stop()
	return nil
}

// Start initializes and starts the application. It loads configuration properties,
// refreshes the IoC container, performs dependency injection, and runs runners
// and servers.
func (app *Application) Start() error {
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
	for _, r := range app.Jobs {
		goutil.GoFunc(func() {
			if err := r.Run(app.ctx); err != nil {
				app.ShutDown(fmt.Sprintf("job run error: %s", err.Error()))
			}
		})
	}

	// starts all servers
	for _, svr := range app.Servers {
		app.wg.Add(1)
		goutil.GoFunc(func() {
			defer app.wg.Done()
			if err := svr.Serve(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				app.ShutDown(fmt.Sprintf("server serve error: %s", err.Error()))
			}
		})
	}

	return nil
}

// Stop gracefully stops the application. This method is used to clean up
// resources and stop servers started by the Start method.
func (app *Application) Stop() {
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
func (app *Application) Exiting() bool {
	return app.exiting.Load()
}

// ShutDown gracefully terminates the application. It should be used when
// shutting down the application started by Run.
func (app *Application) ShutDown(msg ...string) {
	app.exiting.Store(true)
	app.cancel()
}
