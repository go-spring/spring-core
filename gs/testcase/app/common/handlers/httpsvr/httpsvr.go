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
	"fmt"
	"net/http"

	"github.com/go-spring/spring-core/gs"
	"github.com/go-spring/spring-core/gs/testcase/app/controller"
	"github.com/go-spring/spring-core/util/syslog"
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

func (s *Server) OnAppStart(ctx *gs.AppContext) {

	var c *controller.Controller
	if err := ctx.Unsafe().Get(&c); err != nil {
		gs.ShutDown(fmt.Sprintf("get controller error: %v", err))
		return
	}

	s.mux.HandleFunc("GET /books", c.ListBooks)
	s.mux.HandleFunc("GET /books/{id}", c.GetBook)
	s.mux.HandleFunc("POST /books", c.SaveBook)
	s.mux.HandleFunc("DELETE /books/{id}", c.DeleteBook)

	ctx.Go(func(ctx context.Context) {
		if err := s.svr.ListenAndServe(); err != nil {
			gs.ShutDown(fmt.Sprintf("server listen error: %v", err))
		}
	})
}

func (s *Server) OnAppStop(ctx context.Context) {
	if err := s.svr.Shutdown(ctx); err != nil {
		syslog.Errorf("shutdown server failed: %s", err.Error())
	}
}
