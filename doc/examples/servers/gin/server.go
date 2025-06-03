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
	"context"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-spring/spring-core/gs"
)

func init() {
	gs.Provide(NewSimpleGinServer, gs.TagArg(""), gs.TagArg("${http.server}")).AsServer()
}

type HttpServerConfig struct {
	Address      string        `value:"${addr:=0.0.0.0:9090}"`
	ReadTimeout  time.Duration `value:"${readTimeout:=5s}"`
	WriteTimeout time.Duration `value:"${writeTimeout:=5s}"`
}

type SimpleGinServer struct {
	svr *http.Server
}

func NewSimpleGinServer(e *gin.Engine, cfg HttpServerConfig) *SimpleGinServer {
	return &SimpleGinServer{svr: &http.Server{
		Addr:         cfg.Address,
		Handler:      e,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}}
}

func (s *SimpleGinServer) ListenAndServe(sig gs.ReadySignal) error {
	ln, err := net.Listen("tcp", s.svr.Addr)
	if err != nil {
		return err
	}
	<-sig.TriggerAndWait()
	return s.svr.Serve(ln)
}

func (s *SimpleGinServer) Shutdown(ctx context.Context) error {
	return s.svr.Shutdown(ctx)
}
