// Copyright 2017 The HTTPFS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package httpfs implements http.FileSystem on top of a map[string]string.
package httpfs // import "modernc.org/httpfs"

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

var (
	_ = http.File((*file)(nil))             //TODOOK
	_ = http.FileSystem((*FileSystem)(nil)) //TODOOK
	_ = os.FileInfo((*file)(nil))           //TODOOK
)

// FileSystem is an implementation of http.FileSystem.
type FileSystem struct {
	files   map[string]string
	modTime time.Time
}

// NewFileSystem returns a new FileSystem containing http.Files from the files
// argument.  The map keys must be rooted unix slash-separated paths. The file
// content is whatever the associated map value is. All files will have their
// modTime set to modTime.
func NewFileSystem(files map[string]string, modTime time.Time) *FileSystem {
	return &FileSystem{files, modTime}
}

// Open implements http.FileSystem.
func (f *FileSystem) Open(name string) (fi http.File, err error) {
	if strings.HasSuffix(name, "/") {
		dir := make([]string, 0) // Must be non-nil.
		m := map[string]struct{}{}
		ok := false
		for k := range f.files {
			if strings.HasPrefix(k, name) {
				ok = true
				k = k[len(name):]
				parts := strings.Split(k, "/")
				name := parts[0]
				if name == "" {
					continue
				}

				switch {
				case len(parts) == 1: // file
					dir = append(dir, name)
				default: // dir
					if _, ok := m[name]; ok {
						break
					}

					m[name] = struct{}{}
					dir = append(dir, name+"/")
				}
			}
		}
		if !ok {
			return nil, fmt.Errorf("no such directory: %q", name)
		}

		sort.Strings(dir)
		//fmt.Printf("%q -> %v %v\n", name, len(dir), dir) //TODO-
		return &file{FileSystem: f, name: name, dir: dir, mode: os.ModeDir}, nil
	}

	s, ok := f.files[name]
	if !ok {
		return nil, fmt.Errorf("no such file: %q", name)
	}

	return &file{FileSystem: f, name: name, s: s}, nil
}

type file struct {
	*FileSystem
	dir  []string
	mode os.FileMode
	name string
	off  int64
	s    string
}

func (f *file) Close() (err error)               { return nil }
func (f *file) IsDir() (r bool)                  { return strings.HasSuffix(f.name, "/") }
func (f *file) Mode() (r os.FileMode)            { return f.mode }
func (f *file) ModTime() (r time.Time)           { return f.modTime }
func (f *file) Name() (r string)                 { return f.name }
func (f *file) Size() (r int64)                  { return int64(len(f.s)) }
func (f *file) Stat() (r os.FileInfo, err error) { return f, nil }
func (f *file) Sys() (r interface{})             { return nil }

func (f *file) Read(b []byte) (n int, err error) {
	if f.off >= int64(len(f.s)) {
		return 0, io.EOF
	}

	n = copy(b, f.s[int(f.off):])
	f.off += int64(n)
	return n, nil
}

func (f *file) Readdir(count int) (r []os.FileInfo, err error) {
	for _, fn := range f.dir {
		count--
		if count == 0 {
			break
		}

		name := path.Join(f.name, fn)
		if strings.HasSuffix(fn, "/") {
			name += "/"
		}
		s := f.files[name]
		r = append(r, &file{name: name, s: s})
	}
	return r, err
}

func (f *file) Seek(offset int64, whence int) (r int64, err error) {
	off := f.off
	switch whence {
	case os.SEEK_SET:
		f.off = offset
	case os.SEEK_CUR:
		f.off += offset
	case os.SEEK_END:
		f.off = int64(len(f.s)) + offset
	}
	if f.off < 0 || f.off > int64(len(f.s)) {
		err = fmt.Errorf("invalid offset (%#x/%#x, whence %v)", f.off, len(f.s), whence)
		f.off = off
	}
	return f.off, err
}

func (f *file) String() string {
	return fmt.Sprintf("{name %q, dir %v, off %v, size %v}", f.name, f.dir, f.off, len(f.s))
}
