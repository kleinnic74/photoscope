TMPDIR=tmp
BINDIR=bin
BINDIR_WIN=$(BINDIR)/win
BINDIR_ARM=$(BINDIR)/arm

BINARY_UNIX=$(BINDIR)/photos
BINARY_OSX=$(BINDIR)/osx/photos
BINARY_WIN=$(BINDIR_WIN)/photos.exe
BINARY_ARM=$(BINDIR_ARM)/photos

TOOLS=./cmd/dbinspect ./cmd/dircheck ./cmd/exifprint

PKG=./cmd/photos

BINARIES=$(BINARY_WIN) $(BINARY_ARM) $(BINARY_UNIX) $(TOOLS)
ifeq ($(shell uname -s),Darwin)
	BINARY_MAIN=${BINARY_OSX}
	BINARIES=$(BINARY_OSX) $(BINARIES)
else ifeq ($(shell uname -s),Linux)
	BINARY_MAIN=${BINARY_UNIX}
else
	BINARY_MAIN=${BINARY_WIN}
endif

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

.PHONY: $(BINARY_WIN) 
$(BINARY_WIN): $(BINDIR) generate
	$(GO_WIN) go build -ldflags "$(GO_VARS)" -o $(BINARY_WIN) $(PKG)

.PHONY: $(BINARY_ARM) 
$(BINARY_ARM): $(BINDIR) generate
	$(GO_ARM) go build -ldflags "$(GO_VARS)" -o $(BINARY_ARM) $(PKG)

.PHONY: $(BINARY_UNIX) 
$(BINARY_UNIX): $(BINDIR) generate
	$(GO_UX) go build -ldflags "$(GO_VARS)" -o $(BINARY_UNIX) $(PKG)

.PHONY: $(BINARY_OSX) 
$(BINARY_OSX): $(BINDIR) generate
	$(GO_OSX) go build -ldflags "$(GO_VARS)" -o $(BINARY_OSX) $(PKG)


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
	$(RM) -r $(BINDIR)
	$(RM) embed/embedded_resources.go
	go clean ./...

.PHONY: run
run: GO_DEBUG_VAR=-X 'bitbucket.org/kleinnic74/photos/consts.devmode=false'
run: _run

.PHONY: _run
_run: $(BINARY_WIN) $(TMPDIR)
	cd $(TMPDIR) && ../$(BINARY_WIN) -ui ../frontend/build

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

frontend/build: $(wildcard frontend/src/**/*) $(wildcard frontend/public/**/*) frontend/node_modules
	cd frontend && npm run build
	touch frontend/build

.PHONY: runui
runui:
	cd frontend && npm start

.PHONY: deps

deps: deptree.svg

deptree.svg:
	godepgraph -s ./cmd/photos | dot -Tsvg >deptree.svg
