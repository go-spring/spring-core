/*
 * Copyright 2024 The Go-Spring Authors.
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

// CommandArgsPrefix defines the environment variable name used to override
// the default option prefix. This allows users to customize the prefix used
// for command-line options if needed.
const CommandArgsPrefix = "GS_ARGS_PREFIX"

// CommandArgs represents a structure for handling command-line parameters.
type CommandArgs struct{}

// NewCommandArgs creates and returns a new CommandArgs instance.
func NewCommandArgs() *CommandArgs {
	return &CommandArgs{}
}

// CopyTo processes command-line parameters and sets them as key-value pairs
// in the provided conf.Properties. Parameters should be passed in the form
// of `-D key[=value/true]`.
func (c *CommandArgs) CopyTo(out *conf.MutableProperties) error {
	if len(os.Args) == 0 {
		return nil
	}

	// Default option prefix is "-D", but it can be overridden by the
	// environment variable `GS_ARGS_PREFIX`.
	option := "-D"
	if s := strings.TrimSpace(os.Getenv(CommandArgsPrefix)); s != "" {
		option = s
	}

	cmdArgs := os.Args[1:]
	n := len(cmdArgs)
	for i := range n {
		if cmdArgs[i] == option {
			if i+1 >= n {
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
