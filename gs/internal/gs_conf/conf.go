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

package gs_conf

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/util/sysconf"
)

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

func merge(out *conf.MutableProperties, sources ...interface {
	CopyTo(out *conf.MutableProperties) error
}) error {
	for _, s := range sources {
		if s != nil {
			if err := s.CopyTo(out); err != nil {
				return err
			}
		}
	}
	return nil
}

// Refresh merges all layers of configurations into a read-only properties.
func (c *AppConfig) Refresh() (conf.Properties, error) {
	p := sysconf.Clone()
	err := merge(p, c.Environment, c.CommandArgs)
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

	var sources []interface {
		CopyTo(out *conf.MutableProperties) error
	}
	for _, file := range localFiles {
		sources = append(sources, file)
	}
	for _, file := range remoteFiles {
		sources = append(sources, file)
	}
	sources = append(sources, c.RemoteProp)
	sources = append(sources, c.Environment)
	sources = append(sources, c.CommandArgs)

	p = sysconf.Clone()
	err = merge(p, sources...)
	if err != nil {
		return nil, err
	}
	return p, nil
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

	p := sysconf.Clone()
	err := merge(p, c.Environment, c.CommandArgs)
	if err != nil {
		return nil, err
	}

	localFiles, err := c.LocalFile.loadFiles(p)
	if err != nil {
		return nil, err
	}

	var sources []interface {
		CopyTo(out *conf.MutableProperties) error
	}
	for _, file := range localFiles {
		sources = append(sources, file)
	}
	sources = append(sources, c.Environment)
	sources = append(sources, c.CommandArgs)

	p = sysconf.Clone()
	err = merge(p, sources...)
	if err != nil {
		return nil, err
	}
	return p, nil
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
}

// AddDir adds a or more than one extra directories.
func (p *PropertySources) AddDir(dirs ...string) {
	for _, d := range dirs {
		info, err := os.Stat(d)
		if err != nil {
			panic(err)
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
		info, err := os.Stat(f)
		if err != nil {
			panic(err)
		}
		if info.IsDir() {
			panic("should be a file")
		}
	}
	p.extraFiles = append(p.extraFiles, files...)
}

// getDefaultDir returns the default configuration directory based on the configuration type.
func (p *PropertySources) getDefaultDir(resolver conf.Properties) (configDir string, err error) {
	if p.configType == ConfigTypeLocal {
		return resolver.Resolve("${spring.app.config.dir:=./conf}")
	} else if p.configType == ConfigTypeRemote {
		return resolver.Resolve("${spring.cloud.config.dir:=./conf/remote}")
	} else {
		return "", fmt.Errorf("unknown config type: %s", p.configType)
	}
}

// getFiles returns the list of configuration files based on the configuration directory and active profiles.
func (p *PropertySources) getFiles(dir string, resolver conf.Properties) (_ []string, err error) {

	files := []string{
		fmt.Sprintf("%s/%s.properties", dir, p.configName),
		fmt.Sprintf("%s/%s.yaml", dir, p.configName),
		fmt.Sprintf("%s/%s.toml", dir, p.configName),
		fmt.Sprintf("%s/%s.json", dir, p.configName),
	}

	activeProfiles, err := resolver.Resolve("${spring.profiles.active:=}")
	if err != nil {
		return nil, err
	}
	if activeProfiles = strings.TrimSpace(activeProfiles); activeProfiles != "" {
		ss := strings.Split(activeProfiles, ",")
		for _, s := range ss {
			if s = strings.TrimSpace(s); s != "" {
				files = append(files, []string{
					fmt.Sprintf("%s/%s-%s.properties", dir, p.configName, s),
					fmt.Sprintf("%s/%s-%s.yaml", dir, p.configName, s),
					fmt.Sprintf("%s/%s-%s.toml", dir, p.configName, s),
					fmt.Sprintf("%s/%s-%s.json", dir, p.configName, s),
				}...)
			}
		}
	}
	return files, nil
}

// loadFiles loads all configuration files and returns them as a list of Properties.
func (p *PropertySources) loadFiles(resolver conf.Properties) ([]conf.Properties, error) {
	var files []string
	{
		defaultDir, err := p.getDefaultDir(resolver)
		if err != nil {
			return nil, err
		}
		tempFiles, err := p.getFiles(defaultDir, resolver)
		if err != nil {
			return nil, err
		}
		files = append(files, tempFiles...)
	}

	for _, dir := range p.extraDirs {
		tempFiles, err := p.getFiles(dir, resolver)
		if err != nil {
			return nil, err
		}
		files = append(files, tempFiles...)
	}
	files = append(files, p.extraFiles...)

	var ret []conf.Properties
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
		ret = append(ret, c)
	}
	return ret, nil
}
