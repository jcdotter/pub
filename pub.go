// Copyright 2023 james dotter.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://github.com/jcdotter/go/LICENSE
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// ----------------------------------------------------------------------------
// CONSOLE MESSAGES

const (
	versionCmd     = "pub v0.1.5\n"
	versionDefault = "0.0.0"
	versionErr     = "Invalid version: %s\n"
	versionNew     = "New version: %s\n"
	pubInit        = "Preparing to publish.\nParsing go.mod file: "
	pubUpdate      = "Updating go.mod file: "
	pubComplete    = "Your module has been published.\n"
	modParseErr    = "ERROR. Ensure you are in the correct directory and have a go.mod file.\n"
	modUpdateErr   = "ERROR. Could not update go.mod file.\n"
	modDone        = "done. (module: %s, version: %s)\n"
	invalidCmd     = "Invalid command. Use 'pub help' for usage.\n"
	helpCmd        = "" +
		"PUB is a simple commandline app for publishing go modules.\n" +
		"On pulblish, all changes are committed and pushed to github as the\n" +
		"version provided. The module is then tagged and the new version is\n" +
		"listed on the GOPROXY.\n\n" +
		"Usage: pub [command]\n" +
		"Commands:\n" +
		"  h, help      Display this help message\n" +
		"  v, version   Display the current version\n" +
		"  p, patch     Publish changes as the next patch version\n" +
		"  v#.#.#       Publish changes as the specified version\n"
	gitMsg     = "Committing changes: "
	gitAdd     = "git add ."
	gitCom     = "git commit -m 'v%s'"
	gitTag     = "git tag v%s"
	gitPush    = "git push origin main"
	gitPushOrg = "git push origin v%s"
	gitAddErr  = "ERROR. Could not add files to git.\n"
	gitComErr  = "ERROR. Could not commit changes.\n"
	gitTagErr  = "ERROR. Could not tag version.\n"
	gitPushErr = "ERROR. Could not push changes to github.\n"
	goMsg      = "Listing version on GOPROXY: "
	goProxy    = "go list -m %s@v%s"
	goProxyErr = "ERROR. Could not list version on GOPROXY.\n"
	done       = "done.\n"
)

// ----------------------------------------------------------------------------
// APPLICATION FUNCTIONS

// Run checks the command line arguments and runs the appropriate
// function. If no arguments are provided, it prints the help message.
func Run() {
	switch a, uv := cmdAction(); a {
	case 0:
		Console(invalidCmd)
		return
	case 1:
		Console(helpCmd)
		return
	case 2:
		Console(versionCmd)
		return
	case 3:
		Publish(uv)
		return
	}
}

// cmdAction checks if the user has provided a command and
// returns a byte representing the action to take. If no
// flag is provided, it returns the help action.
// Available actions:
// 0: invalid command;
// 1: help;
// 2: version;
// 3: publish.
func cmdAction() (action byte, v string) {
	if len(os.Args) == 1 {
		return 1, ""
	}
	if len(os.Args) > 2 {
		return 0, ""
	}
	f := os.Args[1]
	switch f {
	case "h", "help":
		action = 1
	case "v", "version":
		action = 2
	case "p", "patch":
		action = 3
	default:
		if f[0] == 'v' && isnum(f[1]) {
			action = 3
			v = f[1:]
		}
	}
	return
}

// ----------------------------------------------------------------------------
// PUBLISH FUNCTIONS

func Publish(v string) {
	Console(pubInit)
	n, v, ok := updateVersion(v)
	if !ok {
		return
	}
	gitCommit(v)
	goProxyList(n, v)
	Console(pubComplete)
}

func gitCommit(version string) {
	Console(gitMsg)
	Command(gitAddErr, gitAdd)
	Command(gitComErr, gitCom, version)
	Command(gitTagErr, gitTag, version)
	Command(gitPushErr, gitPushOrg, version)
	Command(gitPushErr, gitPush)
	Console(done)
}

