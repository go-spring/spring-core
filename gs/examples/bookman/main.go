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

package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/go-spring/spring-core/gs"
	"github.com/go-spring/spring-core/util/iterutil"
	"github.com/go-spring/spring-core/util/syslog"

	_ "github.com/go-spring/spring-core/gs/examples/bookman/src/app"
	_ "github.com/go-spring/spring-core/gs/examples/bookman/src/biz"
)

const banner = `
  ____                 _     __  __               
 | __ )   ___    ___  | | __|  \/  |  __ _  _ __  
 |  _ \  / _ \  / _ \ | |/ /| |\/| | / _' || '_ \ 
 | |_) || (_) || (_) ||   < | |  | || (_| || | | |
 |____/  \___/  \___/ |_|\_\|_|  |_| \__,_||_| |_|
`

func init() {
	gs.Banner(banner)
	gs.SetActiveProfiles("online")
	gs.EnableSimplePProfServer(true)
}

func init() {
	gs.FuncJob(runTest).Name("#job")
}

func init() {
	var execDir string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		execDir = filepath.Dir(filename)
	}
	err := os.Chdir(execDir)
	if err != nil {
		panic(err)
	}
	workDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fmt.Println(workDir)
}

func main() {
	// Start the application and log errors if startup fails
	if err := gs.Run(); err != nil {
		syslog.Errorf("app run failed: %s", err.Error())
	}
}

// runTest performs a simple test.
func runTest(ctx context.Context) error {
	time.Sleep(time.Millisecond * 500)

	iterutil.Times(5, func(_ int) {
		url := "http://127.0.0.1:9090/books"
		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		fmt.Print(string(b))
		time.Sleep(time.Millisecond * 400)
	})

	// Shut down the application gracefully
	gs.ShutDown()
	return nil
}
