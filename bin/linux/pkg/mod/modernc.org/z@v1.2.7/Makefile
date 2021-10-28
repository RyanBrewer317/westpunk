# Copyright 2021 The Zlib-Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

.PHONY:	all bench clean cover cpu editor internalError later mem nuke todo edit devbench

grep=--include=*.go
ngrep='TODOOK\|internalError\|testdata\|scanner.go\|parser.go\|ast.go\|assets.go\|TODO-'


all:
	@LC_ALL=C date
	@go version 2>&1 | tee log
	@gofmt -l -s -w *.go
	@go install -v ./...
	@go test -i
	@go test 2>&1 -timeout 1h | tee -a log
	@go vet 2>&1 | grep -v $(ngrep) || true
	@golint 2>&1 | grep -v $(ngrep) || true
	@make todo
	@misspell *.go
	@staticcheck | grep -v 'scanner\.go' || true
	@maligned || true
	@grep -n --color=always 'FAIL\|PASS' log 
	LC_ALL=C date 2>&1 | tee -a log

# generate on current host
generate:
	go generate 2>&1 | tee log-generate
	gofmt -l -s -w . 2>&1 | tee -a log-generate

build_all_targets:
	GOOS=darwin GOARCH=amd64 go build -v
	GOOS=darwin GOARCH=arm64 go build -v
	GOOS=freebsd GOARCH=arm64 go build -v
	GOOS=linux GOARCH=386 go build -v
	GOOS=linux GOARCH=amd64 go build -v
	GOOS=linux GOARCH=arm go build -v
	GOOS=linux GOARCH=arm64 go build -v
	GOOS=linux GOARCH=s390x go build -v
	GOOS=netbsd GOARCH=arm64 go build -v
	GOOS=windows GOARCH=386 go build -v
	GOOS=windows GOARCH=amd64 go build -v
	echo done

# only on darwin host
darwin_amd64:
	make generate
	go test -v

# only on s390x
linux_s390x:
	make generate
	go test -v

# only on darwin host
darwin_arm64:
	make generate
	go test -v

# only on freebsd host
freebsd_amd64:
	AR=/usr/bin/ar make generate
	go test -v

# only on netbsd host
netbsd_amd64:
	AR=/usr/bin/ar make generate
	go test -v

# on linux/amd64
linux_amd64:
	TARGET_GOOS=linux TARGET_GOARCH=amd64 make generate
	GOOS=linux GOARCH=amd64 go test -v

# on linux/amd64
linux_386:
	CCGO_CPP=i686-linux-gnu-cpp GO_GENERATE_CC=i686-linux-gnu-gcc TARGET_GOOS=linux TARGET_GOARCH=386 make generate
	GOOS=linux GOARCH=386 go test -v

# on linux/amd64
linux_arm:
	CCGO_CPP=arm-linux-gnueabi-cpp-8 GO_GENERATE_CC=arm-linux-gnueabi-gcc-8 TARGET_GOOS=linux TARGET_GOARCH=arm make generate
	GOOS=linux GOARCH=arm go test -v

# on linux/amd64
linux_arm64:
	CCGO_CPP=aarch64-linux-gnu-cpp-8 GO_GENERATE_CC=aarch64-linux-gnu-gcc-8 TARGET_GOOS=linux TARGET_GOARCH=arm64 make generate
	GOOS=linux GOARCH=arm64 go test -v

# only on windows host with mingw gcc installed
windows_amd64:
	make generate
	go test -v

# only on windows host with mingw gcc installed
windows_386:
	make generate
	go test -v

devbench:
	date 2>&1 | tee log-devbench
	go test -timeout 24h -dev -run @ -bench . 2>&1 | tee -a log-devbench
	grep -n 'FAIL\|SKIP' log-devbench || true

bench:
	date 2>&1 | tee log-bench
	go test -timeout 24h -v -run '^[^E]' -bench . 2>&1 | tee -a log-bench
	grep -n 'FAIL\|SKIP' log-bench || true

clean:
	go clean
	rm -f *~ *.test *.out

cover:
	t=$(shell mktemp) ; go test -coverprofile $$t && go tool cover -html $$t && unlink $$t

cpu: clean
	go test -run @ -bench . -cpuprofile cpu.out
	go tool pprof -lines *.test cpu.out

edit:
	@touch log
	@if [ -f "Session.vim" ]; then gvim -S & else gvim -p Makefile *.go & fi

editor:
	gofmt -l -s -w *.go
	nilness .
	GO111MODULE=off go install -v 2>&1 | tee log-install

later:
	@grep -n $(grep) LATER * || true
	@grep -n $(grep) MAYBE * || true

mem: clean
	go test -run @ -dev -bench . -memprofile mem.out -timeout 24h
	go tool pprof -lines -web -alloc_space *.test mem.out

nuke: clean
	go clean -i

todo:
	@grep -nr $(grep) ^[[:space:]]*_[[:space:]]*=[[:space:]][[:alpha:]][[:alnum:]]* * | grep -v $(ngrep) || true
	@grep -nrw $(grep) 'TODO\|panic' * | grep -v $(ngrep) || true
	@grep -nr $(grep) BUG * | grep -v $(ngrep) || true
	@grep -nr $(grep) [^[:alpha:]]println * | grep -v $(ngrep) || true
	@grep -nir $(grep) 'work.*progress' || true
