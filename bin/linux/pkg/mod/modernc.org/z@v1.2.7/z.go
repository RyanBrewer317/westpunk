// Copyright 2021 The Zlib-Go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate go run generator.go

// Package z implements the native Go API for zlib.
//
// The API si currently empty.
//
// zlib 1.2.11 is a general purpose data compression library.  All the code is
// thread safe.  The data format used by the zlib library is described by RFCs
// (Request for Comments) 1950 to 1952 in the files
// http://tools.ietf.org/html/rfc1950 (zlib format), rfc1951 (deflate format) and
// rfc1952 (gzip format).
//
// Installation
//
// 	$ go get modernc.org/z
//
// Linking using ccgo
//
// 	$ ccgo foo.c -lmodernc.org/z/lib
//
// Documentation
//
// 	http://godoc.org/modernc.org/z
//
// Builders
//
// 	https://modern-c.appspot.com/-/builder/?importpath=modernc.org%2fz
package z // import "modernc.org/z"

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

func origin(skip int) string {
	pc, fn, fl, _ := runtime.Caller(skip)
	f := runtime.FuncForPC(pc)
	var fns string
	if f != nil {
		fns = f.Name()
		if x := strings.LastIndex(fns, "."); x > 0 {
			fns = fns[x+1:]
		}
	}
	return fmt.Sprintf("%s:%d:%s", fn, fl, fns)
}

func todo(s string, args ...interface{}) string {
	switch {
	case s == "":
		s = fmt.Sprintf(strings.Repeat("%v ", len(args)), args...)
	default:
		s = fmt.Sprintf(s, args...)
	}
	r := fmt.Sprintf("%s: TODO %s", origin(2), s) //TODOOK
	fmt.Fprintf(os.Stdout, "%s\n", r)
	os.Stdout.Sync()
	return r
}

func trc(s string, args ...interface{}) string {
	switch {
	case s == "":
		s = fmt.Sprintf(strings.Repeat("%v ", len(args)), args...)
	default:
		s = fmt.Sprintf(s, args...)
	}
	r := fmt.Sprintf("%s: TRC %s", origin(2), s)
	fmt.Fprintf(os.Stdout, "%s\n", r)
	os.Stdout.Sync()
	return r
}
