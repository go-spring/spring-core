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
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-spring/spring-core/gs"
	"github.com/go-spring/spring-core/util/iterutil"
	"github.com/go-spring/spring-core/util/syslog"

	_ "github.com/go-spring/spring-core/gs/examples/bookman/app"
	_ "github.com/go-spring/spring-core/gs/examples/bookman/biz"
	_ "github.com/go-spring/spring-core/gs/examples/bookman/idl"
)

func init() {
	gs.SetActiveProfiles("online")
	gs.EnableSimplePProfServer(true)
}

func main() {
	// Unset certain environment variables before running the application
	_ = os.Unsetenv("_")
	_ = os.Unsetenv("TERM")
	_ = os.Unsetenv("TERM_SESSION_ID")

	// Run test after a short delay in a separate goroutine
	go func() {
		time.Sleep(time.Millisecond * 500)
		runTest()
	}()

	// Start the application and log errors if startup fails
	if err := gs.Run(); err != nil {
		syslog.Errorf("app run failed: %s", err.Error())
	}
}

// runTest performs a simple test.
func runTest() {

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
}
