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

package monitor

import (
	"context"
	"net"
	"net/http"
	"net/http/pprof"

	"github.com/go-spring/spring-core/gs"
)

func init() {
	gs.Provide(NewServer, gs.TagArg("${spring.monitor}")).AsServer().Condition(
		gs.OnProperty("spring.monitor.enable").HavingValue("true"),
	)
}

// ServerConfig holds the configuration for the server.
type ServerConfig struct {
	Addr string `value:"${addr:=0.0.0.0:9393}"`
}

// Server represents an HTTP server.
type Server struct {
	svr *http.Server
}

// NewServer creates a new Server instance with the specified configuration.
func NewServer(cfg ServerConfig) *Server {

	// Register pprof handlers for performance profiling.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /debug/pprof/", pprof.Index)
	mux.HandleFunc("GET /debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("GET /debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("GET /debug/pprof/trace", pprof.Trace)

	return &Server{svr: &http.Server{
		Addr:    cfg.Addr,
		Handler: mux,
	}}
}

// ListenAndServe starts the HTTP server and listens for incoming connections.
func (s *Server) ListenAndServe(sig gs.ReadySignal) error {
	ln, err := net.Listen("tcp", s.svr.Addr)
	if err != nil {
		return err
	}
	<-sig.TriggerAndWait()
	return s.svr.Serve(ln)
}

// Shutdown gracefully stops the server when called.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.svr.Shutdown(ctx)
}
