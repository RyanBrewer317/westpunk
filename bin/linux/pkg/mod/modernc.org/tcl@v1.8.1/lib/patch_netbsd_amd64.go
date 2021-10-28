// Copyright 2021 The Tcl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tcl

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"unsafe"

	"modernc.org/libc"
	"modernc.org/libc/errno"
)

func todo(s string, args ...interface{}) string { //TODO-
	switch {
	case s == "":
		s = fmt.Sprintf(strings.Repeat("%v ", len(args)), args...)
	default:
		s = fmt.Sprintf(s, args...)
	}
	pc, fn, fl, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	var fns string
	if f != nil {
		fns = f.Name()
		if x := strings.LastIndex(fns, "."); x > 0 {
			fns = fns[x+1:]
		}
	}
	r := fmt.Sprintf("%s:%d:%s: TODOTODO %s", fn, fl, fns, s) //TODOOK
	fmt.Fprintf(os.Stdout, "%s\n", r)
	os.Stdout.Sync()
	return r
}

func trc(s string, args ...interface{}) string { //TODO-
	switch {
	case s == "":
		s = fmt.Sprintf(strings.Repeat("%v ", len(args)), args...)
	default:
		s = fmt.Sprintf(s, args...)
	}
	_, fn, fl, _ := runtime.Caller(1)
	r := fmt.Sprintf("\n%s:%d: TRC %s", fn, fl, s)
	fmt.Fprintf(os.Stdout, "%s\n", r)
	os.Stdout.Sync()
	return r
}

var createProcessMsg = [...]byte{'%', 's', ':', ' ', '%', 's', 0}

func XTclpCreateProcess(tls *libc.TLS, interp uintptr, argc int32, argv uintptr, inputFile, outputFile, errorFile, pidPtr uintptr) int32 { /* tclUnixPipe.c:379:1: */
	bp := tls.Alloc(2 * 8)
	defer tls.Free(2 * 8)
	var args []string
	for i := 0; i < int(argc); i++ {
		p := *(*uintptr)(unsafe.Pointer(argv + unsafe.Sizeof(uintptr(0))*uintptr(i)))
		args = append(args, libc.GoString(p))
	}
	if len(args) == 0 {
		panic(todo(""))
	}

	args0, err := exec.LookPath(args[0])
	if err != nil {
		*(*int32)(unsafe.Pointer(libc.X__errno_location(tls))) = errno.ENOENT
		s, err := libc.CString(fmt.Sprintf("couldn't execute \"%.150s\"", args[0]))
		if err != nil {
			panic(todo(""))
		}

		defer libc.Xfree(tls, s)

		XTcl_SetObjResult(tls, interp, XTcl_ObjPrintf(tls, uintptr(unsafe.Pointer(&createProcessMsg[0])), libc.VaList(bp, s, XTcl_PosixError(tls, interp))))
		return TCL_ERROR
	}

	args[0] = args0
	env := libc.GetEnviron()
	attr := &syscall.ProcAttr{
		Env:   env,
		Files: []uintptr{^uintptr(0), ^uintptr(0), ^uintptr(0)},
	}
	if inputFile != 0 {
		attr.Files[syscall.Stdin] = inputFile - 1 // tclUnixPipe.c:27: #define GetFd(file) (PTR2INT(file) - 1)
	}
	if outputFile != 0 {
		attr.Files[syscall.Stdout] = outputFile - 1
	}
	if errorFile != 0 {
		attr.Files[syscall.Stderr] = errorFile - 1
	}
	pid, err := syscall.ForkExec(args0, args, attr)
	if err != nil {
		trc("TclpCreateProcess(%#x, %d, %q, %v, %v, %v, %#x): %v", interp, len(args), args, inputFile, outputFile, errorFile, pidPtr, err)
		panic(todo(""))
	}

	// trc("TclpCreateProcess(%#x, %d, %q, %v, %v, %v, %#x): pid %v", interp, len(args), args, inputFile, outputFile, errorFile, pidPtr, pid)
	*(*uintptr)(unsafe.Pointer(pidPtr)) = uintptr(pid)
	return TCL_OK
}
