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

package bootstrap

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-spring/spring-core/gs"
)

func init() {
	gs.Boot().Object(gs.FuncRunner(initRemoteConfig)).AsRunner().OnProfiles("online")
}

// initRemoteConfig initializes the remote configuration setup
func initRemoteConfig() error {
	if err := getRemoteConfig(); err != nil {
		return err
	}
	gs.Object(gs.FuncJob(refreshRemoteConfig)).AsJob()
	return nil
}

// refreshRemoteConfig periodically refreshes the remote configuration
func refreshRemoteConfig(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("config updater exit")
			return nil
		case <-time.After(time.Millisecond * 500):
			if err := getRemoteConfig(); err != nil {
				fmt.Println("get remote config error:", err)
				return err
			}
			if err := gs.RefreshProperties(); err != nil {
				fmt.Println("refresh properties error:", err)
				return err
			}
			fmt.Println("refresh properties success")
		}
	}
}

// getRemoteConfig fetches and writes the remote configuration to a local file
func getRemoteConfig() error {
	err := os.MkdirAll("./conf/remote", os.ModePerm)
	if err != nil {
		return err
	}

	const data = `
server.addr=0.0.0.0:9090

log.access.name=access.log
log.access.dir=./log

log.biz.name=biz.log
log.biz.dir=./log

log.dao.name=dao.log
log.dao.dir=./log

refresh_time=%v`

	const file = "conf/remote/app-online.properties"
	str := fmt.Sprintf(data, time.Now().UnixMilli())
	return os.WriteFile(file, []byte(str), os.ModePerm)
}
