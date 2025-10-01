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
// applications. It unifies multiple configuration sources and resolves them
// into a single immutable property set.
//
// The supported sources include:
//
//   - Built-in system defaults (SysConf)
//   - Local configuration files (e.g., ./conf/app.yaml)
//   - Remote configuration files (from config servers)
//   - Dynamically supplied remote properties
//   - Operating system environment variables
//   - Command-line arguments
//
// Sources are merged in a defined order so that later sources override
// properties from earlier ones. This enables flexible deployment patterns:
// defaults and packaged files supply baseline values, while environment
// variables and CLI options can easily override them in containerized or
// cloud-native environments.
//
// The package also supports profile-specific configuration files (e.g.,
// app-dev.yaml) and allows adding extra directories or files at runtime.
package gs_conf

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-spring/spring-base/util"
	"github.com/go-spring/spring-core/conf"
)

// osStat only for test.
var osStat = os.Stat

// SysConf is the global built-in configuration instance
// which usually holds the framework’s own default properties.
// It is loaded before any environment, file or command-line overrides.
var SysConf = conf.New()

// PropertyCopier defines the interface for any configuration source
// that can copy its key-value pairs into a target conf.MutableProperties.
type PropertyCopier interface {
	CopyTo(out *conf.MutableProperties) error
}

// NamedPropertyCopier is a wrapper around PropertyCopier that also
// carries a human-readable Name. The Name is used for logging,
// debugging or error reporting when merging multiple sources.
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

/******************************** SysConfig **********************************/

// SysConfig represents the init-level configuration layer
// composed of environment variables and command-line arguments.
type SysConfig struct {
	Environment *Environment // Environment variables as configuration source.
	CommandArgs *CommandArgs // Command-line arguments as configuration source.
}

// Refresh collects properties from the system configuration sources
// (built-in SysConf, environment variables, and command-line arguments)
// and merges them into a single immutable conf.Properties.
func (c *SysConfig) Refresh() (conf.Properties, error) {
	return merge(
		NewNamedPropertyCopier("sys", SysConf),
		NewNamedPropertyCopier("env", c.Environment),
		NewNamedPropertyCopier("cmd", c.CommandArgs),
	)
}

/******************************** AppConfig **********************************/

// AppConfig represents a layered configuration for the application runtime.
// The layers, in their merge order, typically include:
//
//  1. System defaults (SysConf)
//  2. Local configuration files
//  3. Remote configuration files
//  4. Dynamically supplied remote properties
//  5. Environment variables
//  6. Command-line arguments
//
// Layers appearing later in the list override earlier ones when keys conflict.
type AppConfig struct {
	LocalFile   *PropertySources // Configuration sources from local files.
	RemoteFile  *PropertySources // Configuration sources from remote files.
	RemoteProp  conf.Properties  // Properties fetched from a remote server.
	Environment *Environment     // Environment variables as configuration source.
	CommandArgs *CommandArgs     // Command-line arguments as configuration source.
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

// merge combines multiple NamedPropertyCopier instances into a single
// conf.Properties. The sources are applied in order; properties from
// later sources override earlier ones. If any source fails to copy,
// the merge aborts and returns an error indicating the failing source.
func merge(sources ...*NamedPropertyCopier) (conf.Properties, error) {
	out := conf.New()
	for _, s := range sources {
		if s != nil {
			if err := s.CopyTo(out); err != nil {
				return nil, util.WrapError(err, "merge error in source %s", s.Name)
			}
		}
	}
	return out, nil
}

// Refresh merges all layers of configurations into a read-only properties.
func (c *AppConfig) Refresh() (conf.Properties, error) {
	p, err := new(SysConfig).Refresh()
	if err != nil {
		return nil, util.WrapError(err, "refresh error in source sys")
	}

	localFiles, err := c.LocalFile.loadFiles(p)
	if err != nil {
		return nil, util.WrapError(err, "refresh error in source local")
	}

	remoteFiles, err := c.RemoteFile.loadFiles(p)
	if err != nil {
		return nil, util.WrapError(err, "refresh error in source remote")
	}

	var sources []*NamedPropertyCopier
	sources = append(sources, NewNamedPropertyCopier("sys", SysConf))
	sources = append(sources, localFiles...)
	sources = append(sources, remoteFiles...)
	sources = append(sources, NewNamedPropertyCopier("remote", c.RemoteProp))
	sources = append(sources, NewNamedPropertyCopier("env", c.Environment))
	sources = append(sources, NewNamedPropertyCopier("cmd", c.CommandArgs))
	return merge(sources...)
}

/******************************** BootConfig *********************************/

// BootConfig represents a layered configuration used during application boot.
// It typically includes only system, local file, environment and command-line
// sources — no remote sources.
type BootConfig struct {
	LocalFile   *PropertySources // Configuration sources from local files.
	Environment *Environment     // Environment variables as configuration source.
	CommandArgs *CommandArgs     // Command-line arguments as configuration source.
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
	p, err := new(SysConfig).Refresh()
	if err != nil {
		return nil, util.WrapError(err, "refresh error in source sys")
	}

	localFiles, err := c.LocalFile.loadFiles(p)
	if err != nil {
		return nil, util.WrapError(err, "refresh error in source local")
	}

	var sources []*NamedPropertyCopier
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

// PropertySources represents a collection of configuration files
// associated with a particular configuration type and logical name.
// It supports both default directories and additional user-supplied
// directories or files.
type PropertySources struct {
	configType ConfigType // Type of the configuration (local or remote).
	configName string     // Base name of the configuration files.
	extraDirs  []string   // Extra directories to search for configuration files.
	extraFiles []string   // Extra individual files to include.
}

// NewPropertySources creates a new instance of PropertySources.
func NewPropertySources(configType ConfigType, configName string) *PropertySources {
	return &PropertySources{
		configType: configType,
		configName: configName,
	}
}

// Reset clears all previously added extra directories and files.
func (p *PropertySources) Reset() {
	p.extraFiles = nil
	p.extraDirs = nil
}

// AddDir registers one or more additional directories to search for
// configuration files. Non-existent directories are silently ignored,
// but if the path exists and is not a directory, it panics.
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
			panic(util.FormatError(nil, "should be a directory %s", d))
		}
	}
	p.extraDirs = append(p.extraDirs, dirs...)
}

