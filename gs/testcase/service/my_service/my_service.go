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

package my_service

import (
	"github.com/go-spring/spring-core/gs"
	"github.com/go-spring/spring-core/gs/testcase/model/my_model_a"
	"github.com/go-spring/spring-core/gs/testcase/model/my_model_b"
)

func init() {
	gs.Provide(NewService, gs.ValueArg("svr_test"))
}

type Service struct {
	SvrName string
	AppName string             `value:"${spring.app.name:=test}"`
	ModelA  *my_model_a.ModelA `autowire:""`
	ModelB  *my_model_b.ModelB `autowire:""`
	//DyncValue gs.Dync[int64]     `value:"${spring.app.name:=test}"`
}

func NewService(svrName string) *Service {
	return &Service{SvrName: svrName}
}
