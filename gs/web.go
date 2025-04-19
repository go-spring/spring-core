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

package gs

import (
	"net/http"
)

func init() {
	Object(http.DefaultServeMux).Condition(
		OnProperty("spring.enable.default-serve-mux").HavingValue("true"),
	)
}

// Handle registers the handler for the given pattern
func Handle(pattern string, handler http.Handler) {
	EnableDefaultServeMux(true)
	http.Handle(pattern, handler)
}

// HandleFunc registers the handler function for the given pattern
func HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	EnableDefaultServeMux(true)
	http.HandleFunc(pattern, handler)
}
