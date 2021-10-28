// Copyright 2020 The Tcl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tcl

import (
	"modernc.org/libc"
)

func XTcl_MakeFileChannel(tls *libc.TLS, rawHandle ClientData, mode int32) Tcl_Channel { /* tclWinChan.c:1050:1: */
	panic("TODO")
}

func sDoRenameFile(tls *libc.TLS, nativeSrc uintptr, nativeDst uintptr) int32 { /* tclWinFCmd.c:153:1: */
	panic("TODO")
}

func sDoCopyFile(tls *libc.TLS, nativeSrc uintptr, nativeDst uintptr) int32 { /* tclWinFCmd.c:542:1: */
	panic("TODO")
}
