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
	"context"
	"path/filepath"
	"strings"

	"github.com/go-spring/log"
	"github.com/go-spring/spring-base/util"
	"github.com/go-spring/spring-core/gs/internal/gs_conf"
)

// initLog initializes the application's logging system.
func initLog() error {

	// Step 1: Refresh the global system configuration.
	p, err := new(gs_conf.SysConfig).Refresh()
	if err != nil {
		return util.FormatError(err, "refresh error in source sys")
	}

	// Step 2: Load logging-related configuration parameters.
	var c struct {
		// LocalDir is the directory that contains configuration files.
		// Defaults to "./conf" if not provided.
		LocalDir string `value:"${spring.app.config-local.dir:=./conf}"`

		// Profiles specifies the active application profile(s),
		// such as "dev", "prod", etc.
		// Multiple profiles can be provided as a comma-separated list.
		Profiles string `value:"${spring.profiles.active:=}"`
	}
	if err = p.Bind(&c); err != nil {
		return util.FormatError(err, "bind error in source sys")
	}

	extensions := []string{".properties", ".yaml", ".yml", ".xml", ".json"}

	// Step 3: Build a list of candidate configuration files.
	var files []string
	if profiles := strings.TrimSpace(c.Profiles); profiles != "" {
		for s := range strings.SplitSeq(profiles, ",") { // NOTE: range returns index
			if s = strings.TrimSpace(s); s != "" {
				for _, ext := range extensions {
					files = append(files, filepath.Join(c.LocalDir, "log-"+s+ext))
				}
			}
		}
	}
	for _, ext := range extensions {
		files = append(files, filepath.Join(c.LocalDir, "log"+ext))
	}

	// Step 4: Detect existing configuration files.
	var logFiles []string
	for _, s := range files {
		if ok, err := util.PathExists(s); err != nil {
			return err
		} else if ok {
			logFiles = append(logFiles, s)
		}
	}

	// Step 5: Apply logging configuration or fall back to defaults.
	switch n := len(logFiles); {
	case n == 0:
		log.Infof(context.Background(), log.TagAppDef, "no log configuration file found, using default logger")
		return nil
	case n > 1:
		return util.FormatError(nil, "multiple log files found: %s", logFiles)
	default:
		return log.RefreshFile(logFiles[0])
	}
}
