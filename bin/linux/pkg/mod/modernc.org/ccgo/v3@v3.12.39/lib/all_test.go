// Copyright 2020 The CCGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ccgo // import "modernc.org/ccgo/v3/lib"

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/dustin/go-humanize"
	"modernc.org/cc/v3"
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

func stack() string { return string(debug.Stack()) }

func use(...interface{}) {}

func init() {
	use(caller, dbg, TODO, stack) //TODOOK
}

// ----------------------------------------------------------------------------

var (
	oBlackBox   = flag.String("blackbox", "", "Record CSmith file to this file")
	oCSmith     = flag.Duration("csmith", 15*time.Minute, "")
	oCpp        = flag.Bool("cpp", false, "Amend compiler errors with preprocessor output")
	oDebug      = flag.Bool("debug", false, "")
	oKeep       = flag.Bool("keep", false, "keep temp directories")
	oRE         = flag.String("re", "", "")
	oStackTrace = flag.Bool("trcstack", false, "")
	oTrace      = flag.Bool("trc", false, "Print tested paths.")
	oTraceF     = flag.Bool("trcf", false, "Print test file content")
	oTraceO     = flag.Bool("trco", false, "Print test output")
	oXTags      = flag.String("xtags", "", "passed to go build of TestSQLite")
	writeFailed = flag.Bool("write-failed", false, "Write all failed tests into a file called FAILED in the cwd, in the format of go maps for easy copy-pasting.")

	gccDir    = filepath.FromSlash("testdata/gcc-9.1.0")
	sqliteDir = filepath.FromSlash("testdata/sqlite-amalgamation-3330000")
	tccDir    = filepath.FromSlash("testdata/tcc-0.9.27")

	testWD string

	csmithDefaultArgs = strings.Join([]string{
		"--bitfields",                     // --bitfields | --no-bitfields: enable | disable full-bitfields structs (disabled by default).
		"--max-nested-struct-level", "10", // --max-nested-struct-level <num>: limit maximum nested level of structs to <num>(default 0). Only works in the exhaustive mode.
		"--no-const-pointers",    // --const-pointers | --no-const-pointers: enable | disable const pointers (enabled by default).
		"--no-consts",            // --consts | --no-consts: enable | disable const qualifier (enabled by default).
		"--no-packed-struct",     // --packed-struct | --no-packed-struct: enable | disable packed structs by adding #pragma pack(1) before struct definition (disabled by default).
		"--no-volatile-pointers", // --volatile-pointers | --no-volatile-pointers: enable | disable volatile pointers (enabled by default).
		"--no-volatiles",         // --volatiles | --no-volatiles: enable | disable volatiles (enabled by default).
		"--paranoid",             // --paranoid | --no-paranoid: enable | disable pointer-related assertions (disabled by default).
	}, " ")
)

func TestMain(m *testing.M) {
	var rc int
	defer func() {
		if err := recover(); err != nil {
			rc = 1
			fmt.Fprintf(os.Stderr, "PANIC: %v\n%s\n", err, debug.Stack())
		}
		os.Exit(rc)
	}()

	// fmt.Printf("test binary compiled for %s/%s\n", runtime.GOOS, runtime.GOARCH)
	// fmt.Printf("temp dir: %s\n", os.TempDir()) //TODO-
	// if s := os.Getenv("CCGO_CPP"); s != "" {
	// 	fmt.Printf("CCGO_CPP=%s\n", os.Getenv("CCGO_CPP"))
	// }

	flag.BoolVar(&oTraceW, "trcw", false, "Print generator writes")
	flag.BoolVar(&oTraceG, "trcg", false, "Print generator output")
	flag.BoolVar(&oTracePin, "trcpin", false, "Print pinning")
	flag.Parse()
	var err error
	if testWD, err = os.Getwd(); err != nil {
		panic("Cannot determine working dir: " + err.Error())
	}

	rc = m.Run()
}

type golden struct {
	t *testing.T
	f *os.File
	w *bufio.Writer
}

func newGolden(t *testing.T, fn string) *golden {
	if *oRE != "" {
		return &golden{w: bufio.NewWriter(ioutil.Discard)}
	}

	f, err := os.Create(filepath.FromSlash(fn))
	if err != nil { // Possibly R/O fs in a VM
		base := filepath.Base(filepath.FromSlash(fn))
		f, err = ioutil.TempFile("", base)
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("writing results to %s\n", f.Name())
	}

	w := bufio.NewWriter(f)
	return &golden{t, f, w}
}

func (g *golden) close() {
	if g.f == nil {
		return
	}

	if err := g.w.Flush(); err != nil {
		g.t.Fatal(err)
	}

	if err := g.f.Close(); err != nil {
		g.t.Fatal(err)
	}
}

func h(v interface{}) string {
	switch x := v.(type) {
	case int:
		return humanize.Comma(int64(x))
	case int64:
		return humanize.Comma(x)
	case uint64:
		return humanize.Comma(int64(x))
	case float64:
		return humanize.CommafWithDigits(x, 0)
	default:
		panic(fmt.Errorf("%T", x)) //TODOOK
	}
}

func TestTCC(t *testing.T) {
	root := filepath.Join(testWD, filepath.FromSlash(tccDir))
	g := newGolden(t, fmt.Sprintf("testdata/tcc_%s_%s.golden", runtime.GOOS, runtime.GOARCH))

	defer g.close()

	var files, ok int
	const dir = "tests/tests2"
	f, o := testTCCExec(g.w, t, filepath.Join(root, filepath.FromSlash(dir)))
	files += f
	ok += o
	t.Logf("files %s, ok %s", h(files), h(ok))
}

