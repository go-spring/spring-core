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

/*
Package gs_conf provides hierarchical configuration management
with multi-source support for Go-Spring framework.

Key Features:

1. Command-line argument parsing
  - Supports `-D key[=value]` format arguments
  - Customizable prefix via `GS_ARGS_PREFIX` environment variable
  - Example: `./app -D server.port=8080 -D debug`

2. Environment variable handling
  - Automatic loading of `GS_` prefixed variables
  - Conversion rules: `GS_DB_HOST=127.0.0.1` → `db.host=127.0.0.1`
  - Direct mapping of non-prefixed environment variables

3. Configuration file management
  - Supports properties/yaml/toml/json formats
  - Local configurations: ./conf/app.{properties|yaml|toml|json}
  - Remote configurations: ./conf/remote/app.{properties|yaml|toml|json}
  - Profile-based configurations (e.g., app-dev.properties)

4. Layered configuration hierarchy
  - Priority order: System config → File config → Env variables → CLI arguments
  - Provides AppConfig (application context) and BootConfig (boot context)
  - High-priority configurations override lower ones
*/
package gs_conf

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-spring/spring-core/conf"
)

// osStat only for test.
var osStat = os.Stat

// SysConf is the builtin configuration.
var SysConf = conf.New()

// PropertyCopier defines the interface for copying properties.
type PropertyCopier interface {
	CopyTo(out *conf.MutableProperties) error
}

// NamedPropertyCopier defines the interface for copying properties with a name.
type NamedPropertyCopier struct {
	PropertyCopier
	Name string
}

// NewNamedPropertyCopier creates a new instance of NamedPropertyCopier.
func NewNamedPropertyCopier(name string, p PropertyCopier) *NamedPropertyCopier {
	return &NamedPropertyCopier{PropertyCopier: p, Name: name}
}

func (c *NamedPropertyCopier) CopyTo(out *conf.MutableProperties) error {
	if c.PropertyCopier != nil {
		return c.PropertyCopier.CopyTo(out)
	}
	return nil
}

/******************************** AppConfig **********************************/

// AppConfig represents a layered application configuration.
type AppConfig struct {
	LocalFile   *PropertySources // Configuration sources from local files.
	RemoteFile  *PropertySources // Configuration sources from remote files.
	RemoteProp  conf.Properties  // Remote properties.
	Environment *Environment     // Environment variables as configuration source.
	CommandArgs *CommandArgs     // Command line arguments as configuration source.
}

// NewAppConfig creates a new instance of AppConfig.
func NewAppConfig() *AppConfig {
	return &AppConfig{
		LocalFile:   NewPropertySources(ConfigTypeLocal, "app"),
		RemoteFile:  NewPropertySources(ConfigTypeRemote, "app"),
		Environment: NewEnvironment(),
		CommandArgs: NewCommandArgs(),
	}
}

