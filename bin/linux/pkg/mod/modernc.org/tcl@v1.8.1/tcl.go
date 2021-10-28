// Copyright 2020 The Tcl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate go run generator.go
//go:generate assets -package tcl -dir
//go:generate go fmt ./...

// Package tcl is a CGo-free port of the Tool Command Language (Tcl).
//
// Tcl is a very powerful but easy to learn dynamic programming language,
// suitable for a very wide range of uses, including web and desktop
// applications, networking, administration, testing and many more.
//
// A separate Tcl shell is in the gotclsh directory.
//
// Changelog:
//
// 2020-09-13 v1.4.0 supports linux/{amd64,386,arm,arm64}. The arm, arm64 ports
// fail the http tests.
//
// 2020-09-03 v1.2.0 is now completelely CGo-free.
//
// 2020-08-04: beta2 released for linux/amd64 only. Support for threads,
// sockets and fork is not yet implemented. Some tests still crash, those are
// disabled at the moment.
package tcl // import "modernc.org/tcl"

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"modernc.org/httpfs"
	"modernc.org/libc"
	"modernc.org/tcl/lib"
)

const (
	// #define TCL_VOLATILE		((Tcl_FreeProc *) 1)
	tclVolatile = 1
)

var (
	fToken   uintptr
	libOnce  sync.Once
	objectMu sync.Mutex
	objects  = map[uintptr]interface{}{}

	_ = trc
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

func todo(s string, args ...interface{}) string { //TODO-
	switch {
	case s == "":
		s = fmt.Sprintf(strings.Repeat("%v ", len(args)), args...)
	default:
		s = fmt.Sprintf(s, args...)
	}
	r := fmt.Sprintf("%s: TODOTODO %s", origin(2), s) //TODOOK
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
	r := fmt.Sprintf("\n%s: TRC %s", origin(2), s)
	fmt.Fprintf(os.Stdout, "%s\n", r)
	os.Stdout.Sync()
	return r
}

func token() uintptr { return atomic.AddUintptr(&fToken, 1) }

func addObject(o interface{}) uintptr {
	t := token()
	objectMu.Lock()
	objects[t] = o
	objectMu.Unlock()
	return t
}

func getObject(t uintptr) interface{} {
	objectMu.Lock()
	o := objects[t]
	if o == nil {
		panic(todo("%#x", t))
	}

	objectMu.Unlock()
	return o
}

func removeObject(t uintptr) {
	objectMu.Lock()
	if _, ok := objects[t]; !ok {
		panic(todo("%#x"))
	}

	delete(objects, t)
	objectMu.Unlock()
}

// LibraryFileSystem returns a http.FileSystem containing the Tcl library.
func LibraryFileSystem() http.FileSystem {
	return httpfs.NewFileSystem(assets, time.Now())
}

// Library writes the Tcl library to directory.
func Library(directory string) error {
	var a []string
	for k := range assets {
		a = append(a, k)
	}
	sort.Strings(a)
	dirs := map[string]struct{}{}
	for _, nm := range a {
		pth := filepath.Join(directory, filepath.FromSlash(nm))
		dir := pth
		if !strings.HasSuffix(nm, "/") {
			dir = filepath.Dir(pth)
		}
		if _, ok := dirs[dir]; !ok {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}

			dirs[dir] = struct{}{}
		}
		if strings.HasSuffix(nm, "/") {
			continue
		}

		f, err := os.Create(pth)
		if err != nil {
			return err
		}

		if _, err := f.Write([]byte(assets[nm])); err != nil {
			f.Close()
			return err
		}

		if err = f.Close(); err != nil {
			return err
		}
	}
	return nil
}

// CmdProc is a Tcl command implemented in Go.
type CmdProc func(clientData interface{}, in *Interp, args []string) int

// DeleteProc is a function called when CmdProc is deleted.
type DeleteProc func(clientData interface{})

// Command represents a Tcl command.
type Command struct {
	cmd uintptr
}

// Interp represents a Tcl interpreter.
type Interp struct {
	tls    *libc.TLS
	interp uintptr
}

// NewInterp returns a newly created Interp or an error, if any.
func NewInterp() (*Interp, error) {
	var err error
	libOnce.Do(func() {
		if os.Getenv("TCL_LIBRARY") != "" {
			return
		}

		var dir string
		if dir, err = ioutil.TempDir("", "tcl-library-"); err != nil {
			return
		}

		err = Library(dir)
	})
	if err != nil {
		return nil, err
	}

	tls := libc.NewTLS()
	interp := tcl.XTcl_CreateInterp(tls)
	if interp == 0 {
		tls.Close()
		return nil, fmt.Errorf("failed to create Tcl interpreter")
	}

	return &Interp{tls, interp}, nil
}

// MustNewInterp is like NewInterp but panics on error.
func MustNewInterp() *Interp {
	in, err := NewInterp()
	if err != nil {
		panic(err)
	}

	return in
}