func testTCCExec(w io.Writer, t *testing.T, dir string) (files, ok int) {
	blacklist := map[string]struct{}{
		"34_array_assignment.c":       {}, // gcc: 16:6: error: assignment to expression with array type
		"60_errors_and_warnings.c":    {}, // Not a standalone test. gcc fails.
		"76_dollars_in_identifiers.c": {}, // `int $ = 10;` etc.
		"77_push_pop_macro.c":         {}, //
		"81_types.c":                  {}, // invalid function type cast
		"86_memory-model.c":           {},
		"93_integer_promotion.c":      {}, // The expected output does not agree with gcc.
		"95_bitfields.c":              {}, // Included from 95_bitfields_ms.c
		"95_bitfields_ms.c":           {}, // The expected output does not agree with gcc.
		"96_nodata_wanted.c":          {}, // Not a standalone test. gcc fails.
		"99_fastcall.c":               {}, // 386 only

		"40_stdio.c":                {}, //TODO
		"73_arm64.c":                {}, //TODO struct varargs
		"80_flexarray.c":            {}, //TODO Flexible array member
		"85_asm-outside-function.c": {}, //TODO
		"87_dead_code.c":            {}, //TODO expression statement
		"88_codeopt.c":              {}, //TODO expression statement
		"89_nocode_wanted.c":        {}, //TODO expression statement
		"90_struct-init.c":          {}, //TODO cc ... in designator
		"92_enum_bitfield.c":        {}, //TODO bit fields
		"94_generic.c":              {}, //TODO cc _Generic
		"98_al_ax_extend.c":         {}, //TODO
	}
	if runtime.GOARCH == "s390x" {
		blacklist["91_ptr_longlong_arith32.c"] = struct{}{} // OK, invalid result on big endian
	}
	if runtime.GOOS == "windows" {
		blacklist["46_grep.c"] = struct{}{} //TODO
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	defer os.Chdir(wd)

	temp, err := ioutil.TempDir("", "ccgo-test-")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(temp)

	if err := os.Chdir(temp); err != nil {
		t.Fatal(err)
	}

	if os.Getenv("GO111MODULE") != "off" {
		if out, err := Shell("go", "mod", "init", "example.com/ccgo/v3/lib/tcc"); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}

		if out, err := Shell("go", "get", "modernc.org/libc"); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}
	}

	var re *regexp.Regexp
	if s := *oRE; s != "" {
		re = regexp.MustCompile(s)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	failed := make([]string, 0, 0)
	success := make([]string, 0, 0)
	limiter := make(chan int, runtime.GOMAXPROCS(0))
	// fill the limiter
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		limiter <- i
	}
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				err = nil
			}
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".c" || info.Mode()&os.ModeType != 0 {
			return nil
		}

		if _, ok := blacklist[filepath.Base(path)]; ok {
			return nil
		}

		files++

		if re != nil && !re.MatchString(path) {
			return nil
		}

		main_file, err := ioutil.TempFile(temp, "*.go")
		if err != nil {
			return nil
		}
		main := main_file.Name()
		main_file.Close()
		wg.Add(1)
		go func(id int) {
			if *oTrace {
				fmt.Fprintln(os.Stderr, path)
			}
			var ret bool

			defer func() {
				mu.Lock()
				if ret {
					ok++
					success = append(success, filepath.Base(path))
				} else {
					failed = append(failed, filepath.Base(path))
				}
				mu.Unlock()
				limiter <- id
				wg.Done()
			}()

			ccgoArgs := []string{
				"ccgo",

				"-all-errors",
				"-hide", "__sincosf,__sincos,__sincospif,__sincospi",
				"-o", main,
				"-verify-structs",
			}
			var args []string
			switch base := filepath.Base(path); base {
			case "31_args.c":
				args = []string{"arg1", "arg2", "arg3", "arg4", "arg5"}
			case "46_grep.c":
				if err := copyFile(path, filepath.Join(temp, base)); err != nil {
					t.Error(err)
					return
				}

				args = []string{`[^* ]*[:a:d: ]+\:\*-/: $`, base}
			}

			ret = testSingle(t, main, path, ccgoArgs, args)
		}(<-limiter)
		return nil
	}); err != nil {
		t.Errorf("%v", err)
	}

	wg.Wait()
	sort.Strings(failed)
	sort.Strings(success)
	if *writeFailed {
		failedFile, _ := os.Create("FAILED")
		for _, fpath := range failed {
			failedFile.WriteString("\"")
			failedFile.WriteString(fpath)
			failedFile.WriteString("\": {},\n")
		}
	}
	for _, fpath := range success {
		w.Write([]byte(fpath))
		w.Write([]byte{'\n'})
	}

	return len(failed) + len(success), len(success)
}

func cpp(enabled bool, args []string, err0 error) error {
	if !enabled {
		return err0
	}

	args = append(args, "-E")
	var out bytes.Buffer
	if err := NewTask(args, &out, &out).Main(); err != nil {
		return fmt.Errorf("error while acquiring preprocessor output: %v\n%v", err, err0)
	}

	return fmt.Errorf("preprocessor output:\n%s\n%v", out.Bytes(), err0)
}

func trim(b []byte) (r []byte) {
	b = bytes.ReplaceAll(b, []byte{'\r'}, nil)
	b = bytes.TrimLeft(b, "\n")
	b = bytes.TrimRight(b, "\n")
	a := bytes.Split(b, []byte("\n"))
	for i, v := range a {
		a[i] = bytes.TrimRight(v, " ")
	}
	return bytes.Join(a, []byte("\n"))
}

func noExt(s string) string {
	ext := filepath.Ext(s)
	if ext == "" {
		panic("internal error") //TODOOK
	}
	return s[:len(s)-len(ext)]
}

func copyFile(src, dst string) error {
	b, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(dst, b, 0660)
}

func skipDir(path string) error {
	if strings.HasPrefix(filepath.Base(path), ".") {
		return filepath.SkipDir
	}

	return nil
}

func TestCAPI(t *testing.T) {
	task := NewTask(nil, nil, nil)
	pkgName, capi, err := task.capi("modernc.org/libc")
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := capi["printf"]; !ok {
		t.Fatal("default libc does not export printf")
	}

	t.Log(pkgName, capi)
}

const text = "abcd\nefgh\x00ijkl"

var (
	text1 = text
	text2 = (*reflect.StringHeader)(unsafe.Pointer(&text1)).Data
)

func TestText(t *testing.T) {
	p := text2
	var b []byte
	for i := 0; i < len(text); i++ {
		b = append(b, *(*byte)(unsafe.Pointer(p)))
		p++
	}
	if g, e := string(b), text; g != e {
		t.Fatalf("%q %q", g, e)
	}
}

func TestGCCExec(t *testing.T) {
	root := filepath.Join(testWD, filepath.FromSlash(gccDir))
	g := newGolden(t, fmt.Sprintf("testdata/gcc_exec_%s_%s.golden", runtime.GOOS, runtime.GOARCH))

	defer g.close()

	var files, ok int
	const dir = "gcc/testsuite/gcc.c-torture/execute"
	f, o := testGCCExec(g.w, t, filepath.Join(root, filepath.FromSlash(dir)), false)
	files += f
	ok += o
	t.Logf("files %s, ok %s", h(files), h(ok))
}

