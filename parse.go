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
	"os"
)

const (
	lin  = '\n'
	ret  = '\r'
	tab  = '\t'
	spc  = ' '
	esc  = '\\'
	aps  = '\''
	qot  = '"'
	sep  = '='
	com  = '/'
	file = "go.mod"
	mod  = "module "
	pre  = "//v"
)

// parseModFile parses the go.mod file in the current directory
// and returns the module name and version, if it exists.
func parseModFile() (name, version string, bytes []byte, vpos, vlen int, err error) {
	if bytes, err = os.ReadFile(file); err != nil {
		return
	}
	name, version, vlen, vpos = parseMod(bytes, 0)
	return
}

// updateModFile updates the go.mod file in the current directory
// with the new version provided by the user.
func updateModFile(bytes []byte, vpos, vlen int, version string) (err error) {
	bytes = updateModVersion(bytes, vpos, vlen, version)
	return os.WriteFile(file, bytes, 0644)
}

// search returns the index of the byte provided in the bytes
// provided from the go.mod file.
func search(b byte, in []byte, at int) (i int) {
	i = at
	for i < len(in) && in[i] != b {
		i++
	}
	return
}

// isspace checks if a byte is a space character.
func isspace(b byte) bool {
	return b == ret || b == lin || b == tab || b == spc
}

// isnum checks if a byte is a number.
func isnum(b byte) bool {
	return b >= '0' && b <= '9'
}

// iscomm checks if the bytes provided are a comment.
func iscomm(b []byte, at int) bool {
	return b[at] == com && b[at+1] == com
}

// next returns the next non-space character in the bytes provided
// from the go.mod file.
func next(b []byte, at int, skipcom bool) int {
	for at < len(b) {
		switch c := b[at]; {
		case skipcom && c == com:
			at = skipComment(b, at)
			fallthrough
		case isspace(c):
			at++
		default:
			return at
		}
	}
	return at
}

// skipComment skips a comment in the bytes provided
// from the go.mod file.
func skipComment(b []byte, at int) int {
	if iscomm(b, at) {
		return search(lin, b, at+2)
	}
	return at
}

// parseMod parses the module name and version from the bytes
// provided from the go.mod file.
func parseMod(b []byte, at int) (name, version string, vlen, i int) {
	for i = at; i < len(b); i++ {
		i = next(b, i, true)
		if len(b) < i+7 {
			i = len(b)
			return
		}
		if string(b[i:i+7]) == mod {
			i += 7
			name, i = parseModName(b, i)
			version, vlen, i = parseModVersion(b, i)
			return
		}
	}
	return
}

// parseModName parses the name of a module
// from the bytes provided from the go.mod file.
func parseModName(b []byte, at int) (name string, i int) {
	at = next(b, at, false)
	i = at
	if c := b[at]; c == aps || c == qot {
		i = search(c, b, at+1)
		if b[i-1] != esc {
			name = string(b[at+1 : i])
			i++
			return
		}
	}
	for ; i < len(b); i++ {
		if isspace(b[i]) || iscomm(b, i) {
			break
		}
	}
	name = string(b[at:i])
	return
}

// parseModVersion parses the version of a module
// from the bytes provided from the go.mod file,
// if it exists.
func parseModVersion(b []byte, at int) (version string, vlen, i int) {
	pos := next(b, at, false)
	if b[pos] == com && string(b[pos:pos+len(pre)]) == pre {
		pos += len(pre)
		i = search(lin, b, pos)
		if v := b[pos:i]; validVersion(v) {
			version = string(v)
			vlen = i - at
		}
	}
	i = at
	return
}

// updateModVersion updates the go.mod file in the current directory
// with the new version provided by the user.
func updateModVersion(bytes []byte, vpos, vlen int, version string) []byte {
	v := make([]byte, 0, len(bytes))
	return append(bytes[:vpos], append(append(append(v, spc), []byte(pre+version)...), bytes[vpos+vlen:]...)...)
}
