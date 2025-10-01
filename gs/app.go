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
	"syscall"

	"github.com/go-spring/log"
	"github.com/go-spring/spring-base/util"
	"github.com/go-spring/spring-core/gs/internal/gs_app"
)

// AppStarter is a wrapper to manage the lifecycle of a Spring application.
// It handles initialization, running, graceful shutdown, and logging.
type AppStarter struct{}

// startApp initializes logging, runs the Boot implementation,
// and then starts the main application.
func (s *AppStarter) startApp() error {

	// Print application banner at startup
	printBanner()

	// Initialize logger
	if err := initLog(); err != nil {
		return err
	}

	// Run Boot implementation (pre-app setup)
	if err := B.(*gs_app.BootImpl).Run(); err != nil {
		return err
	}
	B = nil // Release Boot instance after running

	// Start the application
	if err := app.Start(); err != nil {
		return err
	}
	return nil
}

// stopApp waits for the application to shut down and cleans up resources.
// NOTE: ShutDown() must be called before invoking this function.
func (s *AppStarter) stopApp() {
	app.WaitForShutdown()
	log.Destroy()
}

// Run starts the application, optionally runs a user-defined callback,
// and waits for termination signals (e.g., SIGTERM, Ctrl+C) to trigger graceful shutdown.
func (s *AppStarter) Run(fn ...func() error) {

	// Start application
	if err := s.startApp(); err != nil {
		err = util.WrapError(err, "start app failed")
		log.Errorf(context.Background(), log.TagAppDef, "%s", err)
		return
	}

	// Execute user-provided callback after app starts
	if len(fn) > 0 && fn[0] != nil {
		if err := fn[0](); err != nil {
			err = util.WrapError(err, "start app failed")
			log.Errorf(context.Background(), log.TagAppDef, "%s", err)
			return
		}
	}

	// Start a goroutine to listen for OS interrupt or termination signals
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		sig := <-ch
		signal.Stop(ch)
		close(ch)
		log.Infof(context.Background(), log.TagAppDef, "Received signal: %v", sig)
		app.ShutDown()
	}()

	// Wait until shutdown completes
	s.stopApp()
}

// RunAsync starts the application asynchronously and returns a function
// that can be used to trigger shutdown from outside.
func (s *AppStarter) RunAsync() (func(), error) {

	// Start application
	if err := s.startApp(); err != nil {
		err = util.WrapError(err, "start app failed")
		log.Errorf(context.Background(), log.TagAppDef, "%s", err)
		return nil, err
	}

	// Return a shutdown function
	return func() {
		app.ShutDown()
		s.stopApp()
	}, nil
}
