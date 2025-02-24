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
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_conf"
	"github.com/go-spring/spring-core/gs/internal/gs_core"
)

// AppContext provides a wrapper around the application's [gs.Context].
// It offers controlled access to the internal context and is designed
// to ensure safe usage in the application's lifecycle.
type AppContext struct {
	c gs.Context
}

// Unsafe exposes the underlying [gs.Context]. Using this method in new
// goroutines is unsafe because [gs.Context] may release its resources
// (e.g., bean definitions), making binding and injection operations invalid.
func (p *AppContext) Unsafe() gs.Context {
	return p.c
}

// Go executes a function in a new goroutine. The provided function will receive
// a cancellation signal when the application begins shutting down.
func (p *AppContext) Go(fn func(ctx context.Context)) {
	p.c.(interface {
		Go(fn func(ctx context.Context))
	}).Go(fn)
}

// AppRunner defines an interface for tasks that should be executed after all
// beans are injected but before the application's servers are started.
// It is commonly used to initialize background jobs or tasks.
type AppRunner interface {
	Run(ctx *AppContext)
}

// AppServer defines an interface for managing the lifecycle of application servers,
// such as HTTP, gRPC, Thrift, or MQ servers. Servers must implement methods for
// starting and stopping gracefully.
type AppServer interface {
	OnAppStart(ctx *AppContext)
	OnAppStop(ctx context.Context)
}

// App represents the core application, managing its lifecycle, configuration,
// and the injection of dependencies.
type App struct {
	C gs.Container
	P *gs_conf.AppConfig

	exitChan chan struct{}

	Runners []AppRunner `autowire:"${spring.app.runners:=*?}"`
	Servers []AppServer `autowire:"${spring.app.servers:=*?}"`
}

// NewApp creates and initializes a new application instance.
func NewApp() *App {
	app := &App{
		C:        gs_core.New(),
		P:        gs_conf.NewAppConfig(),
		exitChan: make(chan struct{}),
	}
	app.C.Object(app)
	return app
}

// Run starts the application and listens for termination signals
// (e.g., SIGINT, SIGTERM). When a signal is received, it shuts down
// the application gracefully. Use ShutDown but not Stop to end
// the application lifecycle.
func (app *App) Run() error {
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

	c := &AppContext{c: app.C.(gs.Context)}

	// runs all runners
	for _, r := range app.Runners {
		r.Run(c)
	}

	// starts all servers
	for _, svr := range app.Servers {
		svr.OnAppStart(c)
	}

	// listens the cancel signal then stop the servers
	c.Go(func(ctx context.Context) {
		<-ctx.Done()
		var wg sync.WaitGroup
		for _, svr := range app.Servers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				svr.OnAppStop(ctx)
			}()
		}
		wg.Wait()
	})

	app.C.ReleaseUnusedMemory()
	return nil
}

// Stop gracefully stops the application. This method is used to clean up
// resources and stop servers started by the Start method.
func (app *App) Stop() {
	app.C.Close()
}

// ShutDown gracefully terminates the application. It should be used when
// shutting down the application started by Run.
func (app *App) ShutDown(msg ...string) {
	select {
	case <-app.exitChan:
		// do nothing if the exit channel is already closed
	default:
		close(app.exitChan)
	}
}
