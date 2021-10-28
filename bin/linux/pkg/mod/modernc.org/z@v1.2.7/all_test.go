// Copyright 2021 The Zlib-Go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package z // import "modernc.org/z"

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"

	"modernc.org/ccgo/v3/lib"
)

func caller(s string, va ...interface{}) {
	if s == "" {
		s = strings.Repeat("%v ", len(va))
	}
	_, fn, fl, _ := runtime.Caller(2)
	fmt.Fprintf(os.Stderr, "# caller: %s:%d: ", path.Base(fn), fl)
	fmt.Fprintf(os.Stderr, s, va...)
	fmt.Fprintln(os.Stderr)
	_, fn, fl, _ = runtime.Caller(1)
	fmt.Fprintf(os.Stderr, "# \tcallee: %s:%d: ", path.Base(fn), fl)
	fmt.Fprintln(os.Stderr)
	os.Stderr.Sync()
}

func dbg(s string, va ...interface{}) {
	if s == "" {
		s = strings.Repeat("%v ", len(va))
	}
	_, fn, fl, _ := runtime.Caller(1)
	fmt.Fprintf(os.Stderr, "# dbg %s:%d: ", path.Base(fn), fl)
	fmt.Fprintf(os.Stderr, s, va...)
	fmt.Fprintln(os.Stderr)
	os.Stderr.Sync()
}

func TODO(...interface{}) string { //TODOOK
	_, fn, fl, _ := runtime.Caller(1)
	return fmt.Sprintf("# TODO: %s:%d:\n", path.Base(fn), fl) //TODOOK
}

func stack() string { return string(debug.Stack()) }

func use(...interface{}) {}

func init() {
	use(caller, dbg, TODO, stack) //TODOOK
}

// ----------------------------------------------------------------------------

var (
	xtags = flag.String("xtags", "non.existent.tag", "")
)

func TestMain(m *testing.M) {
	flag.Parse()
	rc := m.Run()
	os.Exit(rc)
}

func Test(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "go-test-zlib-")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(tmpDir)

	wd, err := ccgo.AbsCwd()
	if err != nil {
		t.Fatal(err)
	}

	goos := runtime.GOOS
	goarch := runtime.GOARCH
	mg := filepath.Join(wd, "internal", fmt.Sprintf("minigzip_%s_%s.go", goos, goarch))
	ex := filepath.Join(wd, "internal", fmt.Sprintf("example_%s_%s.go", goos, goarch))
	mgBin := "minigzip"
	exBin := "example"
	if goos == "windows" {
		mgBin += ".exe"
		exBin += ".exe"
	}
	if ccgo.Shell("go", "build", "-tags", *xtags, "-o", filepath.Join(tmpDir, mgBin), mg); err != nil {
		t.Fatal(err)
	}

	if ccgo.Shell("go", "build", "-tags", *xtags, "-o", filepath.Join(tmpDir, exBin), ex); err != nil {
		t.Fatal(err)
	}

	if err := ccgo.InDir(tmpDir, func() error {
		switch goos {
		case "windows":
			if err := ccgo.InDir(tmpDir, func() error {
				if out, err := ccgo.Shell("cmd.exe", "/c", fmt.Sprintf("echo hello world | %s | %[1]s -d", mgBin, exBin)); err != nil {
					return fmt.Errorf("%s\nFAIL: %v", out, err)
				}

				return nil
			}); err != nil {
				t.Fatal(err)
			}
		default:
			if err := ccgo.InDir(tmpDir, func() error {
				mgBin = "./" + mgBin
				exBin = "./" + exBin
				if out, err := ccgo.Shell("sh", "-c", fmt.Sprintf("echo hello world | %s | %[1]s -d && %s tmp", mgBin, exBin)); err != nil {
					return fmt.Errorf("%s\nFAIL: %v", out, err)
				}

				return nil
			}); err != nil {
				t.Fatal(err)
			}
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
