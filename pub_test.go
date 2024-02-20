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
	"testing"
)

var printSuccess = true

func assert(test *testing.T, actual, expected any, err string) {
	if actual != expected {
		test.Errorf("FAIL. "+err+" expected: %v, got: %v", expected, actual)
	} else if printSuccess {
		test.Logf("PASS. expected: %v, got: %v", expected, actual)
	}
}

func testMod(v string) []byte {
	return []byte(Stringf("// pub is a simple module for publishing go modules.\nmodule github.com/jcdotter/pub //%s\n\ngo 1.17", v))
}

func printMod(n, v string, i, l int) {
	fmt.Printf("module parsed\n\tmodule: %s\n\tversion: %s\n\tending at: %d\n", n, v, i)
}

func TestParser(t *testing.T) {
	var mod []byte
	var modErr = "module not parsed correctly."
	var verErr = "version not parsed correctly."
	var lenErr = "length not returned correctly."
	var idxErr = "index not returned correctly."

	// test without version
	mod = testMod(" go version 1.17")
	m, v, l, i := parseMod(mod, 0)
	printMod(m, v, i, l)
	assert(t, m, "github.com/jcdotter/pub", modErr)
	assert(t, v, "", verErr)
	assert(t, l, 0, lenErr)
	assert(t, i, 83, idxErr)

	// test with version
	mod = testMod("v0.1.0")
	m, v, l, i = parseMod(mod, 0)
	printMod(m, v, i, l)
	assert(t, m, "github.com/jcdotter/pub", modErr)
	assert(t, v, "0.1.0", verErr)
	assert(t, l, 9, lenErr)
	assert(t, i, 83, idxErr)

	// test version update
	mod = updateModVersion(mod, i, l, "0.1.1")
	m, v, l, i = parseMod(mod, 0)
	printMod(m, v, i, l)
	assert(t, m, "github.com/jcdotter/pub", modErr)
	assert(t, v, "0.1.1", verErr)
	assert(t, l, 9, lenErr)
	assert(t, i, 83, idxErr)
}

func TestVersion(t *testing.T) {
	var v string
	var valid bool
	var err = "version not returned correctly."
	var valErr = "version validity not returned correctly."

	// test without version
	v, valid = assessVersion("", "")
	assert(t, valid, true, valErr)
	assert(t, v, "0.0.0", err)

	// test patch version
	v, valid = assessVersion("0.1.0", "")
	assert(t, valid, true, valErr)
	assert(t, v, "0.1.1", err)

	// test user version
	v, valid = assessVersion("0.1.0", "0.1.1")
	assert(t, valid, true, valErr)
	assert(t, v, "0.1.1", err)

	// test invalid version
	v, valid = assessVersion("0.1.0", "v.1.1")
	assert(t, valid, false, valErr)
	assert(t, v, "", err)

	// test lesser user version
	v, valid = assessVersion("0.1.0", "0.0.9")
	assert(t, valid, false, valErr)
	assert(t, v, "", err)
}