// Handle returns the handle of the underlying interpreter. It is used when
// calling libtcl directly. For example:
//
//	import (
//		"modernc.org/libc"
//		"modernc.org/tcl"
//		libtcl "modernc.org/tcl/lib"
//	)
//
//	in := tcl.MustNewInterp()
//	returnCode := libtcl.XTcl_Eval(in.TLS(), in.Handle(), "set a 42")
//	returnString := libc.GoString(libtcl.XTcl_GetStringResult(in.TLS(), in.Handle()))
func (in *Interp) Handle() uintptr { return in.interp }

// TLS returns the thread local storage of the underlying interpreter. It is
// used when calling libtcl directly. For example:
//
//	import (
//		"modernc.org/libc"
//		"modernc.org/tcl"
//		libtcl "modernc.org/tcl/lib"
//	)
//
//	in := tcl.MustNewInterp()
//	returnCode := libtcl.XTcl_Eval(in.TLS(), in.Handle(), "set a 42")
//	returnString := libc.GoString(libtcl.XTcl_GetStringResult(in.TLS(), in.Handle()))
func (in *Interp) TLS() *libc.TLS { return in.tls }

// Close invalidates the interpreter and releases all its associated resources.
func (in *Interp) Close() error {
	tcl.XTcl_DeleteInterp(in.tls, in.interp)
	in.tls.Close()
	in.tls = nil
	in.interp = 0
	return nil
}

// MustClose is like Close but panics on error.
func (in *Interp) MustClose() {
	if err := in.Close(); err != nil {
		panic(err)
	}
}

// Eval evaluates script and returns the interpreter; result and error, if any.
func (in *Interp) Eval(script string) (string, error) {
	s, err := libc.CString(script)
	if err != nil {
		return "", err
	}

	tcl.XTcl_Preserve(in.tls, in.interp)

	defer func() {
		libc.Xfree(in.tls, s)
		tcl.XTcl_Release(in.tls, in.interp)
	}()

	rc := tcl.XTcl_Eval(in.tls, in.interp, s)
	rs := libc.GoString(tcl.XTcl_GetStringResult(in.tls, in.interp))
	if rc == tcl.TCL_OK {
		return rs, nil
	}

	return rs, fmt.Errorf("return code: %d", rc)
}

// MustEval is like Eval but panics on error.
func (in *Interp) MustEval(script string) string {
	s, err := in.Eval(script)
	if err != nil {
		panic(err)
	}

	return s
}

type cmdProc struct {
	clientData interface{}
	del        DeleteProc
	f          CmdProc
	in         *Interp
}

func runCmd(tls *libc.TLS, clientData, in uintptr, argc int32, argv uintptr) int32 {
	cmd := getObject(clientData).(*cmdProc)
	var a []string
	for i := int32(0); i < argc; i++ {
		p := *(*uintptr)(unsafe.Pointer(argv)) //TODOOK
		argv += unsafe.Sizeof(argv)
		a = append(a, libc.GoString(p))
	}
	return int32(cmd.f(cmd.clientData, cmd.in, a))
}

func delCmd(tls *libc.TLS, clientData uintptr) {
	cmd := getObject(clientData).(*cmdProc)
	if cmd.del != nil {
		cmd.del(cmd.clientData)
	}
	removeObject(clientData)
}

var (
	runCmdP = *(*uintptr)(unsafe.Pointer(&struct {
		f func(tls *libc.TLS, clientData, in uintptr, argc int32, argv uintptr) int32
	}{runCmd}))
	delCmdP = *(*uintptr)(unsafe.Pointer(&struct {
		f func(tls *libc.TLS, clientData uintptr)
	}{delCmd}))
)

// NewCommand returns a newly created Tcl command or an error, if any.
func (in *Interp) NewCommand(name string, proc CmdProc, clientData interface{}, del DeleteProc) (*Command, error) {
	nm, err := libc.CString(name)
	if err != nil {
		return nil, err
	}

	tcl.XTcl_Preserve(in.tls, in.interp)

	defer func() {
		libc.Xfree(in.tls, nm)
		tcl.XTcl_Release(in.tls, in.interp)
	}()

	p := &cmdProc{f: proc, clientData: clientData, del: del, in: in}
	h := addObject(p)
	cmd := tcl.XTcl_CreateCommand(in.tls, in.interp, nm, runCmdP, h, delCmdP)
	if cmd == 0 {
		return nil, fmt.Errorf("failed to create command: %s", name)
	}

	return &Command{cmd}, nil
}

// MustNewCommand is like NewCommand but panics on error.
func (in *Interp) MustNewCommand(name string, proc CmdProc, clientData interface{}, del DeleteProc) *Command {
	cmd, err := in.NewCommand(name, proc, clientData, del)
	if err != nil {
		panic(err)
	}

	return cmd
}

// SetResult sets the result of the interpreter.
func (in *Interp) SetResult(s string) error {
	cs, err := libc.CString(s)
	if err != nil {
		return err
	}

	defer libc.Xfree(in.tls, cs)

	tcl.XTcl_SetResult(in.tls, in.interp, cs, tclVolatile)
	return nil
}
