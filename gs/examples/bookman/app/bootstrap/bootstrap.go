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

package bootstrap

import (
	"os"

	"github.com/go-spring/spring-core/gs"
)

func init() {
	gs.Boot().Object(&Runner{}).AsRunner().OnProfiles("online")
}

type Runner struct{}

func (r *Runner) Run() error {
	err := os.MkdirAll("./conf", os.ModePerm)
	if err != nil {
		return err
	}

	const data = `
server.addr=0.0.0.0:9090

log.access.name=access.log
log.access.dir=./log

log.biz.name=biz.log
log.biz.dir=./log

log.dao.name=dao.log
log.dao.dir=./log`

	const file = "conf/app-online.properties"
	return os.WriteFile(file, []byte(data), os.ModePerm)
}