func merge(sources ...PropertyCopier) (conf.Properties, error) {
	out := conf.New()
	for _, s := range sources {
		if s != nil {
			if err := s.CopyTo(out); err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}

// Refresh merges all layers of configurations into a read-only properties.
func (c *AppConfig) Refresh() (conf.Properties, error) {
	p, err := merge(
		NewNamedPropertyCopier("sys", SysConf),
		NewNamedPropertyCopier("env", c.Environment),
		NewNamedPropertyCopier("cmd", c.CommandArgs),
	)
	if err != nil {
		return nil, err
	}

	localFiles, err := c.LocalFile.loadFiles(p)
	if err != nil {
		return nil, err
	}

	remoteFiles, err := c.RemoteFile.loadFiles(p)
	if err != nil {
		return nil, err
	}

	var sources []PropertyCopier
	sources = append(sources, NewNamedPropertyCopier("sys", SysConf))
	sources = append(sources, localFiles...)
	sources = append(sources, remoteFiles...)
	sources = append(sources, NewNamedPropertyCopier("remote", c.RemoteProp))
	sources = append(sources, NewNamedPropertyCopier("env", c.Environment))
	sources = append(sources, NewNamedPropertyCopier("cmd", c.CommandArgs))

	return merge(sources...)
}

/******************************** BootConfig *********************************/

// BootConfig represents a layered boot configuration.
type BootConfig struct {
	LocalFile   *PropertySources // Configuration sources from local files.
	Environment *Environment     // Environment variables as configuration source.
	CommandArgs *CommandArgs     // Command line arguments as configuration source.
}

// NewBootConfig creates a new instance of BootConfig.
func NewBootConfig() *BootConfig {
	return &BootConfig{
		LocalFile:   NewPropertySources(ConfigTypeLocal, "boot"),
		Environment: NewEnvironment(),
		CommandArgs: NewCommandArgs(),
	}
}

// Refresh merges all layers of configurations into a read-only properties.
func (c *BootConfig) Refresh() (conf.Properties, error) {
	p, err := merge(
		NewNamedPropertyCopier("sys", SysConf),
		NewNamedPropertyCopier("env", c.Environment),
		NewNamedPropertyCopier("cmd", c.CommandArgs),
	)
	if err != nil {
		return nil, err
	}

	localFiles, err := c.LocalFile.loadFiles(p)
	if err != nil {
		return nil, err
	}

	var sources []PropertyCopier
	sources = append(sources, NewNamedPropertyCopier("sys", SysConf))
	sources = append(sources, localFiles...)
	sources = append(sources, NewNamedPropertyCopier("env", c.Environment))
	sources = append(sources, NewNamedPropertyCopier("cmd", c.CommandArgs))

	return merge(sources...)
}

/****************************** PropertySources ******************************/

// ConfigType defines the type of configuration: local or remote.
type ConfigType string

const (
	ConfigTypeLocal  ConfigType = "local"
	ConfigTypeRemote ConfigType = "remote"
)

// PropertySources is a collection of configuration files.
type PropertySources struct {
	configType ConfigType // Type of the configuration (local or remote).
	configName string     // Name of the configuration.
	extraDirs  []string   // Extra directories to be included in the configuration.
	extraFiles []string   // Extra files to be included in the configuration.
}

// NewPropertySources creates a new instance of PropertySources.
func NewPropertySources(configType ConfigType, configName string) *PropertySources {
	return &PropertySources{
		configType: configType,
		configName: configName,
	}
}

// Reset resets all the extra files.
func (p *PropertySources) Reset() {
	p.extraFiles = nil
	p.extraDirs = nil
}

// AddDir adds a or more than one extra directories.
func (p *PropertySources) AddDir(dirs ...string) {
	for _, d := range dirs {
		info, err := osStat(d)
		if err != nil {
			if !os.IsNotExist(err) {
				panic(err)
			}
			continue
		}
		if !info.IsDir() {
			panic("should be a directory")
		}
	}
	p.extraDirs = append(p.extraDirs, dirs...)
}

// AddFile adds a or more than one extra files.
func (p *PropertySources) AddFile(files ...string) {
	for _, f := range files {
		info, err := osStat(f)
		if err != nil {
			if !os.IsNotExist(err) {
				panic(err)
			}
			continue
		}
		if info.IsDir() {
			panic("should be a file")
		}
	}
	p.extraFiles = append(p.extraFiles, files...)
}

// getDefaultDir returns the default configuration directory based on the configuration type.
func (p *PropertySources) getDefaultDir(resolver conf.Properties) (configDir string, err error) {
	switch p.configType {
	case ConfigTypeLocal:
		return resolver.Resolve("${spring.app.config-local.dir:=./conf}")
	case ConfigTypeRemote:
		return resolver.Resolve("${spring.app.config-remote.dir:=./conf/remote}")
	default:
		return "", fmt.Errorf("unknown config type: %s", p.configType)
	}
}

// getFiles returns the list of configuration files based on the configuration directory and active profiles.
func (p *PropertySources) getFiles(dir string, resolver conf.Properties) (files []string, err error) {
	extensions := []string{".properties", ".yaml", ".yml", ".toml", ".tml", ".json"}

	for _, ext := range extensions {
		files = append(files, fmt.Sprintf("%s/%s%s", dir, p.configName, ext))
	}

	activeProfiles, err := resolver.Resolve("${spring.profiles.active:=}")
	if err != nil {
		return nil, err
	}

	if activeProfiles = strings.TrimSpace(activeProfiles); activeProfiles != "" {
		for s := range strings.SplitSeq(activeProfiles, ",") {
			if s = strings.TrimSpace(s); s != "" {
				for _, ext := range extensions {
					files = append(files, fmt.Sprintf("%s/%s-%s%s", dir, p.configName, s, ext))
				}
			}
		}
	}
	return files, nil
}

// loadFiles loads all configuration files and returns them as a list of Properties.
func (p *PropertySources) loadFiles(resolver conf.Properties) ([]PropertyCopier, error) {

	defaultDir, err := p.getDefaultDir(resolver)
	if err != nil {
		return nil, err
	}
	dirs := append([]string{defaultDir}, p.extraDirs...)

	var files []string
	for _, dir := range dirs {
		var temp []string
		temp, err = p.getFiles(dir, resolver)
		if err != nil {
			return nil, err
		}
		files = append(files, temp...)
	}
	files = append(files, p.extraFiles...)

	var ret []PropertyCopier
	for _, s := range files {
		filename, err := resolver.Resolve(s)
		if err != nil {
			return nil, err
		}
		c, err := conf.Load(filename)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		ret = append(ret, NewNamedPropertyCopier(filename, c))
	}
	return ret, nil
}
