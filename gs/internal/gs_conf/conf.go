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
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/util/sysconf"
)

/******************************** AppConfig **********************************/

// AppConfig is a layered app configuration.
type AppConfig struct {
	LocalFile   *PropertySources
	RemoteFile  *PropertySources
	RemoteProp  gs.Properties
	Environment *Environment
	CommandArgs *CommandArgs
}

func NewAppConfig() *AppConfig {
	return &AppConfig{
		LocalFile:   NewPropertySources(ConfigTypeLocal, "application"),
		RemoteFile:  NewPropertySources(ConfigTypeRemote, "application"),
		Environment: NewEnvironment(),
		CommandArgs: NewCommandArgs(),
	}
}

func merge(out *conf.Properties, sources ...interface {
	CopyTo(out *conf.Properties) error
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

// Refresh merges all layers into a properties as read-only.
func (c *AppConfig) Refresh() (gs.Properties, error) {

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
		CopyTo(out *conf.Properties) error
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

// BootConfig is a layered boot configuration.
type BootConfig struct {
	LocalFile   *PropertySources
	Environment *Environment
	CommandArgs *CommandArgs
}

func NewBootConfig() *BootConfig {
	return &BootConfig{
		LocalFile:   NewPropertySources(ConfigTypeLocal, "bootstrap"),
		Environment: NewEnvironment(),
		CommandArgs: NewCommandArgs(),
	}
}

// Refresh merges all layers into a properties as read-only.
func (c *BootConfig) Refresh() (gs.Properties, error) {

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
		CopyTo(out *conf.Properties) error
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

type ConfigType string

const (
	ConfigTypeLocal  ConfigType = "local"
	ConfigTypeRemote ConfigType = "remote"
)

// PropertySources is a collection of files.
type PropertySources struct {
	configType ConfigType
	configName string
	extraFiles []string
}

func NewPropertySources(configType ConfigType, configName string) *PropertySources {
	return &PropertySources{
		configType: configType,
		configName: configName,
	}
}

// Reset resets the files.
func (p *PropertySources) Reset() {
	p.extraFiles = nil
}

// AddFile adds a file.
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

// getDefaultFiles returns the default files.
func (p *PropertySources) getDefaultFiles(resolver *conf.Properties) (_ []string, err error) {

	var configDir string
	if p.configType == ConfigTypeLocal {
		configDir, err = resolver.Resolve("${spring.app.config.dir:=./conf}")
	} else if p.configType == ConfigTypeRemote {
		configDir, err = resolver.Resolve("${spring.cloud.config.dir:=./conf/remote}")
	} else {
		return nil, fmt.Errorf("unknown config type: %s", p.configType)
	}
	if err != nil {
		return nil, err
	}

	files := []string{
		fmt.Sprintf("%s/%s.properties", configDir, p.configName),
		fmt.Sprintf("%s/%s.yaml", configDir, p.configName),
		fmt.Sprintf("%s/%s.toml", configDir, p.configName),
		fmt.Sprintf("%s/%s.json", configDir, p.configName),
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
					fmt.Sprintf("%s/%s-%s.properties", configDir, p.configName, s),
					fmt.Sprintf("%s/%s-%s.yaml", configDir, p.configName, s),
					fmt.Sprintf("%s/%s-%s.toml", configDir, p.configName, s),
					fmt.Sprintf("%s/%s-%s.json", configDir, p.configName, s),
				}...)
			}
		}
	}
	return files, nil
}

// loadFiles loads all files and returns a list of properties.
func (p *PropertySources) loadFiles(resolver *conf.Properties) (ret []*conf.Properties, err error) {
	files, err := p.getDefaultFiles(resolver)
	if err != nil {
		return nil, err
	}
	files = append(files, p.extraFiles...)
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
	return
}
