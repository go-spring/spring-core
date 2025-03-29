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

package gs_app_test

import (
	"sync"
	"testing"

	"github.com/go-spring/spring-core/gs/internal/gs_app"
	"github.com/go-spring/spring-core/util/assert"
)

func TestReadySignal(t *testing.T) {

	t.Run("intercept", func(t *testing.T) {
		const workers = 3

		signal := gs_app.NewReadySignal()
		for i := 0; i < workers; i++ {
			num := i
			signal.Add()
			go func() {
				if num == 0 {
					signal.Intercept()
				} else {
					<-signal.TriggerAndWait()
				}
			}()
		}

		signal.Wait()
		assert.True(t, signal.Intercepted())
	})

	t.Run("success", func(t *testing.T) {
		const workers = 3

		var wg sync.WaitGroup
		wg.Add(workers)
		defer wg.Wait()

		signal := gs_app.NewReadySignal()
		for i := 0; i < workers; i++ {
			signal.Add()
			go func() {
				defer wg.Done()
				<-signal.TriggerAndWait()
			}()
		}

		signal.Wait()
		assert.False(t, signal.Intercepted())

		signal.Close()
	})
}