func testGCCExec(w io.Writer, t *testing.T, dir string, opt bool) (files, ok int) {
	blacklist := map[string]struct{}{
		// nested func
		"20000822-1.c": {}, // nested func
		"20010209-1.c": {},
		"20010605-1.c": {},
		"20030501-1.c": {},

		// asm
		"20001009-2.c": {},
		"20020107-1.c": {},
		"20030222-1.c": {},
		"960830-1.c":   {},

		// unsupported alignment
		"20010904-1.c": {},
		"20010904-2.c": {},
		"align-3.c":    {},

		"20010122-1.c": {}, // __builtin_return_address
		"20021127-1.c": {}, // gcc specific optimization
		"20101011-1.c": {}, // sigfpe
		"991014-1.c":   {}, // Struct type too big
		"eeprof-1.c":   {}, // requires instrumentation

		// unsupported volatile declarator size
		"20021120-1.c": {},
		"20030128-1.c": {},
		"pr53160.c":    {},
		"pr71631.c":    {},
		"pr83269.c":    {},
		"pr89195.c":    {},

		// implementation defined conversion result
		"20031003-1.c": {},

		// goto * expr
		"20040302-1.c":  {},
		"comp-goto-1.c": {},
		"comp-goto-2.c": {},

		//TODO initializing zero sized fields not supported
		"zero-struct-2.c": {},

		//TODO flexible array member
		"20010924-1.c": {},
		"20030109-1.c": {},
		"20050613-1.c": {},
		"pr28865.c":    {},
		"pr33382.c":    {},

		//TODO _Complex
		"20041124-1.c": {},
		"20041201-1.c": {},
		"20010605-2.c": {},
		"20020227-1.c": {},
		"20020411-1.c": {},
		"20030910-1.c": {},

		//TODO bitfield
		"20000113-1.c": {},

		//TODO designator
		"20000801-3.c": {},

		//TODO __builtin_types_compatible_p
		"20020206-2.c": {},

		//TODO alloca
		"20020314-1.c": {},
		"20021113-1.c": {},
		"20040223-1.c": {},

		//TODO statement expression
		"20020320-1.c": {},

		//TODO VLA
		"20040308-1.c": {},
		"20040411-1.c": {},
		"20040423-1.c": {},

		//TODO link error
		"fp-cmp-7.c": {},

		//TODO __builtin_isunordered
		"compare-fp-1.c": {},
		"compare-fp-3.c": {}, //TODO
		"compare-fp-4.c": {},
		"fp-cmp-4.c":     {}, //TODO
		"fp-cmp-4f.c":    {},
		"fp-cmp-4l.c":    {},
		"fp-cmp-5.c":     {},
		"fp-cmp-8.c":     {},
		"fp-cmp-8f.c":    {},
		"fp-cmp-8l.c":    {},
		"pr38016.c":      {},

		//TODO __builtin_infl
		"inf-1.c":   {},
		"inf-2.c":   {},
		"pr36332.c": {}, //TODO

		//TODO __builtin_huge_vall
		"inf-3.c": {},

		//TODO undefined: tanf
		"mzero4.c": {},

		//TODO __builtin_isgreater
		"pr50310.c": {},

		//TODO convert: TODOTODO t1 -> t2
		"pr72824-2.c": {},

		//TODO struct var arg
		"20020412-1.c": {},

		//TODO undefined: tmpnam
		"fprintf-2.c":   {},
		"printf-2.c":    {},
		"user-printf.c": {},

		"20040520-1.c":                 {}, //TODO
		"20040629-1.c":                 {}, //TODO
		"20040705-1.c":                 {}, //TODO
		"20040705-2.c":                 {}, //TODO
		"20040707-1.c":                 {}, //TODO
		"20040709-1.c":                 {}, //TODO
		"20040709-2.c":                 {}, //TODO
		"20040709-3.c":                 {}, //TODO
		"20041011-1.c":                 {}, //TODO 48:1: unsupported volatile declarator size: 128
		"20041214-1.c":                 {}, //TODO
		"20041218-2.c":                 {}, //TODO
		"20050121-1.c":                 {}, //TODO
		"20050316-1.c":                 {}, //TODO
		"20050316-2.c":                 {}, //TODO
		"20050316-3.c":                 {}, //TODO
		"20050604-1.c":                 {}, //TODO
		"20050607-1.c":                 {}, //TODO
		"20050929-1.c":                 {}, //TODO
		"20051012-1.c":                 {}, //TODO
		"20060420-1.c":                 {}, //TODO sizeof vector
		"20061220-1.c":                 {}, //TODO
		"20070614-1.c":                 {}, //TODO
		"20070824-1.c":                 {}, //TODO
		"20070919-1.c":                 {}, //TODO
		"20071210-1.c":                 {}, //TODO
		"20071211-1.c":                 {}, //TODO
		"20071220-1.c":                 {}, //TODO
		"20071220-2.c":                 {}, //TODO
		"20080502-1.c":                 {}, //TODO
		"20090219-1.c":                 {}, //TODO
		"20100430-1.c":                 {}, //TODO unsupported attribute: packed
		"20180921-1.c":                 {}, //TODO
		"920302-1.c":                   {}, //TODO
		"920415-1.c":                   {}, //TODO
		"920428-2.c":                   {}, //TODO
		"920501-1.c":                   {}, //TODO
		"920501-3.c":                   {}, //TODO
		"920501-4.c":                   {}, //TODO
		"920501-5.c":                   {}, //TODO
		"920501-7.c":                   {}, //TODO
		"920612-2.c":                   {}, //TODO
		"920625-1.c":                   {}, //TODO
		"920721-4.c":                   {}, //TODO
		"920908-1.c":                   {}, //TODO
		"921017-1.c":                   {}, //TODO
		"921202-1.c":                   {}, //TODO
		"921208-2.c":                   {}, //TODO
		"921215-1.c":                   {}, //TODO
		"930406-1.c":                   {}, //TODO
		"931002-1.c":                   {}, //TODO
		"931004-10.c":                  {}, //TODO
		"931004-12.c":                  {}, //TODO
		"931004-14.c":                  {}, //TODO
		"931004-2.c":                   {}, //TODO
		"931004-4.c":                   {}, //TODO
		"931004-6.c":                   {}, //TODO
		"931004-8.c":                   {}, //TODO
		"941014-1.c":                   {}, //TODO
		"941202-1.c":                   {}, //TODO
		"960312-1.c":                   {}, //TODO
		"960416-1.c":                   {}, //TODO
		"960512-1.c":                   {}, //TODO
		"970217-1.c":                   {}, //TODO VLA paramater
		"980526-1.c":                   {}, //TODO
		"990130-1.c":                   {}, //TODO
		"990208-1.c":                   {}, //TODO
		"990413-2.c":                   {}, //TODO
		"990524-1.c":                   {}, //TODO
		"991112-1.c":                   {}, //TODO
		"991227-1.c":                   {}, //TODO
		"alias-2.c":                    {}, //TODO
		"alias-3.c":                    {}, //TODO
		"alias-4.c":                    {}, //TODO
		"align-nest.c":                 {}, //TODO
		"alloca-1.c":                   {}, //TODO
		"anon-1.c":                     {}, //TODO nested field access
		"bcp-1.c":                      {}, //TODO
		"bitfld-3.c":                   {}, //TODO
		"built-in-setjmp.c":            {}, //TODO
		"builtin-bitops-1.c":           {}, //TODO
		"builtin-constant.c":           {}, //TODO
		"builtin-prefetch-3.c":         {}, //TODO volatile struct
		"builtin-types-compatible-p.c": {}, //TODO
		"call-trap-1.c":                {}, //TODO
		"complex-1.c":                  {}, //TODO
		"complex-2.c":                  {}, //TODO
		"complex-4.c":                  {}, //TODO
		"complex-5.c":                  {}, //TODO
		"complex-6.c":                  {}, //TODO
		"complex-7.c":                  {}, //TODO
		"ffs-1.c":                      {}, //TODO
		"ffs-2.c":                      {}, //TODO
		"frame-address.c":              {}, //TODO
		"medce-1.c":                    {}, //TODO
		"nest-align-1.c":               {}, //TODO
		"nest-stdar-1.c":               {}, //TODO
		"nestfunc-1.c":                 {}, //TODO
		"nestfunc-2.c":                 {}, //TODO
		"nestfunc-3.c":                 {}, //TODO
		"nestfunc-5.c":                 {}, //TODO
		"nestfunc-6.c":                 {}, //TODO
		"nestfunc-7.c":                 {}, //TODO
		"pr17377.c":                    {}, //TODO
		"pr22061-1.c":                  {}, //TODO
		"pr22061-3.c":                  {}, //TODO
		"pr22061-4.c":                  {}, //TODO
		"pr23135.c":                    {}, //TODO
		"pr23324.c":                    {}, //TODO
		"pr23467.c":                    {}, //TODO
		"pr24135.c":                    {}, //TODO
		"pr28289.c":                    {}, //TODO
		"pr34154.c":                    {}, //TODO
		"pr35456.c":                    {}, //TODO
		"pr36321.c":                    {}, //TODO
		"pr37780.c":                    {}, //TODO
		"pr38151.c":                    {}, //TODO
		"pr38533.c":                    {}, //TODO
		"pr38969.c":                    {}, //TODO
		"pr39228.c":                    {}, //TODO
		"pr40022.c":                    {}, //TODO
		"pr40657.c":                    {}, //TODO
		"pr41239.c":                    {}, //TODO
		"pr41935.c":                    {}, //TODO
		"pr42248.c":                    {}, //TODO
		"pr42570":                      {}, //TODO uint8_t foo[1][0];
		"pr43385.c":                    {}, //TODO
		"pr43560.c":                    {}, //TODO
		"pr44575.c":                    {}, //TODO
		"pr45695.c":                    {}, //TODO
		"pr46309.c":                    {}, //TODO
		"pr47237.c":                    {}, //TODO
		"pr49279.c":                    {}, //TODO
		"pr49390.c":                    {}, //TODO
		"pr49644.c":                    {}, //TODO
		"pr51447.c":                    {}, //TODO
		"pr51877.c":                    {}, //TODO
		"pr51933.c":                    {}, //TODO
		"pr52286.c":                    {}, //TODO
		"pr53645-2.c":                  {}, //TODO
		"pr53645.c":                    {}, //TODO
		"pr55750.c":                    {}, //TODO
		"pr56205.c":                    {}, //TODO
		"pr56837.c":                    {}, //TODO
		"pr56866.c":                    {}, //TODO
		"pr56982.c":                    {}, //TODO
		"pr57344-1.c":                  {}, //TODO
		"pr57344-2.c":                  {}, //TODO
		"pr57344-3.c":                  {}, //TODO
		"pr57344-4.c":                  {}, //TODO
		"pr60003.c":                    {}, //TODO
		"pr60960.c":                    {}, //TODO
		"pr61725.c":                    {}, //TODO
		"pr63641.c":                    {}, //TODO
		"pr64006.c":                    {}, //TODO
		"pr64242.c":                    {}, //TODO
		"pr65053-2.c":                  {}, //TODO
		"pr65427.c":                    {}, //TODO
		"pr65648.c":                    {}, //TODO
		"pr65956.c":                    {}, //TODO
		"pr66556.c":                    {}, //TODO unsupported volatile declarator size: 237
		"pr67037.c":                    {}, //TODO
		"pr68249.c":                    {}, //TODO
		"pr68328.c":                    {}, //TODO
		"pr68381.c":                    {}, //TODO
		"pr69320-2.c":                  {}, //TODO
		"pr70460.c":                    {}, //TODO
		"pr70903.c":                    {}, //TODO
		"pr71494.c":                    {}, //TODO
		"pr71554.c":                    {}, //TODO
		"pr71626-1.c":                  {}, //TODO
		"pr71626-2.c":                  {}, //TODO
		"pr77767.c":                    {}, //TODO VLA parameter
		"pr78438.c":                    {}, //TODO
		"pr78726.c":                    {}, //TODO
		"pr79354.c":                    {}, //TODO
		"pr79737-2.c":                  {}, //TODO
		"pr80421.c":                    {}, //TODO
		"pr80692.c":                    {}, //TODO
		"pr81588.c":                    {}, //TODO
		"pr82210.c":                    {}, //TODO
		"pr82954.c":                    {}, //TODO
		"pr84478.c":                    {}, //TODO
		"pr84524.c":                    {}, //TODO
		"pr85156.c":                    {}, //TODO
		"pr85169.c":                    {}, //TODO
		"pr85331.c":                    {}, //TODO
		"pr85529-1.c":                  {}, //TODO :24:5: unsupported volatile declarator type: volatile struct S
		"pr86528.c":                    {}, //TODO
		"pr89434.c":                    {}, //TODO
		"pushpop_macro.c":              {}, //TODO #pragma push_macro("_")
		"scal-to-vec1.c":               {}, //TODO
		"scal-to-vec2.c":               {}, //TODO
		"scal-to-vec3.c":               {}, //TODO
		"simd-1.c":                     {}, //TODO
		"simd-2.c":                     {}, //TODO
		"simd-4.c":                     {}, //TODO
		"simd-5.c":                     {}, //TODO
		"simd-6.c":                     {}, //TODO
		"stdarg-3.c":                   {}, //TODO
		"stkalign.c":                   {}, //TODO
		"strct-stdarg-1.c":             {}, //TODO
		"strct-varg-1.c":               {}, //TODO
		"string-opt-18.c":              {}, //TODO
		"string-opt-5.c":               {}, //TODO
		"va-arg-22.c":                  {}, //TODO
		"va-arg-pack-1.c":              {}, //TODO

	}
	if runtime.GOOS == "windows" && runtime.GOARCH == "amd64" {
		blacklist["fp-cmp-1.c"] = struct{}{} // needs signal.h
		blacklist["fp-cmp-2.c"] = struct{}{} // needs signal.h
		blacklist["fp-cmp-3.c"] = struct{}{} // needs signal.h
		blacklist["pr36339.c"] = struct{}{}  // typedef unsigned long my_uintptr_t;
	}
	if runtime.GOARCH == "386" {
		blacklist["rbug.c"] = struct{}{}     // https://github.com/golang/go/issues/48807
		blacklist["960830-1.c"] = struct{}{} // assembler statements not supported
	}
	if runtime.GOARCH == "arm" {
		blacklist["rbug.c"] = struct{}{} // https://github.com/golang/go/issues/48807
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	defer os.Chdir(wd)

	temp, err := ioutil.TempDir("", "ccgo-test-")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(temp)

	if err := os.Chdir(temp); err != nil {
		t.Fatal(err)
	}

	if os.Getenv("GO111MODULE") != "off" {
		if out, err := Shell("go", "mod", "init", "example.com/ccgo/v3/lib/gcc"); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}

		if out, err := Shell("go", "get", "modernc.org/libc"); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}
	}

	var re *regexp.Regexp
	if s := *oRE; s != "" {
		re = regexp.MustCompile(s)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	failed := make([]string, 0, 0)
	success := make([]string, 0, 0)
	limiter := make(chan int, runtime.GOMAXPROCS(0))
	// fill the limiter
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		limiter <- i
	}
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				err = nil
			}
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.Contains(filepath.ToSlash(path), "/builtins/") {
			return nil
		}

		if filepath.Ext(path) != ".c" || info.Mode()&os.ModeType != 0 {
			return nil
		}

		if _, ok := blacklist[filepath.Base(path)]; ok {
			return nil
		}

		files++

		if re != nil && !re.MatchString(path) {
			return nil
		}

		main_file, err := ioutil.TempFile(temp, "*.go")
		if err != nil {
			return nil
		}
		main := main_file.Name()
		main_file.Close()
		wg.Add(1)
		go func(id int) {
			if *oTrace {
				fmt.Fprintln(os.Stderr, path)
			}
			var ret bool

			defer func() {
				mu.Lock()
				if ret {
					ok++
					success = append(success, filepath.Base(path))
				} else {
					failed = append(failed, filepath.Base(path))
				}
				mu.Unlock()
				limiter <- id
				wg.Done()
			}()

			ccgoArgs := []string{
				"ccgo",

				"-D__FUNCTION__=__func__",
				"-export-defines", "",
				"-o", main,
				"-verify-structs",
			}

			ret = testSingle(t, main, path, ccgoArgs, nil)
		}(<-limiter)
		return nil
	}); err != nil {
		t.Errorf("%v", err)
	}

	wg.Wait()
	sort.Strings(failed)
	sort.Strings(success)
	if *writeFailed {
		failedFile, _ := os.Create("FAILED")
		for _, fpath := range failed {
			failedFile.WriteString("\"")
			failedFile.WriteString(fpath)
			failedFile.WriteString("\": {},\n")
		}
	}
	for _, fpath := range success {
		w.Write([]byte(fpath))
		w.Write([]byte{'\n'})
	}

	return len(failed) + len(success), len(success)
}

