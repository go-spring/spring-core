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

package main

import (
	"net/http"

	"github.com/go-spring/spring-core/gs"
	"github.com/go-spring/spring-core/util/sysconf"
	"github.com/go-spring/spring-core/util/syslog"
)

func init() {
	// Register the Service struct as a bean.
	gs.Object(&Service{})

	// Provide a [*http.ServeMux] as a bean.
	gs.Provide(func(s *Service) *http.ServeMux {
		http.HandleFunc("/echo", s.Echo)
		http.HandleFunc("/refresh", s.Refresh)
		return http.DefaultServeMux
	})
}

type Service struct {
	AppName gs.Dync[string] `value:"${spring.app.name}"`
}

func (s *Service) Echo(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(s.AppName.Value()))
}

func (s *Service) Refresh(w http.ResponseWriter, r *http.Request) {
	_ = sysconf.Set("spring.app.name", "refreshed")
	_ = gs.RefreshProperties()
	_, _ = w.Write([]byte("OK!"))
}

func main() {
	// Set the application name in the configuration.
	_ = sysconf.Set("spring.app.name", "go-spring")

	// Start the Go-Spring application. If it fails, log the error.
	if err := gs.Run(); err != nil {
		syslog.Errorf("app run failed: %s", err.Error())
	}
}

// ➜ curl http://127.0.0.1:9090/echo
// go-spring
// ➜ curl http://127.0.0.1:9090/refresh
// OK!
// ➜ curl http://127.0.0.1:9090/echo
// refreshed
