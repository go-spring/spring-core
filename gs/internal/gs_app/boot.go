/*
 * Copyright 2012-2019 the original author or authors.
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
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_conf"
	"github.com/go-spring/spring-core/gs/internal/gs_core"
)

type Boot struct {
	C gs.Container
	P *gs_conf.BootConfig

	Runners []AppRunner `autowire:"${spring.boot.runners:=*?}"`
}

func NewBoot() *Boot {
	b := &Boot{
		C: gs_core.New(),
		P: gs_conf.NewBootConfig(),
	}
	b.C.Object(b)
	return b
}

func (b *Boot) Run() error {

	p, err := b.P.Refresh()
	if err != nil {
		return err
	}

	err = b.C.RefreshProperties(p)
	if err != nil {
		return err
	}

	err = b.C.Refresh()
	if err != nil {
		return err
	}

	// 执行命令行启动器
	for _, r := range b.Runners {
		r.Run(b.C.(gs.Context))
	}

	b.C.Close()
	return nil
}
