// Copyright 2021 The Tcl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Parts of code and or documentation in this file is copied/translated from
// Tcl C code and subject to a BSD-style license found in the license.terms
// file.

package tcl // import "modernc.org/tcl"

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
	"unsafe"

	"modernc.org/httpfs"
	"modernc.org/libc"
	"modernc.org/libc/errno"
	"modernc.org/mathutil"
	"modernc.org/tcl/lib"
)

const (
	tclChannelVersion_2   = 2
	tclFilesystemVersion1 = 1
	vfsName               = "govfs"
)

var (
	_               = copy(cVFSName[:], vfsName)
	cVFSName        [len(vfsName) + 1]byte
	vfsIsRegistered bool
	vfsMounts       = map[string]*fileSystem{}
	vfsMu           sync.Mutex
	vfsPoints       []string
)

type fileSystem struct {
	files map[string]string
	*httpfs.FileSystem
}

func newVFS(files map[string]string) *fileSystem {
	return &fileSystem{files, httpfs.NewFileSystem(files, time.Now())}
}

// MountFileSystem mounts a virtual file system at point, which should be an
// absolute, slash separated path. The map keys must be rooted unix
// slash-separated paths. The file content is whatever the associated map value
// is. The resulting path of the VFS is the join of point and the map key.
func MountFileSystem(point string, files map[string]string) error {
	vfsMu.Lock()

	defer vfsMu.Unlock()

	point, err := normalizeMountPoint(point)
	if err != nil {
		return err
	}

	if !vfsIsRegistered {
		tls := libc.NewTLS()

		defer tls.Close()

		if rc := tcl.XTcl_FSRegister(tls, 0, uintptr(unsafe.Pointer(&vfs))); rc != tcl.TCL_OK {
			return fmt.Errorf("virtual file system initialization failed: %d", rc)
		}

		vfsIsRegistered = true
	}

	lockedUnmountFileSystem(point)
	vfsMounts[point] = newVFS(files)
	vfsPoints = append(vfsPoints, point)
	sort.Strings(vfsPoints)
	return nil
}

// UnmountFileSystem unmounts a virtual file system at point, which should be
// an absolute, slash separated path.
func UnmountFileSystem(point string) error {
	vfsMu.Lock()

	defer vfsMu.Unlock()

	return lockedUnmountFileSystem(point)
}

func lockedUnmountFileSystem(point string) error {
	point, err := normalizeMountPoint(point)
	if err != nil {
		return err
	}

	if vfsMounts[point] == nil {
		return fmt.Errorf("no file system mounted: %q", point)
	}

	i := sort.Search(len(vfsPoints), func(i int) bool { return vfsPoints[i] >= point })
	vfsPoints = append(vfsPoints[:i], vfsPoints[i+1:]...)
	delete(vfsMounts, point)
	return nil
}

func normalizeMountPoint(s string) (string, error) {
	s = path.Clean(s)
	if s == "." {
		return "", fmt.Errorf("invalid file system mount point: %s", s)
	}

	if s != "/" {
		s += "/"
	}

	return s, nil
}

var vfs = tcl.Tcl_Filesystem{
	FtypeName:        uintptr(unsafe.Pointer(&cVFSName[0])),
	FstructureLength: int32(unsafe.Sizeof(tcl.Tcl_Filesystem{})),
	Fversion:         tclFilesystemVersion1,
	FpathInFilesystemProc: *(*uintptr)(unsafe.Pointer(&struct {
		f func(tls *libc.TLS, pathPtr uintptr, clientDataPtr uintptr) int32
	}{vfsPathInFilesystem})),
	FstatProc: *(*uintptr)(unsafe.Pointer(&struct {
		f func(tls *libc.TLS, pathPtr uintptr, bufPtr uintptr) int32
	}{vfsStat})),
	FaccessProc: *(*uintptr)(unsafe.Pointer(&struct {
		f func(tls *libc.TLS, pathPtr uintptr, mode int32) int32
	}{vfsAccess})),
	FopenFileChannelProc: *(*uintptr)(unsafe.Pointer(&struct {
		f func(tls *libc.TLS, interp uintptr, pathPtr uintptr, mode int32, permissions int32) tcl.Tcl_Channel
	}{vfsOpenFileChannel})),
	FmatchInDirectoryProc: *(*uintptr)(unsafe.Pointer(&struct {
		f func(tls *libc.TLS, interp uintptr, resultPtr uintptr, pathPtr uintptr, pattern uintptr, types1 uintptr) int32
	}{vfsMatchInDirectory})),
}