func TestSQLite(t *testing.T) {
	root := filepath.Join(testWD, filepath.FromSlash(sqliteDir))
	testSQLite(t, root)
}

func testSQLite(t *testing.T, dir string) {
	const main = "main.go"
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	defer os.Chdir(wd)

	temp, err := ioutil.TempDir("", "ccgo-test-")
	if err != nil {
		t.Fatal(err)
	}

	switch {
	case *oKeep:
		t.Log(temp)
	default:
		defer os.RemoveAll(temp)
	}

	if _, _, err := CopyDir(temp, dir, nil); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(temp); err != nil {
		t.Fatal(err)
	}

	ccgoArgs := []string{
		"ccgo",

		"-DHAVE_USLEEP",
		"-DLONGDOUBLE_TYPE=double",
		"-DSQLITE_DEBUG",
		"-DSQLITE_DEFAULT_MEMSTATUS=0",
		"-DSQLITE_ENABLE_DBPAGE_VTAB",
		"-DSQLITE_LIKE_DOESNT_MATCH_BLOBS",
		"-DSQLITE_MEMDEBUG",
		"-DSQLITE_THREADSAFE=0",
		"-all-errors",
		"-o", main,
		"-verify-structs",
		"shell.c",
		"sqlite3.c",
	}
	if *oDebug {
		ccgoArgs = append(ccgoArgs, "-DSQLITE_DEBUG_OS_TRACE", "-DSQLITE_FORCE_OS_TRACE")
	}
	if !func() (r bool) {
		defer func() {
			if err := recover(); err != nil {
				if *oStackTrace {
					fmt.Printf("%s\n", stack())
				}
				if *oTrace {
					fmt.Println(err)
				}
				t.Errorf("%v", err)
				r = false
			}
			if *oTraceF {
				b, _ := ioutil.ReadFile(main)
				fmt.Printf("\n----\n%s\n----\n", b)
			}
		}()

		if err := NewTask(ccgoArgs, nil, nil).Main(); err != nil {
			if *oTrace {
				fmt.Println(err)
			}
			err = cpp(*oCpp, ccgoArgs, err)
			t.Errorf("%v", err)
			return false
		}

		return true
	}() {
		return
	}

	if os.Getenv("GO111MODULE") != "off" {
		if out, err := Shell("go", "mod", "init", "example.com/ccgo/v3/lib/sqlite"); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}

		if out, err := Shell("go", "get", "modernc.org/libc"); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}
	}

	shell := "./shell"
	if runtime.GOOS == "windows" {
		shell = "shell.exe"
	}
	args := []string{"build"}
	if s := *oXTags; s != "" {
		args = append(args, "-tags", s)
	}
	args = append(args, "-o", shell, main)
	if out, err := exec.Command("go", args...).CombinedOutput(); err != nil {
		s := strings.TrimSpace(string(out))
		if s != "" {
			s += "\n"
		}
		t.Errorf("%s%v", s, err)
		return
	}

	var out []byte
	switch {
	case *oDebug:
		out, err = exec.Command(shell, "tmp", ".log stdout", "create table t(i); insert into t values(42); select 11*i from t;").CombinedOutput()
	default:
		out, err = exec.Command(shell, "tmp", "create table t(i); insert into t values(42); select 11*i from t;").CombinedOutput()
	}
	if err != nil {
		if *oTrace {
			fmt.Printf("%s\n%s\n", out, err)
		}
		t.Errorf("%s\n%v", out, err)
		return
	}

	if g, e := strings.TrimSpace(string(out)), "462"; g != e {
		t.Errorf("got: %s\nexp: %s", g, e)
	}
	if *oTraceO {
		fmt.Printf("%s\n", out)
	}

	if out, err = exec.Command(shell, "tmp", "select 13*i from t;").CombinedOutput(); err != nil {
		if *oTrace {
			fmt.Printf("%s\n%s\n", out, err)
		}
		t.Errorf("%v", err)
		return
	}

	if g, e := strings.TrimSpace(string(out)), "546"; g != e {
		t.Errorf("got: %s\nexp: %s", g, e)
	}
	if *oTraceO {
		fmt.Printf("%s\n", out)
	}
}

