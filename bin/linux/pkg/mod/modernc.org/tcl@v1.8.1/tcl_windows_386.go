// Copyright 2021 The Tcl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Parts of code and or documentation in this file is copied/translated from
// Tcl C code and subject to a BSD-style license found in the license.terms
// file.

package tcl // import "modernc.org/tcl"

import (
	"unsafe"

	"modernc.org/libc"
	"modernc.org/libc/sys/types"
	"modernc.org/libc/time"
	"modernc.org/tcl/lib"
)

// Function to process a Tcl_FSStat call. Must be implemented for any
// reasonable filesystem, since many Tcl level commands depend crucially upon
// it (e.g. file atime, file isdirectory, file size, glob).

// The Tcl_FSStatProc fills the stat structure statPtr with information about
// the specified file. You do not need any access rights to the file to get
// this information but you need search rights to all directories named in the
// path leading to the file. The stat structure includes info regarding device,
// inode (always 0 on Windows), privilege mode, nlink (always 1 on Windows),
// user id (always 0 on Windows), group id (always 0 on Windows), rdev (same as
// device on Windows), size, last access time, last modification time, and last
// metadata change time.
//
// If the file represented by pathPtr exists, the Tcl_FSStatProc returns 0 and
// the stat structure is filled with data. Otherwise, -1 is returned, and no
// stat info is given.
func vfsStat(tls *libc.TLS, pathPtr uintptr, bufPtr uintptr) int32 {
	vfsMu.Lock()

	defer vfsMu.Unlock()

	fi := vfsFileInfo(libc.GoString(tcl.XTcl_GetString(tls, pathPtr)))
	if fi == nil {
		return -1
	}

	tm := time.Time_t(fi.ModTime().Unix())
	*(*tcl.Tcl_StatBuf)(unsafe.Pointer(bufPtr)) = tcl.Tcl_StatBuf{
		Fst_atime: tm,
		Fst_ctime: tm,
		Fst_mode:  types.Mode_t(fi.Mode()),
		Fst_mtime: tm,
		Fst_size:  types.Off_t(fi.Size()),
	}
	return 0
}

// MountLibraryVFS mounts the Tcl library virtual file system and returns the
// mount point. This is how it's used, for example, in gotclsh:
//
//	package main
//
//	import (
//		"os"
//
//		"modernc.org/libc"
//		"modernc.org/tcl"
//		"modernc.org/tcl/internal/tclsh"
//	)
//
//	const envVar = "TCL_LIBRARY"
//
//	func main() {
//		if os.Getenv(envVar) == "" {
//			if s, err := tcl.MountLibraryVFS(); err == nil {
//				os.Setenv(envVar, s)
//			}
//		}
//		libc.Start(tclsh.Main)
//	}
func MountLibraryVFS() (string, error) {
	//TODO point := tcl.TCL_LIBRARY
	point := "c:\\govfs\\tcl_library"
	if err := MountFileSystem(point, assets); err != nil {
		return "", err
	}

	return point, nil
}