// The pathInFilesystemProc field contains the address of a function which is
// called to determine whether a given path value belongs to this filesystem or
// not. Tcl will only call the rest of the filesystem functions with a path for
// which this function has returned TCL_OK. If the path does not belong, -1
// should be returned (the behavior of Tcl for any other return value is not
// defined). If TCL_OK is returned, then the optional clientDataPtr output
// parameter can be used to return an internal (filesystem specific)
// representation of the path, which will be cached inside the path value, and
// may be retrieved efficiently by the other filesystem functions. Tcl will
// simultaneously cache the fact that this path belongs to this filesystem.
// Such caches are invalidated when filesystem structures are added or removed
// from Tcl's internal list of known filesystems.
func vfsPathInFilesystem(tls *libc.TLS, pathPtr uintptr, clientDataPtr uintptr) int32 {
	path := path.Clean(libc.GoString(tcl.XTcl_GetString(tls, pathPtr)))

	vfsMu.Lock()

	defer vfsMu.Unlock()

	if file := vfsFile(path); file != nil {
		return tcl.TCL_OK
	}

	return -1
}

// Function to process a Tcl_FSAccess call. Must be implemented for any
// reasonable filesystem, since many Tcl level commands depend crucially upon
// it (e.g. file exists, file readable).
//
// The Tcl_FSAccessProc checks whether the process would be allowed to read,
// write or test for existence of the file (or other filesystem object) whose
// name is in pathPtr. If the pathname refers to a symbolic link, then the
// permissions of the file referred by this symbolic link should be tested.
//
// On success (all requested permissions granted), zero is returned. On error
// (at least one bit in mode asked for a permission that is denied, or some
// other error occurred), -1 is returned.
func vfsAccess(tls *libc.TLS, pathPtr uintptr, mode int32) int32 {
	vfsMu.Lock()

	defer vfsMu.Unlock()

	fi := vfsFileInfo(path.Clean(libc.GoString(tcl.XTcl_GetString(tls, pathPtr))))
	if fi == nil {
		return -1
	}

	switch {
	case fi.IsDir():
		if mode&0222 != 0 { // deny write
			return -1
		}
	default:
		if mode&0333 != 0 { // deny write, exec
			return -1
		}
	}

	return 0
}

// Function to process a Tcl_FSOpenFileChannel call. Must be implemented for
// any reasonable filesystem, since any operations which require open or
// accessing a file's contents will use it (e.g. open, encoding, and many Tk
// commands).
//
// The Tcl_FSOpenFileChannelProc opens a file specified by pathPtr and returns
// a channel handle that can be used to perform input and output on the file.
// This API is modeled after the fopen procedure of the Unix standard I/O
// library. The syntax and meaning of all arguments is similar to those given
// in the Tcl open command when opening a file, where the mode argument is a
// combination of the POSIX flags O_RDONLY, O_WRONLY, etc. If an error occurs
// while opening the channel, the Tcl_FSOpenFileChannelProc returns NULL and
// records a POSIX error code that can be retrieved with Tcl_GetErrno. In
// addition, if interp is non-NULL, the Tcl_FSOpenFileChannelProc leaves an
// error message in interp's result after any error.
//
// The newly created channel must not be registered in the supplied interpreter
// by a Tcl_FSOpenFileChannelProc; that task is up to the caller of
// Tcl_FSOpenFileChannel (if necessary). If one of the standard channels,
// stdin, stdout or stderr was previously closed, the act of creating the new
// channel also assigns it as a replacement for the standard channel.
func vfsOpenFileChannel(tls *libc.TLS, interp uintptr, pathPtr uintptr, mode int32, permissions int32) tcl.Tcl_Channel {
	vfsMu.Lock()

	defer vfsMu.Unlock()

	cPath := tcl.XTcl_GetString(tls, pathPtr)
	path := path.Clean(libc.GoString(cPath))
	file := vfsFile(path)
	if file == nil {
		return 0
	}

	return tcl.XTcl_CreateChannel(tls, uintptr(unsafe.Pointer(&channel)), cPath, addObject(file), tcl.TCL_READABLE)
}

