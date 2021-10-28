// Copyright 2017 The HTTPFS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httpfs // import "modernc.org/httpfs"

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"
	"time"
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

func use(...interface{}) {}

func init() {
	use(caller, dbg, TODO) //TODOOK
}

// ============================================================================

func Test(t *testing.T) {
	fs := NewFileSystem(map[string]string{
		"/a/a":   "a",
		"/a/b/c": "b/c",
		"/a/b/d": "b/d",
		"/a/d":   "d",
	}, time.Now())
	for k := range fs.files {
		f, err := fs.Open(k)
		if err != nil {
			t.Fatal(err)
		}
		s, err := ioutil.ReadAll(f)
		if g, e := string(s), k[3:]; g != e {
			t.Fatalf("%q %q", g, e)
		}
	}
}
