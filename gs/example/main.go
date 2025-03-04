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
	"os"

	"github.com/go-spring/spring-core/gs"
	"github.com/go-spring/spring-core/util/syslog"

	_ "github.com/go-spring/spring-core/gs/example/app"
	_ "github.com/go-spring/spring-core/gs/example/biz"
	_ "github.com/go-spring/spring-core/gs/example/idl"
)

func main() {
	_ = os.Unsetenv("_")
	_ = os.Unsetenv("TERM")
	_ = os.Unsetenv("TERM_SESSION_ID")
	if err := gs.Run(); err != nil {
		syslog.Errorf("app run failed: %s", err.Error())
	}
}