// Function to process a Tcl_FSMatchInDirectory call. If not implemented, then
// glob and recursive copy functionality will be lacking in the filesystem (and
// this may impact commands like encoding names which use glob functionality
// internally).
//
// The function should return all files or directories (or other filesystem
// objects) which match the given pattern and accord with the types
// specification given. There are two ways in which this function may be
// called. If pattern is NULL, then pathPtr is a full path specification of a
// single file or directory which should be checked for existence and correct
// type. Otherwise, pathPtr is a directory, the contents of which the function
// should search for files or directories which have the correct type. In
// either case, pathPtr can be assumed to be both non-NULL and non-empty. It is
// not currently documented whether pathPtr will have a file separator at its
// end of not, so code should be flexible to both possibilities.
//
// The return value is a standard Tcl result indicating whether an error
// occurred in the matching process. Error messages are placed in interp,
// unless interp in NULL in which case no error message need be generated; on a
// TCL_OK result, results should be added to the resultPtr value given (which
// can be assumed to be a valid unshared Tcl list). The matches added to
// resultPtr should include any path prefix given in pathPtr (this usually
// means they will be absolute path specifications). Note that if no matches
// are found, that simply leads to an empty result; errors are only signaled
// for actual file or filesystem problems which may occur during the matching
// process.
//
// The Tcl_GlobTypeData structure passed in the types parameter contains the
// following fields:
//
//	typedef struct Tcl_GlobTypeData {
//		/* Corresponds to bcdpfls as in 'find -t' */
//		int type;
//		/* Corresponds to file permissions */
//		int perm;
//		/* Acceptable mac type */
//		Tcl_Obj *macType;
//		/* Acceptable mac creator */
//		Tcl_Obj *macCreator;
//	} Tcl_GlobTypeData;
//
// There are two specific cases which it is important to handle correctly, both
// when types is non-NULL. The two cases are when types->types &
// TCL_GLOB_TYPE_DIR or types->types & TCL_GLOB_TYPE_MOUNT are true (and in
// particular when the other flags are false). In the first of these cases, the
// function must list the contained directories. Tcl uses this to implement
// recursive globbing, so it is critical that filesystems implement directory
// matching correctly. In the second of these cases, with TCL_GLOB_TYPE_MOUNT,
// the filesystem must list the mount points which lie within the given pathPtr
// (and in this case, pathPtr need not lie within the same filesystem -
// different to all other cases in which this function is called). Support for
// this is critical if Tcl is to have seamless transitions between from one
// filesystem to another.
func vfsMatchInDirectory(tls *libc.TLS, interp uintptr, resultPtr uintptr, pathPtr uintptr, pattern uintptr, types1 uintptr) int32 {
	vfsMu.Lock()

	defer vfsMu.Unlock()

	pth := path.Clean(libc.GoString(tcl.XTcl_GetString(tls, pathPtr)))
	// If pattern is NULL, then pathPtr is a full path specification of a
	// single file or directory which should be checked for existence and correct
	// type.
	if pattern == 0 || *(*byte)(unsafe.Pointer(pattern)) == 0 {
		if types1 != 0 {
			switch typ := (*tcl.Tcl_GlobTypeData)(unsafe.Pointer(types1)).Ftype; typ {
			default:
				panic(todo("%q: %#x", pth, typ))
			}
		}

		file := vfsFile(pth)
		if file == nil {
			return tcl.TCL_ERROR
		}

		tcl.XTcl_ListObjAppendElement(tls, interp, resultPtr, pathPtr)
		return tcl.TCL_OK
	}

	var dirOnly bool
	if types1 != 0 {
		switch typ := (*tcl.Tcl_GlobTypeData)(unsafe.Pointer(types1)).Ftype; typ {
		case tcl.TCL_GLOB_TYPE_DIR:
			dirOnly = true
		case tcl.TCL_GLOB_TYPE_MOUNT:
			var a []string
			for _, v := range vfsPoints {
				if strings.HasPrefix(pth, v) {
					a = append(a, v)
				}
			}

			for _, v := range a {
				cs, err := libc.CString(v)
				if err != nil {
					return tcl.TCL_ERROR
				}

				if tcl.XTcl_StringCaseMatch(tls, cs, pattern, 0) != 0 {
					item := tcl.XTcl_NewStringObj(tls, cs, int32(len(v)))
					tcl.XTcl_ListObjAppendElement(tls, interp, resultPtr, item)
				}

				libc.Xfree(tls, cs)
			}
			return tcl.TCL_OK
		default:
			panic(todo("%q: %#x", pth, typ))
		}
	}

	// Otherwise, pathPtr is a directory, the contents of which the function
	// should search for files or directories which have the correct type.
	if !strings.HasSuffix(pth, "/") {
		pth += "/"
	}
	file := vfsFile(pth)
	if file == nil {
		return tcl.TCL_ERROR
	}

	if _, err := file.Stat(); err != nil {
		return tcl.TCL_ERROR
	}

	fis, err := file.Readdir(-1)
	if err != nil {
		return tcl.TCL_ERROR
	}

	for _, fi := range fis {
		if dirOnly && !fi.IsDir() {
			continue
		}

		s := path.Join(pth, fi.Name())
		cs, err := libc.CString(s)
		if err != nil {
			return tcl.TCL_ERROR
		}

		if tcl.XTcl_StringCaseMatch(tls, cs, pattern, 0) != 0 {
			item := tcl.XTcl_NewStringObj(tls, cs, int32(len(s)))
			tcl.XTcl_ListObjAppendElement(tls, interp, resultPtr, item)
		}

		libc.Xfree(tls, cs)
	}
	return 0
}

