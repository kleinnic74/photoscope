TMPDIR=tmp
BINDIR=bin
APPNAME=photoscope

BINDIR_WIN=$(BINDIR)/win
BINDIR_ARM=$(BINDIR)/arm

PLATFORMS=darwin win linux arm

BINARY_linux=$(BINDIR)/linux/$(APPNAME)
BINARY_darwin=$(BINDIR)/darwin/$(APPNAME)
BINARY_win=$(BINDIR)/win/$(APPNAME).exe
BINARY_arm=$(BINDIR)/arm/$(APPNAME)

TOOLS=./cmd/dbinspect ./cmd/dircheck ./cmd/exifprint

PKG=./cmd/photos

OS=$(shell uname -s | tr '[:upper:]' '[:lower:]')
BINARIES=$(BINARY_win) $(BINARY_arm) $(BINARY_linux) $(BINARY_darwin) $(TOOLS)
BINARY_MAIN=$(BINARY_$(OS))

FRONTEND=frontend/

GIT_COMMIT=$(shell git rev-list -1 HEAD)
GIT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

GO_DEBUG_VAR=-X 'bitbucket.org/kleinnic74/photos/consts.devmode=true'
GO_VARS=$(GO_DEBUG_VAR) -X 'bitbucket.org/kleinnic74/photos/logging.errorLog=true' \
	-X 'bitbucket.org/kleinnic74/photos/consts.GitCommit=$(GIT_COMMIT)' 
GO_ARM=CGO_ENABLED=0 GOARM=7 GOARCH=arm GOOS=linux
GO_WIN=CGO_ENABLED=0 GOOS=windows
GO_UX=CGO_ENABLED=0 GOOS=linux
GO_OSX=CGO_ENABLED=0 GOOS=darwin

.PHONY: all
all: build frontend/build

$(BINDIR):
	mkdir $(BINDIR)

$(TMPDIR):
	mkdir $(TMPDIR)

$(BINARY_win): $(BINDIR) generate
	$(GO_WIN) go build -ldflags "$(GO_VARS)" -o $@ $(PKG)

$(BINARY_arm): $(BINDIR) generate
	$(GO_ARM) go build -ldflags "$(GO_VARS)" -o $@ $(PKG)

$(BINARY_linux): $(BINDIR) generate
	$(GO_UX) go build -ldflags "$(GO_VARS)" -o $@ $(PKG)

$(BINARY_darwin): $(BINDIR) generate
	$(GO_OSX) go build -ldflags "$(GO_VARS)" -o $@ $(PKG)


.PHONY: tools
tools:
	go build -o $(BINDIR_WIN) $(TOOLS)
	$(GO_ARM) go build -o $(BINDIR_ARM) $(TOOLS)

.PHONY: tools_arm
tools_arm:
	$(GO_ARM) go build -o $(BINDIR_ARM) $(TOOLS)

.PHONY: build
build: $(BINARIES)

.PHONY: test
test:
	go test -v ./...

.PHONY: clean
clean:
	$(RM) -r $(BINDIR) dist
	$(RM) embed/embedded_resources.go
	go clean ./...

.PHONY: run
run: GO_DEBUG_VAR=-X 'bitbucket.org/kleinnic74/photos/consts.devmode=false'
run: _run

.PHONY: _run
_run: $(BINARY_MAIN) $(TMPDIR)
	cd $(TMPDIR) && ../$(BINARY_MAIN) -ui ../frontend/build

.PHONY: rundev
rundev: GO_DEBUG_VAR=-X 'bitbucket.org/kleinnic74/photos/consts.devmode=true'
rundev: _run


.PHONY: generate
generate: embed/embedded_resources.go

embed/embedded_resources.go: frontend/build
	rm -f embed/embedded_resources.go && go generate ./embed

frontend/node_modules: frontend/package.json
	cd frontend && npm install
	touch frontend/node_modules

frontend/build: $(shell find frontend/src -type f) $(shell find frontend/public -type f) frontend/node_modules
	cd frontend && npm run build
	touch frontend/build

.PHONY: runui
runui:
	cd frontend && npm start

.PHONY: deps

deps: deptree.svg

deptree.svg:
	godepgraph -s ./cmd/photos | dot -Tsvg >deptree.svg

.PHONY: dist
dist: $(BINARIES)
	@mkdir -p dist
	@echo os=$(OS) binary_main=$(BINARY_MAIN)
	for p in $(PLATFORMS); do \
	    echo Building dist for $$p... ; \
		files=$$(cd $(BINDIR)/$$p && find . -type f) ; \
		tar cvfz dist/$(APPNAME)-$$p.tar.gz -C $(BINDIR)/$$p $$files ; \
	done
