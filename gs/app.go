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

	"github.com/go-spring/log"
	"github.com/go-spring/spring-core/gs/internal/gs_app"
)

type AppStarter struct{}

// initApp initializes the app.
func (s *AppStarter) initApp() error {
	printBanner()
	if err := initLog(); err != nil {
		return err
	}
	if err := B.(*gs_app.BootImpl).Run(); err != nil {
		return err
	}
	B = nil
	return nil
}

// Run runs the app and waits for an interrupt signal to exit.
func (s *AppStarter) Run() {
	s.RunWith(nil)
}

// RunWith runs the app with a given function and waits for an interrupt signal to exit.
func (s *AppStarter) RunWith(fn func(ctx context.Context) error) {
	var err error
	defer func() {
		if err != nil {
			log.Errorf(context.Background(), log.TagAppDef, "app run failed: %v", err)
		}
	}()
	if err = s.initApp(); err != nil {
		return
	}
	if err = app.RunWith(fn); err != nil {
		return
	}
	log.Destroy()
}

// RunAsync runs the app asynchronously and returns a function to stop the app.
func (s *AppStarter) RunAsync() (func(), error) {
	if err := s.initApp(); err != nil {
		return nil, err
	}
	if err := app.Start(); err != nil {
		return nil, err
	}
	return func() {
		app.Stop()
		log.Destroy()
	}, nil
}
