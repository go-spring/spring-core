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

package gstest_test

import (
	"os"
	"testing"

	"github.com/go-spring/spring-core/gs/gstest"
	"github.com/go-spring/spring-core/util/assert"
)

func TestMain(m *testing.M) {
	err := gstest.Init()
	if err != nil {
		panic(err)
	}
	os.Exit(gstest.Run(m))
}

func TestConfig(t *testing.T) {
	os.Clearenv()
	os.Setenv("GS_SPRING_PROFILES_ACTIVE", "dev")
	assert.Equal(t, gstest.Prop("spring.profiles.active"), "dev")
}