// AddFile registers one or more additional configuration files.
// Non-existent files are silently ignored, but if the path exists
// and is a directory, it panics.
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
			panic(util.FormatError(nil, "should be a file %s", f))
		}
	}
	p.extraFiles = append(p.extraFiles, files...)
}

// getDefaultDir determines the default configuration directory
// according to the configuration type and current resolved properties.
func (p *PropertySources) getDefaultDir(resolver conf.Properties) (string, error) {
	switch p.configType {
	case ConfigTypeLocal:
		return resolver.Resolve("${spring.app.config-local.dir:=./conf}")
	case ConfigTypeRemote:
		return resolver.Resolve("${spring.app.config-remote.dir:=./conf/remote}")
	default:
		return "", util.FormatError(nil, "unknown config type: %s", p.configType)
	}
}

// getFiles generates the list of configuration file paths to try,
// including both the base config name and profile-specific variants.
// For example, with profile "dev", it will try "app-dev.yaml" etc.
func (p *PropertySources) getFiles(dir string, resolver conf.Properties) ([]string, error) {
	extensions := []string{".properties", ".yaml", ".yml", ".toml", ".tml", ".json"}

	var files []string
	for _, ext := range extensions {
		files = append(files, filepath.Join(dir, p.configName+ext))
	}

	activeProfiles, err := resolver.Resolve("${spring.profiles.active:=}")
	if err != nil {
		return nil, err
	}

	if activeProfiles = strings.TrimSpace(activeProfiles); activeProfiles != "" {
		for s := range strings.SplitSeq(activeProfiles, ",") {
			if s = strings.TrimSpace(s); s != "" {
				for _, ext := range extensions {
					files = append(files, filepath.Join(dir, p.configName+"-"+s+ext))
				}
			}
		}
	}
	return files, nil
}

// loadFiles loads all candidate configuration files in order and wraps
// successfully loaded ones as NamedPropertyCopier. Non-existent files
// are skipped silently, while other loading errors abort the process.
func (p *PropertySources) loadFiles(resolver conf.Properties) ([]*NamedPropertyCopier, error) {
	defaultDir, err := p.getDefaultDir(resolver)
	if err != nil {
		return nil, err
	}
	dirs := append([]string{defaultDir}, p.extraDirs...)

	var files []string
	for _, dir := range dirs {
		temp, err := p.getFiles(dir, resolver)
		if err != nil {
			return nil, err
		}
		files = append(files, temp...)
	}
	files = append(files, p.extraFiles...)

	var ret []*NamedPropertyCopier
	for _, s := range files {
		filename, err := resolver.Resolve(s)
		if err != nil {
			return nil, err
		}
		c, err := conf.Load(filename)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, err
		}
		ret = append(ret, NewNamedPropertyCopier(filename, c))
	}
	return ret, nil
}
