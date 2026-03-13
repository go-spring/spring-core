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

//go:generate gs mock -o=app_mock.go -i=Server

package gs_app

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/go-spring/log"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_conf"
	"github.com/go-spring/spring-core/gs/internal/gs_core"
	"github.com/go-spring/stdlib/errutil"
	"github.com/go-spring/stdlib/flatten"
	"github.com/go-spring/stdlib/goutil"
)

// Runner defines an interface for components that need to be executed
// after all beans have been injected but before servers start.
type Runner interface {
	Run(ctx context.Context) error
}

// ReadySignal defines an interface for signaling application readiness.
// Servers can use this to indicate when they are ready to accept requests.
type ReadySignal interface {
	TriggerAndWait() <-chan struct{}
}

// Server defines the lifecycle of application servers (e.g., HTTP, gRPC).
// It provides methods to start and gracefully stop the server.
type Server interface {
	Run(ctx context.Context, sig ReadySignal) error
	Stop() error
}

// ContextAware provides access to the application's root context.
// Users can inject this bean to access the App's context.
type ContextAware struct {
	Context context.Context
}

// ConfigRefresher is an interface for components that need to refresh
// application properties after configuration changes.
type ConfigRefresher struct {
	app *App
}

// RefreshProperties refreshes application properties and
// propagates the changes to the IoC container.
func (c *ConfigRefresher) RefreshProperties() error {
	return c.app.RefreshProperties()
}

// App represents the core application, managing its lifecycle,
// configuration, and dependency injection.
type App struct {
	c *gs_core.Container // IoC container
	p *gs_conf.AppConfig // Application configuration

	exiting atomic.Bool        // Indicates whether the app is shutting down
	ctx     context.Context    // Root context for managing cancellation
	cancel  context.CancelFunc // Function to cancel the root context
	wg      sync.WaitGroup     // WaitGroup to track running servers

	Runners []Runner `autowire:"${spring.app.runners:=?}"`
	Servers []Server `autowire:"${spring.app.servers:=?}"`

	roots []*gs_bean.BeanDefinition // Root beans for container refresh
}

// NewApp creates a new App instance with an initialized root context.
func NewApp() *App {
	ctx := context.WithValue(context.Background(), "app", "")
	ctx, cancel := context.WithCancel(ctx)
	return &App{
		c:      gs_core.New(),
		p:      gs_conf.NewAppConfig(),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Context returns the root context for the application.
func (app *App) Context() context.Context {
	return app.ctx
}

// Property sets an app-level property in the application's configuration.
// It associates the property with the caller file for traceability.
func (app *App) Property(key string, val string) {
	app.p.Properties.Set(key, val)
}

// Provide registers a new bean definition in the IoC container.
// The parameter can be either an existing instance or a constructor function.
// Additional arguments can be passed for dependency injection.
func (app *App) Provide(objOrCtor any, args ...gs.Arg) *gs_bean.BeanDefinition {
	return app.c.Provide(objOrCtor, args...).Caller(2)
}

// Root registers a root bean for container refresh.
func (app *App) Root(b *gs_bean.BeanDefinition) *gs_bean.BeanDefinition {
	app.roots = append(app.roots, b)
	return b
}

// RefreshProperties reloads application properties from all sources
// and propagates the changes to the IoC container.
func (app *App) RefreshProperties() error {
	p, err := app.p.Refresh()
	if err != nil {
		return err
	}
	return app.c.RefreshProperties(p)
}

// initLog initializes the application's logging system.
func (app *App) initLog(p flatten.Storage) error {
	// No logging configuration
	if !p.Exists("logging") {
		return nil
	}
	s := flatten.NewPrefixedStorage(p, "logging")
	return log.Refresh(s)
}

// Start initializes and launches the application.
// The startup sequence is:
//  1. Refresh application properties from all sources
//  2. Initialize logging system
//  3. Register the App, ContextAware, and ConfigRefresher beans in the container
//  4. Refresh the IoC container to wire all beans
//  5. Clear the temporary root bean list after container refresh
//  6. Execute all Runner beans sequentially
//  7. Start all configured servers in separate goroutines
//     - Each server signals readiness via ReadySignal
//     - If a server panics or returns an unexpected error, ReadySignal is intercepted
//     and the application initiates a graceful shutdown
//  8. Wait until all servers signal readiness or intercept occurs
func (app *App) Start() error {

	// Load and refresh application properties
	p, err := app.p.Refresh()
	if err != nil {
		return err
	}

	// Initialize logger
	if err = app.initLog(p); err != nil {
		return err
	}

	app.Root(app.c.Provide(app))
	app.c.Provide(&ContextAware{app.ctx})
	app.c.Provide(&ConfigRefresher{app})

	// Refresh IoC container to wire all beans
	if err = app.c.Refresh(p, app.roots); err != nil {
		return err
	}

	// todo 清理运行时资源
	app.roots = nil

	// Execute all Runner beans sequentially
	for _, r := range app.Runners {
		if err = r.Run(app.ctx); err != nil {
			return err
		}
	}

	// Start all configured servers
	if len(app.Servers) > 0 {
		sig := NewReadySignal() // Coordinate readiness across servers
		for _, svr := range app.Servers {
			sig.Add()
			app.wg.Add(1)
			goutil.Go(app.ctx, func(ctx context.Context) {
				defer app.wg.Done()
				defer func() {
					// Recover from server panics and trigger shutdown
					if r := recover(); r != nil {
						sig.Intercept()
						app.ShutDown()
						panic(r)
					}
				}()
				err := svr.Run(ctx, sig)
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Errorf(ctx, log.TagAppDef, "server serve error: %v", err)
					sig.Intercept()
					app.ShutDown()
				} else {
					log.Infof(ctx, log.TagAppDef, "server closed")
				}
			}, false)
		}

		// Wait until all servers signal readiness
		sig.Wait()
		if sig.Intercepted() {
			log.Infof(app.ctx, log.TagAppDef, "server intercepted")
			return errutil.Explain(nil, "server intercepted")
		}
		log.Infof(app.ctx, log.TagAppDef, "ready to serve requests")
		sig.Close()
	}
	return nil
}

// WaitForShutdown blocks until the application is signaled to shut down.
// After shutdown is triggered:
//  1. All servers are stopped concurrently
//  2. Waits for all server goroutines to complete
//  3. Closes the IoC container
//  4. Cleans up and destroys the logging system
func (app *App) WaitForShutdown() {
	// Block until the root context is cancelled
	<-app.ctx.Done()

	// Stop all servers concurrently
	for _, svr := range app.Servers {
		goutil.Go(app.ctx, func(ctx context.Context) {
			if err := svr.Stop(); err != nil {
				log.Errorf(ctx, log.TagAppDef, "shutdown server failed: %v", err)
			}
		}, true)
	}

	app.wg.Wait()
	app.c.Close()
	log.Infof(app.ctx, log.TagAppDef, "shutdown complete")
	log.Destroy()
}

// Exiting indicates whether the application is currently shutting down.
func (app *App) Exiting() bool {
	return app.exiting.Load()
}

// ShutDown initiates a graceful shutdown of the application.
// It sets the exiting flag and cancels the root context.
func (app *App) ShutDown() {
	if app.exiting.CompareAndSwap(false, true) {
		log.Infof(app.ctx, log.TagAppDef, "shutting down")
		app.cancel()
	}
}
