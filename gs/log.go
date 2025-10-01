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
	"github.com/go-spring/spring-base/util"
	"github.com/go-spring/spring-core/gs/internal/gs_conf"
)

// initLog initializes the application's logging system.
func initLog() error {

	// Refresh the global system configuration.
	p, err := new(gs_conf.SysConfig).Refresh()
	if err != nil {
		return util.FormatError(err, "refresh error in source sys")
	}

	var c struct {
		// LocalDir is the directory that contains configuration files.
		// Defaults to "./conf" if not provided.
		LocalDir string `value:"${spring.app.config-local.dir:=./conf}"`

		// Profiles specifies the active application profile(s),
		// such as "dev" or "prod".
		Profiles string `value:"${spring.profiles.active:=}"`
	}
	if err = p.Bind(&c); err != nil {
		return util.FormatError(err, "bind error in source sys")
	}

	var (
		logFileDefault = filepath.Join(c.LocalDir, "log.xml")
		logFileProfile string
	)

	// If one or more profiles are set, use the first profile to look
	// for a profile-specific log configuration file.
	if c.Profiles != "" {
		profile := strings.Split(c.Profiles, ",")[0]
		logFileProfile = filepath.Join(c.LocalDir, "log-"+profile+".xml")
	}

	// Determine which log configuration file to use.
	var logFile string
	for _, s := range []string{logFileProfile, logFileDefault} {
		if ok, err := util.PathExists(s); err != nil {
			return err
		} else if !ok {
			continue
		}
		logFile = s
		break
	}

	// If no configuration file exists, leave the logger as default.
	if logFile == "" {
		return nil
	}

	// Refresh the logger configuration from the selected file.
	return log.RefreshFile(logFile)
}
