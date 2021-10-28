// Copyright 2020 The Tcl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"modernc.org/libc"
	"modernc.org/tcl"
	"modernc.org/tcl/internal/tclsh"
)

const tclLibrary = "TCL_LIBRARY"

func main() {
	if os.Getenv(tclLibrary) == "" {
		dir, err := ioutil.TempDir("", "gotclsh-")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if err := tcl.Library(dir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		libc.AtExit(func() { os.RemoveAll(dir) })
		os.Setenv(tclLibrary, dir)
	}
	libc.Start(tclsh.Main)
}
