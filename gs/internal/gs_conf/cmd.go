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
	"os"
	"strings"

	"github.com/go-spring/spring-base/util"
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

// CopyTo extracts command-line parameters and stores them as key-value pairs.
// Supported formats include:
//
//   - <prefix> key=value
//   - <prefix> key        (defaults to "true")
//   - <prefix>key=value   (inline form)
//
// The default prefix is "-D", which can be overridden by the environment
// variable `GS_ARGS_PREFIX`.
func (c *CommandArgs) CopyTo(p *conf.MutableProperties) error {
	if len(os.Args) <= 1 {
		return nil
	}

	fileID := p.AddFile("Args")

	// Determine the option prefix.
	option := "-D"
	if s := strings.TrimSpace(os.Getenv(CommandArgsPrefix)); s != "" {
		option = s
	}

	cmdArgs := os.Args[1:]
	for i := 0; i < len(cmdArgs); i++ {
		var str string
		if cmdArgs[i] == option {
			// separated form: <prefix> key=value
			if i+1 >= len(cmdArgs) {
				return util.FormatError(nil, "cmd option %s: needs arg", option)
			}
			i++
			str = cmdArgs[i]
		} else if s, ok := strings.CutPrefix(cmdArgs[i], option); ok {
			// inline form: <prefix>key=value
			str = s
		} else {
			// not a Go-Spring command-line option
			continue
		}
		if str = strings.TrimSpace(str); str == "" {
			return util.FormatError(nil, "cmd option %s: needs arg", option)
		}
		ss := strings.SplitN(str, "=", 2)
		if len(ss) == 1 {
			ss = append(ss, "true")
		}
		if err := p.Set(ss[0], ss[1], fileID); err != nil {
			return util.FormatError(err, "set cmd option %s error", str)
		}
	}
	return nil
}