func goProxyList(module, version string) {
	Console(goMsg)
	os.Setenv("GOPROXY", "proxy.golang.org")
	Command(goProxyErr, goProxy, module, version)
	Console(done)
}

// ----------------------------------------------------------------------------
// VERSION FUNCS

// updateVersion updates the version in the go.mod file and
// returns the module name, new version, and a boolean value
// indicating if the version is valid. If the user version is
// empty, it returns the next version. If the user version is
// invalid, it returns an empty string and false. If the user
// version is valid and greater than the current version, it
// returns the user version and true.
func updateVersion(userVersion string) (name, newVersion string, valid bool) {
	name, cv, b, vpos, vlen, err := parseModFile()
	if err != nil {
		Console(modParseErr)
		return
	}
	Console(modDone, name, cv)
	newVersion, valid = assessVersion(cv, userVersion)
	if !valid {
		Console(versionErr, userVersion)
		return
	}
	Console(versionNew, newVersion)
	Console(pubUpdate)
	if err = updateModFile(b, vpos, vlen, newVersion); err != nil {
		Console(modUpdateErr)
		return
	}
	Console(done)
	return
}

// assessVersion checks if the user provided version is valid and
// greater than the current version. If the user version is empty,
// it returns the next version. If the user version is invalid, it
// returns an empty string and false. If the user version is valid
// and greater than the current version, it returns the user version
// and true.
func assessVersion(cv, uv string) (nv string, valid bool) {
	if uv == "" {
		return nextVersion(cv), true
	}
	if !validUserVersion(cv, uv) {
		return
	}
	return uv, true
}

// nextVersion increments the patch version of the current version
// and returns the new version.
func nextVersion(cv string) string {
	if cv == "" {
		return versionDefault
	}
	vs := strings.Split(cv, ".")
	vi := len(vs) - 1
	pv, _ := strconv.Atoi(vs[vi])
	vs[vi] = strconv.Itoa(pv + 1)
	return strings.Join(vs, ".")
}

// validVersion checks if b has a valid version format.
// eg. 0.0.0
func validVersion(b []byte) bool {
	if len(b) > 4 && isnum(b[0]) && isnum(b[len(b)-1]) {
		for i := 0; i < len(b); i++ {
			if c := b[i]; !isnum(c) && c != '.' {
				return false
			}
		}
		return true
	}
	return false
}

// validVersion checks if a version string provided by the user is valid
func validUserVersion(current, user string) bool {
	if !validVersion([]byte(user)) {
		return false
	}
	if current == "" || current == user {
		return true
	}
	cparts := strings.Split(current, ".")
	uparts := strings.Split(user, ".")
	if len(cparts) != len(uparts) {
		return false
	}
	for i, cp := range cparts {
		if c, err := strconv.Atoi(cp); err != nil {
			return false
		} else if u, _ := strconv.Atoi(uparts[i]); err != nil {
			return false
		} else if u > c {
			return true
		} else if u < c {
			return false
		}
	}
	return false
}

// ----------------------------------------------------------------------------
// CONSOLE FUNCS

func Command(err, c string, a ...string) {
	if len(a) > 0 {
		c = Stringf(c, a...)
	}
	ex := strings.Split(c, " ")
	if e := exec.Command(ex[0], ex[1:]...).Run(); e != nil {
		Console(err)
		Console("%s\n", e.Error())
	}
}

// Console prints a message to the console.
func Console(format string, a ...string) {
	fmt.Print(Stringf(format, a...))
}

// Stringf formats a string with the provided arguments.
func Stringf(format string, a ...string) string {
	if a == nil {
		return format
	}
	s := strings.Builder{}
	s.Grow(2 * len(format))
	defer s.Reset()
	pos, elem := 0, 0
	for i := 0; i < len(format); i++ {
		if format[i] == '%' && format[i+1] == 's' {
			s.WriteString(format[pos:i])
			if elem < len(a) {
				s.WriteString(a[elem])
				pos = i + 2
				elem++
				continue
			}
			panic("not enough arguments for format string")
		}
	}
	s.WriteString(format[pos:])
	return s.String()
}
