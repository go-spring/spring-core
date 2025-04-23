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
	"net"
	"net/http"
	"time"
)

func init() {
	Provide(
		NewSimpleHttpServer,
		IndexArg(1, BindArg(SetHttpServerAddr, TagArg("${http.server.addr:=0.0.0.0:9090}"))),
		IndexArg(1, BindArg(SetHttpServerReadTimeout, TagArg("${http.server.readTimeout:=5s}"))),
		IndexArg(1, BindArg(SetHttpServerWriteTimeout, TagArg("${http.server.writeTimeout:=5s}"))),
	).Condition(
		OnBean[*http.ServeMux](),
		OnProperty(EnableServersProp).HavingValue("true").MatchIfMissing(),
		OnProperty(EnableSimpleHttpServerProp).HavingValue("true").MatchIfMissing(),
	).AsServer()
}

// HttpServerConfig holds configuration options for the HTTP server.
type HttpServerConfig struct {
	Address      string        // The address to bind the server to.
	ReadTimeout  time.Duration // The read timeout duration.
	WriteTimeout time.Duration // The write timeout duration.
}

// HttpServerOption is a function type for setting options on HttpServerConfig.
type HttpServerOption func(arg *HttpServerConfig)

// SetHttpServerAddr sets the address of the HTTP server.
func SetHttpServerAddr(addr string) HttpServerOption {
	return func(arg *HttpServerConfig) {
		arg.Address = addr
	}
}

// SetHttpServerReadTimeout sets the read timeout for the HTTP server.
func SetHttpServerReadTimeout(timeout time.Duration) HttpServerOption {
	return func(arg *HttpServerConfig) {
		arg.ReadTimeout = timeout
	}
}

// SetHttpServerWriteTimeout sets the write timeout for the HTTP server.
func SetHttpServerWriteTimeout(timeout time.Duration) HttpServerOption {
	return func(arg *HttpServerConfig) {
		arg.WriteTimeout = timeout
	}
}

// SimpleHttpServer wraps a [http.Server] instance.
type SimpleHttpServer struct {
	svr *http.Server // The HTTP server instance.
}

// NewSimpleHttpServer creates a new instance of SimpleHttpServer.
func NewSimpleHttpServer(mux *http.ServeMux, opts ...HttpServerOption) *SimpleHttpServer {
	arg := &HttpServerConfig{
		Address:      "0.0.0.0:9090",
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
	}
	for _, opt := range opts {
		opt(arg)
	}
	return &SimpleHttpServer{svr: &http.Server{
		Addr:         arg.Address,
		Handler:      mux,
		ReadTimeout:  arg.ReadTimeout,
		WriteTimeout: arg.WriteTimeout,
	}}
}

// ListenAndServe starts the HTTP server and listens for incoming connections.
func (s *SimpleHttpServer) ListenAndServe(sig ReadySignal) error {
	ln, err := net.Listen("tcp", s.svr.Addr)
	if err != nil {
		return err
	}
	<-sig.TriggerAndWait()
	return s.svr.Serve(ln)
}

// Shutdown gracefully shuts down the HTTP server with the given context.
func (s *SimpleHttpServer) Shutdown(ctx context.Context) error {
	return s.svr.Shutdown(ctx)
}
