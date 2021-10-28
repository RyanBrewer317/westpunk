// Copyright 2020 The Tcl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tcl

import (
	"modernc.org/libc"
)

func XTclWinCPUID(tls *libc.TLS, index uint32, regsPtr uintptr) int32 {
	return 1
}
