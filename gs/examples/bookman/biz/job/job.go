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

package job

import (
	"context"
	"fmt"
	"time"

	"github.com/go-spring/spring-core/gs"
)

func init() {
	gs.Object(&Job{}).AsJob()
}

type Job struct {
}

func (x *Job) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("job exit")
			return nil
		case <-time.After(time.Second * 5):
			fmt.Println("job sleep end")
		}
	}
}
