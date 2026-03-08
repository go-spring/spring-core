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

package gs

import (
	"context"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"testing"

	"github.com/go-spring/log"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_app"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/stdlib/errutil"
	"github.com/go-spring/stdlib/goutil"
)

// inited indicates whether the application has been initialized.
var inited bool

// App defines the configuration interface of a Go-Spring application.
// Methods on App are only valid during application configuration
// and must not be called after the application has started.
type App interface {
	// Property sets a key-value property in the application configuration.
	Property(key string, val string)
	// Provide registers an object or constructor as a bean in the application.
	Provide(objOrCtor any, args ...gs.Arg) *gs_bean.BeanDefinition
}

// AppStarter wraps a gs_app.App and manages its lifecycle.
// It provides methods for initialization, configuration, starting,
// stopping, running, and testing the application.
type AppStarter struct {
	app *gs_app.App
	cfg func(App)
}

// Configure creates a new application and registers a configuration
// function that will be applied before the application starts.
func Configure(cfg func(App)) *AppStarter {
	inited = true
	return &AppStarter{app: gs_app.NewApp(), cfg: cfg}
}

// startApp starts the application lifecycle by printing the banner,
// applying the configuration function, and starting the underlying gs_app.App.
// Returns an error if the application fails to start.
func (s *AppStarter) startApp() error {

	// Print banner
	printBanner()

	// Apply user configuration
	if s.cfg != nil {
		s.cfg(s.app)
	}

	// Start application
	if err := s.app.Start(); err != nil {
		err = errutil.Stack(err, "start app failed")
		log.Errorf(s.app.Context(), log.TagAppDef, "%s", err)
		return err
	}

	return nil
}

// Run creates and starts a new application using default settings.
func Run() error {
	return Configure(nil).Run()
}

// Run starts the application, applies configuration, and waits for
// termination signals (e.g., SIGTERM, Ctrl+C) to trigger a graceful shutdown.
// If no servers are running, the application stops immediately.
func (s *AppStarter) Run() error {
	if err := s.startApp(); err != nil {
		return err
	}

	// Listen for termination signals in a separate goroutine
	goutil.Go(s.app.Context(), func(ctx context.Context) {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		sig := <-ch
		signal.Stop(ch)
		close(ch)
		log.Infof(ctx, log.TagAppDef, "Received signal: %v", sig)
		s.app.ShutDown()
	}, false)

	// Wait for shutdown to complete
	s.app.WaitForShutdown()
	return nil
}

// RunAsync runs the application asynchronously and
// returns a function to stop the application.
func RunAsync(i any) (stop func(), err error) {
	return Configure(nil).RunAsync(i)
}

// RunAsync runs the application asynchronously and
// returns a function to stop the application.
func (s *AppStarter) RunAsync(i any) (stop func(), err error) {

	// Register the root bean
	s.app.Provide(i).
		Name("__root__").
		Export(gs.As[any]())

	if err = s.startApp(); err != nil {
		return func() {}, err
	}

	return func() {
		s.app.ShutDown()
		s.app.WaitForShutdown()
	}, nil
}

// RunTest runs a test function using a new application instance.
// The test function must accept exactly one argument, which must be
// a pointer to a struct. The struct will be managed as a root bean
// in the application context.
func RunTest(t *testing.T, f any) {
	Configure(nil).RunTest(t, f)
}

// RunTest runs a user-defined test function with a provided test object.
// It initializes the application, registers the test object as a bean,
// starts the application, executes the test, and ensures graceful shutdown.
func (s *AppStarter) RunTest(t *testing.T, f any) {
	ft := reflect.TypeOf(f)
	obj := reflect.New(ft.In(0).Elem())

	stop, err := s.RunAsync(obj.Interface())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { stop() }()

	// Execute the test function
	reflect.ValueOf(f).Call([]reflect.Value{obj})
}