type compCertResult struct {
	compiler string
	test     string
	run      time.Duration
	k        float64

	compileOK bool
	execOK    bool
	resultOK  bool
}

func (r *compCertResult) String() string {
	var s string
	if r.k != 0 {
		s = fmt.Sprintf("%8.3f", r.k)
		if r.k == 1 {
			s += " *"
		}
	}
	return fmt.Sprintf("%10v%15v%10.3fms%6v%6v%6v%s", r.compiler, r.test, float64(r.run)/float64(time.Millisecond), r.compileOK, r.execOK, r.resultOK, s)
}

func TestCompCert(t *testing.T) {
	const root = "testdata/CompCert-3.6/test/c"

	b, err := ioutil.ReadFile(filepath.FromSlash(root + "/Results/knucleotide-input.txt"))
	if err != nil {
		t.Fatal(err)
	}

	dir := filepath.FromSlash(root)
	m, err := filepath.Glob(filepath.Join(dir, "*.c"))
	if err != nil {
		t.Fatal(err)
	}

	for i, v := range m {
		v, err := filepath.Abs(v)
		if err != nil {
			t.Fatal(err)
		}

		m[i] = v
	}

	rdir, err := filepath.Abs(filepath.FromSlash(root + "/Results"))
	if err != nil {
		t.Fatal(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	defer os.Chdir(wd)

	temp, err := ioutil.TempDir("", "ccgo-test-")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(temp)

	if err := os.Chdir(temp); err != nil {
		t.Fatal(err)
	}

	if os.Getenv("GO111MODULE") != "off" {
		if out, err := Shell("go", "mod", "init", "example.com/ccgo/v3/lib/compcert"); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}

		if out, err := Shell("go", "get", "modernc.org/libc"); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}
	}

	if err := os.Mkdir("Results", 0770); err != nil {
		t.Fatal(err)
	}

	if err := ioutil.WriteFile(filepath.FromSlash("Results/knucleotide-input.txt"), b, 0660); err != nil {
		t.Fatal(err)
	}

	var r []*compCertResult
	t.Run("gcc", func(t *testing.T) { r = append(r, testCompCertGcc(t, m, 5, rdir)...) })
	t.Run("ccgo", func(t *testing.T) { r = append(r, testCompCertCcgo(t, m, 5, rdir)...) })
	consider := map[string]struct{}{}
	for _, v := range r {
		consider[v.test] = struct{}{}
	}
	for _, v := range r {
		if ok := v.compileOK && v.execOK && v.resultOK; !ok {
			delete(consider, v.test)
		}
	}
	times := map[string]time.Duration{}
	tests := map[string][]*compCertResult{}
	for _, v := range r {
		if _, ok := consider[v.test]; !ok {
			continue
		}

		times[v.compiler] += v.run
		tests[v.test] = append(tests[v.test], v)
	}
	for _, a := range tests {
		if len(a) < 2 {
			continue
		}
		min := time.Duration(-1)
		for _, v := range a {
			if min < 0 || v.run < min {
				min = v.run
			}
		}
		for _, v := range a {
			v.k = float64(v.run) / float64(min)
		}
	}
	t.Log("  compiler           test    T         comp  exec match    coef")
	for _, v := range r {
		t.Log(v)
	}
	var a []string
	for k := range times {
		a = append(a, k)
	}
	sort.Strings(a)
	t.Logf("Considered tests: %d/%d", len(consider), len(m))
	min := time.Duration(-1)
	for _, v := range times {
		if min < 0 || v < min {
			min = v
		}
	}
	for _, k := range a {
		t.Logf("%10s%15v %6.3f", k, times[k], float64(times[k])/float64(min))
	}
}

