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

package gs_app

import (
	"bytes"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/go-spring/gs-assert/assert"
	"github.com/go-spring/spring-core/gs/internal/gs_bean"
	"github.com/go-spring/spring-core/gs/internal/gs_conf"
)

func TestBoot(t *testing.T) {

	t.Run("flag is false", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		fileID := gs_conf.SysConf.AddFile("boot_test.go")
		_ = gs_conf.SysConf.Set("a", "123", fileID)
		_ = os.Setenv("GS_A_B", "456")
		boot := NewBoot().(*BootImpl)
		err := boot.Run()
		assert.ThatError(t, err).Nil()
	})

	t.Run("config refresh error", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		fileID := gs_conf.SysConf.AddFile("boot_test.go")
		_ = gs_conf.SysConf.Set("a", "123", fileID)
		_ = os.Setenv("GS_A_B", "456")
		boot := NewBoot().(*BootImpl)
		boot.Object(bytes.NewBuffer(nil))
		err := boot.Run()
		assert.ThatError(t, err).Matches("property conflict at path a.b")
	})

	t.Run("container refresh error", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		boot := NewBoot().(*BootImpl)
		boot.RootBean(boot.Provide(func() (*bytes.Buffer, error) {
			return nil, errors.New("fail to create bean")
		}))
		err := boot.Run()
		assert.ThatError(t, err).Matches("fail to create bean")
	})

	t.Run("runner return error", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		boot := NewBoot().(*BootImpl)
		boot.FuncRunner(func() error {
			return errors.New("runner return error")
		})
		err := boot.Run()
		assert.ThatError(t, err).Matches("runner return error")
	})

	t.Run("success", func(t *testing.T) {
		Reset()
		t.Cleanup(Reset)

		boot := NewBoot().(*BootImpl)
		bd := gs_bean.NewBean(reflect.ValueOf(funcRunner(func() error {
			return nil
		}))).AsRunner().Caller(1)
		boot.Register(bd)
		boot.Config().LocalFile.Reset()
		err := boot.Run()
		assert.ThatError(t, err).Nil()
	})
}