func vfsFile(path string) http.File {
	point, fs := findVFS(path)
	if fs == nil {
		return nil
	}

	abs := path[len(point)-1:]
	file, err := fs.Open(abs)
	if err != nil {
		if !strings.HasSuffix(abs, "/") {
			file, err = fs.Open(abs + "/")
		}
	}
	if err != nil {
		return nil
	}

	return file
}

func vfsFileInfo(path string) os.FileInfo {
	file := vfsFile(path)
	if file == nil {
		return nil
	}

	fi, err := file.Stat()
	if err != nil {
		return nil
	}

	return fi
}

func findVFS(path string) (string, *fileSystem) {
	if len(vfsPoints) == 0 {
		return "", nil
	}

	i := sort.Search(len(vfsPoints), func(i int) bool { return vfsPoints[i] >= path })
	if point, fs := vfsMatch(path, i); fs != nil {
		return point, fs
	}

	if point, fs := vfsMatch(path, i-1); fs != nil {
		return point, fs
	}

	return "", nil
}

func vfsMatch(path string, i int) (string, *fileSystem) {
	if i >= 0 && i < len(vfsPoints) {
		if strings.HasPrefix(path, vfsPoints[i]) {
			point := vfsPoints[i]
			return point, vfsMounts[point]
		}

		if !strings.HasSuffix(path, "/") && strings.HasPrefix(path+"/", vfsPoints[i]) {
			point := vfsPoints[i]
			return point, vfsMounts[point]
		}
	}

	return "", nil
}