func testCompCertGcc(t *testing.T, files []string, N int, rdir string) (r []*compCertResult) {
	blacklist := map[string]struct{}{}
	if runtime.GOOS == "windows" && runtime.GOARCH == "amd64" {
		blacklist["mandelbrot.c"] = struct{}{} //TODO
	}
	if runtime.GOOS == "linux" && runtime.GOARCH == "s390x" {
		blacklist["aes.c"] = struct{}{} // endian.h:7:1: "unknown endianness"
	}
	const nm = "gcc"
	var re *regexp.Regexp
	if s := *oRE; s != "" {
		re = regexp.MustCompile(s)
	}
next:
	for _, fn := range files {
		base := filepath.Base(fn)
		if *oTrace {
			fmt.Println(base)
		}
		if re != nil && !re.MatchString(base) {
			continue
		}

		if _, ok := blacklist[base]; ok {
			continue
		}

		bin := nm + "-" + base + ".out"
		out, err := exec.Command("gcc", "-O", "-o", bin, fn, "-lm").CombinedOutput()
		if err != nil {
			t.Errorf("%s: %s:\n%s", base, err, out)
			r = append(r, &compCertResult{nm, base, 0, 0, false, false, false})
			continue
		}

		t0 := time.Now()
		for i := 0; i < N; i++ {
			if out, err = exec.Command("./" + bin).CombinedOutput(); err != nil {
				t.Errorf("%s: %s:\n%s", base, err, out)
				r = append(r, &compCertResult{nm, base, 0, 0, true, false, false})
				continue next
			}
		}
		d := time.Since(t0) / time.Duration(N)
		isBinary := base == "mandelbrot.c"
		r = append(r, &compCertResult{nm, base, d, 0, true, true, checkResult(t, out, base, rdir, isBinary)})
	}
	return r
}

