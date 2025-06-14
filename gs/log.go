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
	"path/filepath"
	"strings"

	"github.com/go-spring/log"
	"github.com/go-spring/spring-core/gs/internal/gs_conf"
)

// initLog initializes the log system.
func initLog() error {
	p, err := new(gs_conf.SysConfig).Refresh()
	if err != nil {
		return err
	}
	var c struct {
		LocalDir string `value:"${spring.app.config-local.dir:=./conf}"`
		Profiles string `value:"${spring.profiles.active:=}"`
	}
	if err = p.Bind(&c); err != nil {
		return err
	}
	logFile := "log.xml"
	if c.Profiles != "" {
		profile := strings.Split(c.Profiles, ",")[0]
		logFile = "log-" + profile + ".xml"
	}
	return log.RefreshFile(filepath.Join(c.LocalDir, logFile))
}
