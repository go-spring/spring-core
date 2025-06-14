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
	"strconv"
)

const (
	AllowCircularReferencesProp = "spring.allow-circular-references"
	ForceAutowireIsNullableProp = "spring.force-autowire-is-nullable"
	ActiveProfilesProp          = "spring.profiles.active"
	EnableJobsProp              = "spring.app.enable-jobs"
	EnableServersProp           = "spring.app.enable-servers"
	EnableSimpleHttpServerProp  = "spring.enable.simple-http-server"
	EnableSimplePProfServerProp = "spring.enable.simple-pprof-server"
)

// AllowCircularReferences enables or disables circular references between beans.
func AllowCircularReferences(enable bool) {
	Property(AllowCircularReferencesProp, strconv.FormatBool(enable))
}

// ForceAutowireIsNullable forces autowire to be nullable.
func ForceAutowireIsNullable(enable bool) {
	Property(ForceAutowireIsNullableProp, strconv.FormatBool(enable))
}

// SetActiveProfiles sets the active profiles for the app.
func SetActiveProfiles(profiles string) {
	Property(ActiveProfilesProp, profiles)
}

// EnableJobs enables or disables the app jobs.
func EnableJobs(enable bool) {
	Property(EnableJobsProp, strconv.FormatBool(enable))
}

// EnableServers enables or disables the app servers.
func EnableServers(enable bool) {
	Property(EnableServersProp, strconv.FormatBool(enable))
}

// EnableSimpleHttpServer enables or disables the simple HTTP server.
func EnableSimpleHttpServer(enable bool) {
	Property(EnableSimpleHttpServerProp, strconv.FormatBool(enable))
}

// EnableSimplePProfServer enables or disables the simple pprof server.
func EnableSimplePProfServer(enable bool) {
	Property(EnableSimplePProfServerProp, strconv.FormatBool(enable))
}