var channel = tcl.Tcl_ChannelType{
	FtypeName: uintptr(unsafe.Pointer(&cVFSName[0])),
	Fversion:  tclChannelVersion_2,
	FcloseProc: *(*uintptr)(unsafe.Pointer(&struct {
		f func(tls *libc.TLS, instanceData tcl.ClientData, interp uintptr) int32
	}{channelClose})),
	FinputProc: *(*uintptr)(unsafe.Pointer(&struct {
		f func(tls *libc.TLS, instanceData tcl.ClientData, buf uintptr, toRead int32, errorCodePtr uintptr) int32
	}{channelInput})),
	FseekProc: *(*uintptr)(unsafe.Pointer(&struct {
		f func(tls *libc.TLS, instanceData tcl.ClientData, offset int64, mode int32, errorCodePtr uintptr) int32
	}{channelSeek})),
	FwatchProc: *(*uintptr)(unsafe.Pointer(&struct {
		f func(tls *libc.TLS, instanceData tcl.ClientData, mask int32)
	}{channelWatch})),
	FwideSeekProc: *(*uintptr)(unsafe.Pointer(&struct {
		f func(tls *libc.TLS, instanceData tcl.ClientData, offset tcl.Tcl_WideInt, mode int32, errorCodePtr uintptr) tcl.Tcl_WideInt
	}{channelWideSeek})),
}

// The closeProc field contains the address of a function called by the generic
// layer to clean up driver-related information when the channel is closed.
// CloseProc must match the following prototype:
//
// The instanceData argument is the same as the value provided to
// Tcl_CreateChannel when the channel was created. The function should release
// any storage maintained by the channel driver for this channel, and close the
// input and output devices encapsulated by this channel. All queued output
// will have been flushed to the device before this function is called, and no
// further driver operations will be invoked on this instance after calling the
// closeProc. If the close operation is successful, the procedure should return
// zero; otherwise it should return a nonzero POSIX error code. In addition, if
// an error occurs and interp is not NULL, the procedure should store an error
// message in the interpreter's result.
func channelClose(tls *libc.TLS, instanceData tcl.ClientData, interp uintptr) int32 {
	removeObject(instanceData)
	return 0
}

// The inputProc field contains the address of a function called by the generic
// layer to read data from the file or device and store it in an internal
// buffer. InputProc must match the following prototype:
//
// InstanceData is the same as the value passed to Tcl_CreateChannel when the
// channel was created. The buf argument points to an array of bytes in which
// to store input from the device, and the bufSize argument indicates how many
// bytes are available at buf.
//
// The errorCodePtr argument points to an integer variable provided by the
// generic layer. If an error occurs, the function should set the variable to a
// POSIX error code that identifies the error that occurred.
//
// The function should read data from the input device encapsulated by the
// channel and store it at buf. On success, the function should return a
// nonnegative integer indicating how many bytes were read from the input
// device and stored at buf. On error, the function should return -1. If an
// error occurs after some data has been read from the device, that data is
// lost.
//
// If inputProc can determine that the input device has some data available but
// less than requested by the bufSize argument, the function should only
// attempt to read as much data as is available and return without blocking. If
// the input device has no data available whatsoever and the channel is in
// nonblocking mode, the function should return an EAGAIN error. If the input
// device has no data available whatsoever and the channel is in blocking mode,
// the function should block for the shortest possible time until at least one
// byte of data can be read from the device; then, it should return as much
// data as it can read without blocking.
//
// This value can be retrieved with Tcl_ChannelInputProc, which returns a
// pointer to the function.
func channelInput(tls *libc.TLS, instanceData tcl.ClientData, buf uintptr, toRead int32, errorCodePtr uintptr) int32 {
	if buf == 0 || toRead == 0 {
		return 0
	}

	n, err := getObject(instanceData).(http.File).Read((*libc.RawMem)(unsafe.Pointer(buf))[:toRead:toRead])
	if n != 0 {
		return int32(n)
	}

	if err != nil && err != io.EOF {
		return -1
	}

	return 0
}

