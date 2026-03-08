/*
 * Copyright 2024 The Go-Spring Authors.
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

// Package gs_conf provides a layered configuration system for Go-Spring
// applications. It consolidates multiple configuration sources into a
// single immutable property set, supporting profile-specific files
// and optional import of additional configuration files.
//
// Supported configuration sources include:
//   - Built-in system defaults (SysConf)
//   - Local configuration files (e.g., ./conf/app.yaml)
//   - Remote configuration files (from config servers)
//   - Dynamically supplied remote properties
//   - Operating system environment variables
//   - Command-line arguments
//
// Sources are applied in a defined order; later sources override
// earlier ones when the same key is defined multiple times.
package gs_conf

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/stdlib/errutil"
	"github.com/go-spring/stdlib/flatten"
)

// AppConfig represents the layered configuration of an application.
// The typical merge order is:
//  1. System defaults (SysConf)
//  2. Local configuration files
//  3. Remote configuration files
//  4. Dynamically supplied remote properties
//  5. Environment variables
//  6. Command-line arguments
//
// Later layers override earlier ones in case of key conflicts.
type AppConfig struct {
	Properties *flatten.Properties
}

// NewAppConfig creates a new AppConfig instance.
func NewAppConfig() *AppConfig {
	return &AppConfig{
		Properties: flatten.NewProperties(nil),
	}
}

// Refresh refreshes the configuration by merging multiple sources.
func (c *AppConfig) Refresh() (flatten.Storage, error) {
	env := flatten.NewProperties(nil)
	cmd := flatten.NewProperties(nil)

	if err := NewEnvironment().CopyTo(env); err != nil {
		return nil, err
	}
	if err := NewCommandArgs().CopyTo(cmd); err != nil {
		return nil, err
	}

	l := &flatten.LayeredStorage{}
	l.AddStorage(flatten.StorageCommandLine, flatten.NewPropertiesStorage(cmd), "cmd")
	l.AddStorage(flatten.StorageEnvironment, flatten.NewPropertiesStorage(env), "env")
	l.AddStorage(flatten.StorageDefault, flatten.NewPropertiesStorage(c.Properties), "sys")

	confDir, err := conf.ResolveString(l, "${spring.app.config.dir:=./conf}")
	if err != nil {
		return nil, err
	}

	if err = loadFiles(l, confDir, nil); err != nil {
		return nil, errutil.Stack(err, "refresh error in source local")
	}

	strActiveProfiles, err := conf.ResolveString(l, "${spring.profiles.active:=}")
	if err != nil {
		return nil, err
	}
	activeProfiles := strings.Split(strActiveProfiles, ",")

	if err = loadFiles(l, confDir, activeProfiles); err != nil {
		return nil, errutil.Stack(err, "refresh error in source local")
	}
	return l, nil
}

// loadFiles loads all candidate configuration files in order and returns
// them as NamedPropertyCopier instances. Non-existent files are skipped,
// while other loading errors abort the process.
func loadFiles(l *flatten.LayeredStorage, dir string, activeProfiles []string) error {
	extensions := []string{".properties", ".yaml", ".yml", ".toml", ".tml", ".json"}

	var files []string
	if activeProfiles == nil {
		for _, ext := range extensions {
			files = append(files, filepath.Join(dir, "app"+ext))
		}
	} else {
		for _, s := range activeProfiles {
			for _, ext := range extensions {
				files = append(files, filepath.Join(dir, "app-"+s+ext))
			}
		}
	}

	for _, s := range files {
		// 解析文件名
		filename, err := conf.ResolveString(l, s)
		if err != nil {
			return err
		}

		// Load the file
		p, err := conf.Load(filename)
		if err != nil {
			// Don't use `os.IsNotExist`
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return err
		}

		// Load the file imports
		if err = loadFileImports(l, p, activeProfiles); err != nil {
			return err
		}

		// 优先级高的在前面，优先级低的在后面，因此是前插操作
		if activeProfiles == nil {
			l.AddStorage(flatten.StorageAppFile, flatten.NewPropertiesStorage(p), filename)
		} else {
			l.AddStorage(flatten.StorageProfileFile, flatten.NewPropertiesStorage(p), filename)
		}
	}
	return nil
}

func loadFileImports(l *flatten.LayeredStorage, p *flatten.Properties, activeProfiles []string) error {

	var i struct {
		Imports []string `value:"${spring.app.imports:=}"`
	}

	// 找到 file 里面的 imports
	if err := conf.Bind(flatten.NewPropertiesStorage(p), &i); err != nil {
		return err
	}

	// 没有则退出
	if len(i.Imports) == 0 {
		return nil
	}

	for _, source := range i.Imports {
		// 解析 source
		str, err := conf.ResolveString(l, source)
		if err != nil {
			return err
		}

		// 加载 source
		c, err := conf.Load(str)
		if err != nil {
			return err
		}

		// 优先级高的在前面，优先级低的在后面，因此是前插操作
		if activeProfiles == nil {
			l.AddStorage(flatten.StorageAppFile, flatten.NewPropertiesStorage(c), str)
		} else {
			l.AddStorage(flatten.StorageProfileFile, flatten.NewPropertiesStorage(c), str)
		}
	}
	return nil
}
