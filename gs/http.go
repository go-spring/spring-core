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
)

func init() {
	// Initialize the HTTP server. The server will listen on the address specified
	// by the 'server.addr' configuration, defaulting to "0.0.0.0:9090" if not set.
	// It is only provided as a server if an instance of *http.ServeMux exists.
	Provide(NewSimpleHttpServer, TagArg("${server.addr:=0.0.0.0:9090}")).Condition(
		OnBean(BeanSelectorFor[*http.ServeMux]()),
	).AsServer()
}

// SimpleHttpServer wraps a [http.Server] instance.
type SimpleHttpServer struct {
	svr *http.Server
}

// NewSimpleHttpServer creates a new instance of SimpleHttpServer.
func NewSimpleHttpServer(addr string, mux *http.ServeMux) *SimpleHttpServer {
	return &SimpleHttpServer{svr: &http.Server{
		Addr:    addr,
		Handler: mux,
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