// The seekProc field contains the address of a function called by the generic
// layer to move the access point at which subsequent input or output
// operations will be applied. SeekProc must match the following prototype:
//
// The instanceData argument is the same as the value given to
// Tcl_CreateChannel when this channel was created. Offset and seekMode have
// the same meaning as for the Tcl_Seek procedure (described in the manual
// entry for Tcl_OpenFileChannel).
//
// The errorCodePtr argument points to an integer variable provided by the
// generic layer for returning errno values from the function. The function
// should set this variable to a POSIX error code if an error occurs. The
// function should store an EINVAL error code if the channel type does not
// implement seeking.
//
// The return value is the new access point or -1 in case of error. If an error
// occurred, the function should not move the access point.
func channelSeek(tls *libc.TLS, instanceData tcl.ClientData, offset int64, mode int32, errorCodePtr uintptr) (r int32) {
	e := int32(errno.EINVAL)
	defer func() {
		if r < 0 && errorCodePtr != 0 {
			*(*int32)(unsafe.Pointer(errorCodePtr)) = e
		}
	}()

	if offset < mathutil.MinInt || offset > mathutil.MaxInt {
		return -1
	}

	file := getObject(instanceData).(http.File)
	n0, err := file.Seek(0, os.SEEK_CUR)
	if err != nil {
		return -1
	}

	n, err := file.Seek(offset, int(mode))
	if err != nil {
		return -1
	}

	if n > math.MaxInt32 {
		e = errno.EOVERFLOW
		file.Seek(n0, os.SEEK_SET)
		return -1
	}

	return int32(n)
}

// If there is a non-NULL seekProc field, the wideSeekProc field may contain
// the address of an alternative function to use which handles wide (i.e.
// larger than 32-bit) offsets, so allowing seeks within files larger than 2GB.
// The wideSeekProc will be called in preference to the seekProc, but both must
// be defined if the wideSeekProc is defined. WideSeekProc must match the
// following prototype:
//
// The arguments and return values mean the same thing as with seekProc above,
// except that the type of offsets and the return type are different.
//
// The seekProc value can be retrieved with Tcl_ChannelSeekProc, which returns
// a pointer to the function, and similarly the wideSeekProc can be retrieved
// with Tcl_ChannelWideSeekProc.
func channelWideSeek(tls *libc.TLS, instanceData tcl.ClientData, offset tcl.Tcl_WideInt, mode int32, errorCodePtr uintptr) tcl.Tcl_WideInt {
	file := getObject(instanceData).(http.File)
	n, err := file.Seek(offset, int(mode))
	if err != nil {
		if errorCodePtr != 0 {
			*(*int32)(unsafe.Pointer(errorCodePtr)) = errno.EINVAL
		}
		return -1
	}

	return tcl.Tcl_WideInt(n)
}

// The watchProc field contains the address of a function called by the generic
// layer to initialize the event notification mechanism to notice events of
// interest on this channel. WatchProc should match the following prototype:
//
// The instanceData is the same as the value passed to Tcl_CreateChannel when
// this channel was created. The mask argument is an OR-ed combination of
// TCL_READABLE, TCL_WRITABLE and TCL_EXCEPTION; it indicates events the caller
// is interested in noticing on this channel.
//
// The function should initialize device type specific mechanisms to notice
// when an event of interest is present on the channel. When one or more of the
// designated events occurs on the channel, the channel driver is responsible
// for calling Tcl_NotifyChannel to inform the generic channel module. The
// driver should take care not to starve other channel drivers or sources of
// callbacks by invoking Tcl_NotifyChannel too frequently. Fairness can be
// insured by using the Tcl event queue to allow the channel event to be
// scheduled in sequence with other events. See the description of
// Tcl_QueueEvent for details on how to queue an event.
//
// This value can be retrieved with Tcl_ChannelWatchProc, which returns a
// pointer to the function.
func channelWatch(tls *libc.TLS, instanceData tcl.ClientData, mask int32) {
	if mask != 0 {
		panic(todo(""))
	}
}
