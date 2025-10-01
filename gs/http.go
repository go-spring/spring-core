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
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/go-spring/spring-base/util"
	"github.com/go-spring/spring-core/conf"
)

func init() {
	Module([]ConditionOnProperty{
		OnEnableServers(),
		OnProperty(EnableSimpleHttpServerProp).HavingValue("true").MatchIfMissing(),
	}, func(p conf.Properties) error {

		// Register the default HTTP multiplexer as a bean
		// if no other http.Handler bean has been defined.
		Provide(http.DefaultServeMux).
			Export(As[http.Handler]()).
			Condition(OnMissingBean[http.Handler]())

		// Provide a new SimpleHttpServer instance with
		// http.Handler injection and configuration binding.
		Provide(NewSimpleHttpServer).AsServer()

		return nil
	})
}

// SimpleHttpServerConfig holds configuration for the SimpleHttpServer.
type SimpleHttpServerConfig struct {
	// Address specifies the TCP address the server listens on.
	// Example: ":9090" (listen on all interfaces, port 9090).
	Address string `value:"${http.server.addr:=:9090}"`

	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body.
	ReadTimeout time.Duration `value:"${http.server.readTimeout:=5s}"`

	// HeaderTimeout is the maximum duration for reading request headers.
	HeaderTimeout time.Duration `value:"${http.server.headerTimeout:=1s}"`

	// WriteTimeout is the maximum duration before timing out
	// a response write.
	WriteTimeout time.Duration `value:"${http.server.writeTimeout:=5s}"`

	// IdleTimeout is the maximum amount of time to wait for
	// the next request when keep-alive connections are enabled.
	IdleTimeout time.Duration `value:"${http.server.idleTimeout:=60s}"`
}

// SimpleHttpServer wraps a standard [http.Server] to integrate
// it into the Go-Spring application lifecycle.
type SimpleHttpServer struct {
	svr *http.Server // The HTTP server instance.
}

// NewSimpleHttpServer constructs a new SimpleHttpServer using
// the provided HTTP handler and configuration.
func NewSimpleHttpServer(h http.Handler, cfg SimpleHttpServerConfig) *SimpleHttpServer {
	return &SimpleHttpServer{svr: &http.Server{
		Addr:         cfg.Address,
		Handler:      h,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}}
}

// ListenAndServe starts the HTTP server and blocks until it is stopped.
// It waits for the given ReadySignal to be triggered before accepting traffic.
func (s *SimpleHttpServer) ListenAndServe(sig ReadySignal) error {
	ln, err := net.Listen("tcp", s.svr.Addr)
	if err != nil {
		return util.FormatError(err, "failed to listen on %s", s.svr.Addr)
	}
	<-sig.TriggerAndWait()
	err = s.svr.Serve(ln)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return util.FormatError(err, "failed to serve on %s", s.svr.Addr)
}

// Shutdown gracefully stops the HTTP server using the provided context,
// allowing in-flight requests to complete before closing.
func (s *SimpleHttpServer) Shutdown(ctx context.Context) error {
	return s.svr.Shutdown(ctx)
}
