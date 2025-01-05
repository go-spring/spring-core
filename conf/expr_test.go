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

package conf_test

import (
	"testing"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/util/assert"
)

func TestExpr(t *testing.T) {
	conf.RegisterValidateFunc("checkInt", func(i int) bool {
		return i < 5
	})
	var i struct {
		A int `value:"${a}" expr:"checkInt($)"`
	}
	p, err := conf.Map(map[string]interface{}{
		"a": 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = p.Bind(&i); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 4, i.A)
}
