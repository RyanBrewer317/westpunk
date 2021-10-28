// Copyright 2020 The Tcl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tclsh

import (
	"modernc.org/libc"
)

func Main(tls *libc.TLS, argc int32, argv uintptr) int32 {
	return main(tls, argc, argv)
}
