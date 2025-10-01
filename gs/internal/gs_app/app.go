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
	"sync"
	"sync/atomic"

	"github.com/go-spring/log"
	"github.com/go-spring/spring-base/util"
	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_conf"
	"github.com/go-spring/spring-core/gs/internal/gs_core"
	"github.com/go-spring/spring-core/util/goutil"
)

// App represents the core application, managing its lifecycle,
// configuration, and dependency injection.
type App struct {
	C *gs_core.Container // IoC container
	P *gs_conf.AppConfig // Application configuration

	exiting atomic.Bool        // Indicates whether the application is shutting down
	ctx     context.Context    // Root context for managing cancellation
	cancel  context.CancelFunc // Function to cancel the root context
	wg      sync.WaitGroup     // WaitGroup to track running jobs and servers

	Runners []gs.Runner `autowire:"${spring.app.runners:=?}"`
	Jobs    []gs.Job    `autowire:"${spring.app.jobs:=?}"`
	Servers []gs.Server `autowire:"${spring.app.servers:=?}"`

	EnableJobs    bool `value:"${spring.app.enable-jobs:=true}"`
	EnableServers bool `value:"${spring.app.enable-servers:=true}"`
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

// Start initializes and launches the application. It performs the following steps:
// 1. Registers the App itself as a root bean.
// 2. Loads application configuration.
// 3. Refreshes the IoC container to initialize and wire beans.
// 4. Runs all registered Runners.
// 5. Launches Jobs (if enabled) as background goroutines.
// 6. Starts all Servers (if enabled) and waits for readiness.
func (app *App) Start() error {
	// Register App as a root bean in the container
	app.C.Root(app.C.Object(app))

	// Load layered application properties
	var p conf.Properties
	{
		var err error
		if p, err = app.P.Refresh(); err != nil {
			return err
		}
	}

	// Refresh the container to wire all beans
	if err := app.C.Refresh(p); err != nil {
		return err
	}

	// Run all registered Runners
	for _, r := range app.Runners {
		if err := r.Run(); err != nil {
			return err
		}
	}

	// Launch all Jobs (if enabled) as background tasks
	if app.EnableJobs {
		for _, job := range app.Jobs {
			app.wg.Add(1)
			goutil.Go(app.ctx, func(ctx context.Context) {
				defer app.wg.Done()
				defer func() {
					// Handle unexpected panics by shutting down the app
					if r := recover(); r != nil {
						app.ShutDown()
						panic(r)
					}
				}()
				if err := job.Run(ctx); err != nil {
					log.Errorf(ctx, log.TagAppDef, "job run error: %v", err)
					app.ShutDown()
				}
			})
		}
	}

	// Start all Servers (if enabled)
	if app.EnableServers {
		sig := NewReadySignal() // Used to coordinate readiness among servers
		for _, svr := range app.Servers {
			sig.Add()
			app.wg.Add(1)
			goutil.Go(app.ctx, func(ctx context.Context) {
				defer app.wg.Done()
				defer func() {
					// Handle server panics by intercepting readiness and shutting down
					if r := recover(); r != nil {
						sig.Intercept()
						app.ShutDown()
						panic(r)
					}
				}()
				err := svr.ListenAndServe(sig)
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Errorf(ctx, log.TagAppDef, "server serve error: %v", err)
					sig.Intercept()
					app.ShutDown()
				} else {
					log.Infof(ctx, log.TagAppDef, "server closed")
				}
			})
		}

		// Wait for all servers to be ready
		sig.Wait()
		if sig.Intercepted() {
			log.Infof(app.ctx, log.TagAppDef, "server intercepted")
			return util.FormatError(nil, "server intercepted")
		}
		log.Infof(app.ctx, log.TagAppDef, "ready to serve requests")
		sig.Close()
	}
	return nil
}

// WaitForShutdown waits for the application to be signaled to shut down
// and then gracefully stops all servers and jobs.
func (app *App) WaitForShutdown() {
	// Wait until the application context is cancelled (triggered by ShutDown)
	<-app.ctx.Done()

	// Gracefully shut down all running servers
	for _, svr := range app.Servers {
		goutil.Go(app.ctx, func(ctx context.Context) {
			if err := svr.Shutdown(context.Background()); err != nil {
				log.Errorf(ctx, log.TagAppDef, "shutdown server failed: %v", err)
			}
		})
	}
	app.wg.Wait()
	app.C.Close()
	log.Infof(app.ctx, log.TagAppDef, "shutdown complete")
}

// Exiting returns whether the application is currently in the process of shutting down.
func (app *App) Exiting() bool {
	return app.exiting.Load()
}

// ShutDown initiates a graceful shutdown of the application by
// setting the exiting flag and cancelling the root context.
func (app *App) ShutDown() {
	if app.exiting.CompareAndSwap(false, true) {
		log.Infof(app.ctx, log.TagAppDef, "shutting down")
		app.cancel()
	}
}
