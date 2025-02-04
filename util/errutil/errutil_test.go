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

package errutil_test

import (
	"os"
	"testing"

	"github.com/go-spring/spring-core/util/assert"
	"github.com/go-spring/spring-core/util/errutil"
)

func TestWrapError(t *testing.T) {
	var err error

	err = os.ErrNotExist
	err = errutil.WrapError(err, "open file error: file=%s", "test.php")
	err = errutil.WrapError(err, "read file error")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), `read file error << open file error: file=test.php << file does not exist`)

	errutil.LineBreak = " / "
	err = os.ErrNotExist
	err = errutil.WrapError(err, "open file error: file=%s", "test.php")
	err = errutil.WrapError(err, "read file error")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), `read file error / open file error: file=test.php / file does not exist`)
}