func checkResult(t *testing.T, out []byte, base, rdir string, bin bool) bool {
	base = base[:len(base)-len(filepath.Ext(base))]
	fn := filepath.Join(rdir, base)
	b, err := ioutil.ReadFile(fn)
	if err != nil {
		t.Errorf("%v: %v", base, err)
		return false
	}

	if !bin {
		out = bytes.ReplaceAll(out, []byte("\r"), nil)
		b = bytes.ReplaceAll(out, []byte("\r"), nil)
	}
	if bytes.Equal(out, b) {
		return true
	}

	fn2 := fn + "." + runtime.GOOS
	b2, err := ioutil.ReadFile(fn2)
	if err == nil {
		switch {
		case bytes.Equal(out, b2):
			return true
		default:
			t.Logf("got\n%s", hex.Dump(out))
			t.Logf("exp\n%s", hex.Dump(b2))
			t.Errorf("%v: result differs", base)
			return false
		}
	}

	t.Logf("got\n%s", hex.Dump(out))
	t.Logf("exp\n%s", hex.Dump(b))
	t.Errorf("%v: result differs", base)
	return false
}

func testCompCertCcgo(t *testing.T, files []string, N int, rdir string) (r []*compCertResult) {
	blacklist := map[string]struct{}{}
	if runtime.GOOS == "windows" && runtime.GOARCH == "amd64" {
		blacklist["knucleotide.c"] = struct{}{}
	}
	if runtime.GOOS == "linux" && runtime.GOARCH == "s390x" {
		blacklist["aes.c"] = struct{}{} // endian.h:7:1: "unknown endianness"
	}
	const nm = "ccgo"
	var re *regexp.Regexp
	if s := *oRE; s != "" {
		re = regexp.MustCompile(s)
	}
next:
	for _, fn := range files {
		base := filepath.Base(fn)
		if *oTrace {
			fmt.Println(base)
		}
		if re != nil && !re.MatchString(base) {
			continue
		}

		if _, ok := blacklist[base]; ok {
			continue
		}

		src := nm + "-" + base + ".go"
		bin := nm + "-" + base + ".out"
		var args []string
		if err := func() (err error) {
			defer func() {
				if e := recover(); e != nil && err == nil {
					if *oStackTrace {
						fmt.Printf("%s\n", stack())
					}
					err = fmt.Errorf("%v", e)
				}
			}()

			args = []string{
				"ccgo",

				"-o", src,
				fn,
			}
			return NewTask(args, nil, nil).Main()
		}(); err != nil {
			err = cpp(*oCpp, args, err)
			t.Errorf("%s: %s:", base, err)
			r = append(r, &compCertResult{nm, base, 0, 0, false, false, false})
			continue
		}
		if *oTraceF {
			b, _ := ioutil.ReadFile(src)
			fmt.Printf("\n----\n%s\n----\n", b)
		}

		if out, err := exec.Command("go", "build", "-o", bin, src).CombinedOutput(); err != nil {
			t.Errorf("%s: %s:\n%s", base, err, out)
			r = append(r, &compCertResult{nm, base, 0, 0, false, false, false})
			continue next
		}

		var out []byte
		t0 := time.Now()
		for i := 0; i < N; i++ {
			var err error
			if out, err = exec.Command("./" + bin).CombinedOutput(); err != nil {
				t.Errorf("%s: %s:\n%s", base, err, out)
				r = append(r, &compCertResult{nm, base, 0, 0, true, false, false})
				continue next
			}
		}
		d := time.Since(t0) / time.Duration(N)
		isBinary := base == "mandelbrot.c"
		r = append(r, &compCertResult{nm, base, d, 0, true, true, checkResult(t, out, base, rdir, isBinary)})
	}
	return r
}

func TestBug(t *testing.T) {
	g := newGolden(t, fmt.Sprintf("testdata/bug_%s_%s.golden", runtime.GOOS, runtime.GOARCH))

	defer g.close()

	var files, ok int
	f, o := testBugExec(g.w, t, filepath.Join(testWD, filepath.FromSlash("testdata/bug")))
	files += f
	ok += o
	t.Logf("files %s, ok %s", h(files), h(ok))
}

func testBugExec(w io.Writer, t *testing.T, dir string) (files, ok int) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	defer os.Chdir(wd)

	temp, err := ioutil.TempDir("", "ccgo-test-")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(temp)

	if err := os.Chdir(temp); err != nil {
		t.Fatal(err)
	}

	blacklist := map[string]struct{}{}
	if os.Getenv("GO111MODULE") != "off" {
		if out, err := Shell("go", "mod", "init", "example.com/ccgo/v3/lib/bug"); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}

		if out, err := Shell("go", "get", "modernc.org/libc"); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}
	}

	var re *regexp.Regexp
	if s := *oRE; s != "" {
		re = regexp.MustCompile(s)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	failed := make([]string, 0, 0)
	success := make([]string, 0, 0)
	limiter := make(chan int, runtime.GOMAXPROCS(0))
	// fill the limiter
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		limiter <- i
	}
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				err = nil
			}
			return err
		}

		if info.IsDir() {
			return skipDir(path)
		}

		if filepath.Ext(path) != ".c" || info.Mode()&os.ModeType != 0 {
			return nil
		}

		if _, ok := blacklist[filepath.Base(path)]; ok {
			return nil
		}

		files++

		if re != nil && !re.MatchString(path) {
			return nil
		}

		main_file, err := ioutil.TempFile(temp, "*.go")
		if err != nil {
			return nil
		}
		main := main_file.Name()
		main_file.Close()
		wg.Add(1)
		go func(id int) {
			if *oTrace {
				fmt.Fprintln(os.Stderr, path)
			}
			var ret bool

			defer func() {
				mu.Lock()
				if ret {
					ok++
					success = append(success, filepath.Base(path))
				} else {
					failed = append(failed, filepath.Base(path))
				}
				mu.Unlock()
				limiter <- id
				wg.Done()
			}()

			ccgoArgs := []string{
				"ccgo",

				"-export-defines", "",
				"-o", main,
				"-verify-structs",
			}

			ret = testSingle(t, main, path, ccgoArgs, nil)
		}(<-limiter)
		return nil
	}); err != nil {
		t.Errorf("%v", err)
	}

	wg.Wait()
	sort.Strings(failed)
	sort.Strings(success)
	if *writeFailed {
		failedFile, _ := os.Create("FAILED")
		for _, fpath := range failed {
			failedFile.WriteString("\"")
			failedFile.WriteString(fpath)
			failedFile.WriteString("\": {},\n")
		}
	}
	for _, fpath := range success {
		w.Write([]byte(fpath))
		w.Write([]byte{'\n'})
	}

	return len(failed) + len(success), len(success)
}

