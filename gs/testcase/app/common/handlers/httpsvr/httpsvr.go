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

package httpsvr

import (
	"context"
	"net/http"

	"github.com/go-spring/spring-core/gs"
	"github.com/go-spring/spring-core/gs/testcase/app/controller"
)

func init() {
	gs.Server(NewServer, gs.TagArg("${server}"))
}

// ServerConfig ...
type ServerConfig struct {
	Addr string `value:"${addr}"`
}

// Server ...
type Server struct {
	svr *http.Server
	mux *http.ServeMux
}

// NewServer ...
func NewServer(cfg ServerConfig) *Server {
	mux := http.NewServeMux()
	svr := &http.Server{
		Addr:    cfg.Addr,
		Handler: mux,
	}
	return &Server{svr: svr, mux: mux}
}

func (s *Server) OnBeanInit(ctx gs.Context) error {
	return ctx.Run(func(c *controller.Controller) {
		s.mux.HandleFunc("GET /books", c.ListBooks)
		s.mux.HandleFunc("GET /books/{id}", c.GetBook)
		s.mux.HandleFunc("POST /books", c.SaveBook)
		s.mux.HandleFunc("DELETE /books/{id}", c.DeleteBook)
	})
}

func (s *Server) Serve() error {
	return s.svr.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.svr.Shutdown(ctx)
}
