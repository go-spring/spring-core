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
	"os"
	"path/filepath"
	"strings"

	"github.com/go-spring/log"
	"github.com/go-spring/spring-core/gs/internal/gs_conf"
)

var logRefreshed bool

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
	var (
		logFileDefault string
		logFileProfile string
	)
	logFileDefault = filepath.Join(c.LocalDir, "log.xml")
	if c.Profiles != "" {
		profile := strings.Split(c.Profiles, ",")[0]
		logFileProfile = filepath.Join(c.LocalDir, "log-"+profile+".xml")
	}
	var logFile string
	for _, s := range []string{logFileProfile, logFileDefault} {
		if _, err = os.Stat(s); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		logFile = s
		break
	}
	if logFile == "" { // no log file exists
		return nil
	}
	if err = log.RefreshFile(logFile); err != nil {
		return err
	}
	logRefreshed = true
	return nil
}
