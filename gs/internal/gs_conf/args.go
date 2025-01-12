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

package gs_conf

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-spring/spring-core/conf"
)

const CommandArgsPrefix = "GS_ARGS_PREFIX"

// CommandArgs command-line parameters
type CommandArgs struct{}

func NewCommandArgs() *CommandArgs {
	return &CommandArgs{}
}

// CopyTo loads parameters passed in the form of -D key[=value/true].
func (c *CommandArgs) CopyTo(out *conf.Properties) error {
	if len(os.Args) == 0 {
		return nil
	}

	option := "-D"
	if s := strings.TrimSpace(os.Getenv(CommandArgsPrefix)); s != "" {
		option = s
	}

	cmdArgs := os.Args[1:]
	n := len(cmdArgs)
	for i := 0; i < n; i++ {
		if cmdArgs[i] == option {
			if i >= n-1 {
				return fmt.Errorf("cmd option %s needs arg", option)
			}
			next := cmdArgs[i+1]
			ss := strings.SplitN(next, "=", 2)
			if len(ss) == 1 {
				ss = append(ss, "true")
			}
			if err := out.Set(ss[0], ss[1]); err != nil {
				return err
			}
		}
	}
	return nil
}
