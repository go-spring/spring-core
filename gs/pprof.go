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
	"net/http"
	"net/http/pprof"
)

func init() {
	Register(
		NewBean(
			NewSimplePProfServer,
			TagArg("${pprof.server.addr:=0.0.0.0:9981}"),
		),
	).Condition(
		OnProperty(EnableSimplePProfServerProp).HavingValue("true"),
	).AsServer()
}

// SimplePProfServer is a simple pprof server.
type SimplePProfServer struct {
	*SimpleHttpServer
}

// NewSimplePProfServer creates a new SimplePProfServer.
func NewSimplePProfServer(addr string) *SimplePProfServer {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /debug/pprof/", pprof.Index)
	mux.HandleFunc("GET /debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("GET /debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("GET /debug/pprof/trace", pprof.Trace)
	return &SimplePProfServer{
		SimpleHttpServer: NewSimpleHttpServer(mux, SetHttpServerAddr(addr)),
	}
}
