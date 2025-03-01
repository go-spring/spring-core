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
	C gs.Container
	P *gs_conf.AppConfig

	exiting   atomic.Bool
	exitChan  chan struct{}
	waitGroup sync.WaitGroup

	Runners []gs.Runner `autowire:"${spring.app.runners:=*?}"`
	Jobs    []gs.Job    `autowire:"${spring.app.jobs:=*?}"`
	Servers []gs.Server `autowire:"${spring.app.servers:=*?}"`
}

// NewApplication creates and initializes a new application instance.
func NewApplication() *Application {
	app := &Application{
		C:        gs_core.New(),
		P:        gs_conf.NewAppConfig(),
		exitChan: make(chan struct{}),
	}
	app.C.Register(gs_core.NewBean(app))
	return app
}

// Run starts the application and listens for termination signals
// (e.g., SIGINT, SIGTERM). When a signal is received, it shuts down
// the application gracefully. Use ShutDown but not Stop to end
// the application lifecycle.
func (app *Application) Run() error {
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
	<-app.exitChan
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
		r.Run()
	}

	// runs all jobs
	for _, r := range app.Jobs {
		app.Go(func() {
			r.Run()
		})
	}

	// starts all servers
	for _, svr := range app.Servers {
		app.Go(func() {
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
	timeout := time.Second * 5
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	waitChan := make(chan struct{})
	goutil.Go(ctx, func(ctx context.Context) {
		var wg sync.WaitGroup
		for _, svr := range app.Servers {
			wg.Add(1)
			goutil.GoFunc(func() {
				defer wg.Done()
				if err := svr.Shutdown(ctx); err != nil {
					syslog.Errorf("shutdown server failed: %s", err.Error())
				}
			})
		}
		wg.Wait()
		app.waitGroup.Wait()
		waitChan <- struct{}{}
	})

	select {
	case <-waitChan:
	case <-ctx.Done():
		syslog.Infof("shutdown timeout")
	}

	app.C.Close()
}

// Exiting returns a boolean indicating whether the application is exiting.
func (app *Application) Exiting() bool {
	return app.exiting.Load()
}

// ShutDown gracefully terminates the application. It should be used when
// shutting down the application started by Run.
func (app *Application) ShutDown(msg ...string) {
	select {
	case <-app.exitChan:
		// do nothing if the exit channel is already closed
	default:
		app.exiting.Store(true)
		close(app.exitChan)
	}
}

// Go starts a new goroutine to execute the given function.
func (app *Application) Go(fn func()) {
	app.waitGroup.Add(1)
	goutil.GoFunc(func() {
		defer app.waitGroup.Done()
		fn()
	})
}