func TestCSmith(t *testing.T) {
	gcc := os.Getenv("CC")
	if gcc == "" {
		gcc = "gcc"
	}
	gcc, err := exec.LookPath(gcc)
	if err != nil {
		t.Skip(err)
		return
	}

	if testing.Short() {
		t.Skip("skipped: -short")
	}

	csmith, err := exec.LookPath("csmith")
	if err != nil {
		t.Skip(err)
		return
	}
	binaryName := filepath.FromSlash("./a.out")
	mainName := filepath.FromSlash("main.go")
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	defer os.Chdir(wd)

	temp, err := ioutil.TempDir("", "ccgo-test-")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(temp)

	if err := os.Chdir(temp); err != nil {
		t.Fatal(err)
	}

	if os.Getenv("GO111MODULE") != "off" {
		if out, err := Shell("go", "mod", "init", "example.com/ccgo/v3/lib/csmith"); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}

		if out, err := Shell("go", "get", "modernc.org/libc"); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}
	}

	fixedBugs := []string{
		"--bitfields --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid --max-nested-struct-level 10 -s 1906742816",
		"--bitfields --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid --max-nested-struct-level 10 -s 612971101",
		"--bitfields --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid --max-nested-struct-level 10 -s 3629008936",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 4130344133",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 3130410542",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 1833258637",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 3126091077",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 2205128324",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 3043990076",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 2517344771",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 56498550",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 3645367888",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 169375684",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 3578720023",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 1885311141",
		"--no-bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 3720922579",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 241244373",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 517639208",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 2205128324",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 2876930815",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 3365074920",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 3329111231",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 2648215054",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 3919255949",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 890611563",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 4101947480",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 4058772172",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 2273393378",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 3100949894",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 15739796933983044010", //TODO fails on linux/s390x

		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 963985971",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 3363122597",
		"--bitfields --max-nested-struct-level 10 --no-const-pointers --no-consts --no-packed-struct --no-volatile-pointers --no-volatiles --paranoid -s 4146870674",
	}
	ch := time.After(*oCSmith)
	t0 := time.Now()
	var files, ok int
	var size int64
	var re *regexp.Regexp
	if s := *oRE; s != "" {
		re = regexp.MustCompile(s)
	}
out:
	for i := 0; ; i++ {
		extra := ""
		var args string
		switch {
		case i < len(fixedBugs):
			if re != nil && !re.MatchString(fixedBugs[i]) {
				continue
			}

			args += fixedBugs[i]
			a := strings.Split(fixedBugs[i], " ")
			extra = strings.Join(a[len(a)-2:], " ")
			t.Log(args)
		default:
			select {
			case <-ch:
				break out
			default:
			}

			args += csmithDefaultArgs
		}
		csOut, err := exec.Command(csmith, strings.Split(args, " ")...).Output()
		if err != nil {
			t.Fatalf("%v\n%s", err, csOut)
		}

		if fn := *oBlackBox; fn != "" {
			if err := ioutil.WriteFile(fn, csOut, 0660); err != nil {
				t.Fatal(err)
			}
		}

		if err := ioutil.WriteFile("main.c", csOut, 0660); err != nil {
			t.Fatal(err)
		}

		csp := fmt.Sprintf("-I%s", filepath.FromSlash("/usr/include/csmith"))
		if s := os.Getenv("CSMITH_PATH"); s != "" {
			csp = fmt.Sprintf("-I%s", s)
		}

		ccOut, err := exec.Command(gcc, "-o", binaryName, "main.c", csp).CombinedOutput()
		if err != nil {
			t.Fatalf("%s\n%s\ncc: %v", extra, ccOut, err)
		}

		binOutA, err := func() ([]byte, error) {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			return exec.CommandContext(ctx, binaryName).CombinedOutput()
		}()
		if err != nil {
			continue
		}

		size += int64(len(csOut))

		if err := os.Remove(binaryName); err != nil {
			t.Fatal(err)
		}

		files++
		var stdout, stderr bytes.Buffer
		j := NewTask([]string{
			"ccgo",

			"-o", mainName,
			"-verify-structs",
			"main.c",
			csp,
		}, &stdout, &stderr)
		j.cfg.MaxSourceLine = 1 << 20

		func() {

			defer func() {
				if err := recover(); err != nil {
					t.Errorf("%s\n%s\nccgo: %s\n%s\n%s", extra, csOut, stdout.Bytes(), stderr.Bytes(), debug.Stack())
					t.Fatal(err)
				}
			}()

			if err := j.Main(); err != nil || stdout.Len() != 0 {
				t.Errorf("%s\n%s\nccgo: %s\n%s", extra, csOut, stdout.Bytes(), stderr.Bytes())
				t.Fatal(err)
			}
		}()

		binOutB, err := func() ([]byte, error) {
			ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
			defer cancel()

			return exec.CommandContext(ctx, "go", "run", "-tags=libc.memgrind", mainName).CombinedOutput()
		}()
		if err != nil {
			t.Errorf("%s\n%s\n%s\nccgo: %v", extra, csOut, binOutB, err)
			break
		}

		if g, e := binOutB, binOutA; !bytes.Equal(g, e) {
			t.Errorf("%s\n%s\nccgo: %v\ngot: %s\nexp: %s", extra, csOut, err, g, e)
			break
		}

		ok++
		if *oTrace {
			fmt.Fprintln(os.Stderr, time.Since(t0), files, ok)
		}

		if err := os.Remove(mainName); err != nil {
			t.Fatal(err)
		}
	}
	d := time.Since(t0)
	t.Logf("files %v, bytes %v, ok %v in %v", h(files), h(size), h(ok), d)
}

func dumpInitializer(s []*cc.Initializer) string {
	if len(s) == 0 {
		return "<empty>"
	}
	var a []string
	for _, v := range s {
		var s string
		if f := v.Field; f != nil {
			s = fmt.Sprintf("fld %q bitfield %v bitoff %2d", f.Name(), f.IsBitField(), f.BitFieldOffset())
		}
		a = append(a, fmt.Sprintf("%v: off %#04x val %v %s", v.Position(), v.Offset, v.AssignmentExpression.Operand.Value(), s))
	}
	return strings.Join(a, "\n")
}
