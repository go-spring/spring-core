/*
 * Copyright 2012-2024 the original author or authors.
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
	"github.com/go-spring/spring-core/conf/sysconf"
	"github.com/go-spring/spring-core/gs/internal/gs"
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

// PropertySources is a collection of locations.
type PropertySources struct {
	configType ConfigType
	configName string
	locations  []string
}

func NewPropertySources(configType ConfigType, configName string) *PropertySources {
	return &PropertySources{
		configType: configType,
		configName: configName,
	}
}

// Reset resets the locations.
func (p *PropertySources) Reset() {
	p.locations = nil
}

// AddLocation adds a location.
func (p *PropertySources) AddLocation(location ...string) {
	p.locations = append(p.locations, location...)
}

// getDefaultLocations returns the default locations.
func (p *PropertySources) getDefaultLocations(resolver *conf.Properties) (_ []string, err error) {

	var configDir string
	if p.configType == ConfigTypeLocal {
		configDir, err = resolver.Resolve("${spring.application.config.dir:=./conf}")
	} else if p.configType == ConfigTypeRemote {
		configDir, err = resolver.Resolve("${spring.cloud.config.dir:=./conf/remote}")
	} else {
		return nil, fmt.Errorf("unknown config type: %s", p.configType)
	}
	if err != nil {
		return nil, err
	}

	locations := []string{
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
				locations = append(locations, []string{
					fmt.Sprintf("%s/%s-%s.properties", configDir, p.configName, s),
					fmt.Sprintf("%s/%s-%s.yaml", configDir, p.configName, s),
					fmt.Sprintf("%s/%s-%s.toml", configDir, p.configName, s),
					fmt.Sprintf("%s/%s-%s.json", configDir, p.configName, s),
				}...)
			}
		}
	}
	return locations, nil
}

// loadFiles loads all locations and returns a list of properties.
func (p *PropertySources) loadFiles(resolver *conf.Properties) ([]*conf.Properties, error) {
	locations, err := p.getDefaultLocations(resolver)
	if err != nil {
		return nil, err
	}
	locations = append(locations, p.locations...)
	var files []*conf.Properties
	for _, s := range locations {
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
		files = append(files, c)
	}
	return files, nil
}
