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
	"github.com/go-spring/spring-core/util/sysconf"
)

const (
	AllowCircularReferencesProp = "spring.allow-circular-references"
	ForceAutowireIsNullableProp = "spring.force-autowire-is-nullable"
	ActiveProfilesProp          = "spring.profiles.active"
	EnableAppJobsProp           = "spring.enable.app-jobs"
	EnableAppServersProp        = "spring.enable.app-servers"
	EnableSimpleHttpServerProp  = "spring.enable.simple-http-server"
	EnableSimplePProfServerProp = "spring.enable.simple-pprof-server"
)

// AllowCircularReferences enables or disables circular references between beans.
func AllowCircularReferences(enable bool) {
	err := sysconf.Set(AllowCircularReferencesProp, enable)
	_ = err // Ignore error
}

// ForceAutowireIsNullable forces autowire to be nullable.
func ForceAutowireIsNullable(enable bool) {
	err := sysconf.Set(ForceAutowireIsNullableProp, enable)
	_ = err // Ignore error
}

// SetActiveProfiles sets the active profiles for the app.
func SetActiveProfiles(profiles string) {
	err := sysconf.Set(ActiveProfilesProp, profiles)
	_ = err // Ignore error
}

// EnableAppJobs enables or disables the app jobs.
func EnableAppJobs(enable bool) {
	err := sysconf.Set(EnableAppJobsProp, enable)
	_ = err // Ignore error
}

// EnableAppServers enables or disables the app servers.
func EnableAppServers(enable bool) {
	err := sysconf.Set(EnableAppServersProp, enable)
	_ = err // Ignore error
}

// EnableSimpleHttpServer enables or disables the simple HTTP server.
func EnableSimpleHttpServer(enable bool) {
	err := sysconf.Set(EnableSimpleHttpServerProp, enable)
	_ = err // Ignore error
}

// EnableSimplePProfServer enables or disables the simple pprof server.
func EnableSimplePProfServer(enable bool) {
	err := sysconf.Set(EnableSimplePProfServerProp, enable)
	_ = err // Ignore error
}
