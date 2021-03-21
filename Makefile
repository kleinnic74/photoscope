TMPDIR=tmp
BINDIR=bin
BINDIR_WIN=$(BINDIR)/win
BINDIR_ARM=$(BINDIR)/arm

BINARY_WIN=$(BINDIR_WIN)/photos.exe
BINARY_ARM=$(BINDIR_ARM)/photos

TOOLS=./cmd/dbinspect ./cmd/dircheck ./cmd/exifprint

PKG=./cmd/photos

BINARIES=$(BINARY_WIN) $(BINARY_ARM) $(TOOLS)

FRONTEND=frontend/

GIT_COMMIT=$(shell git rev-list -1 HEAD)

GO_DEBUG_VAR=-X 'bitbucket.org/kleinnic74/photos/consts.devmode=true'
GO_VARS=$(GO_DEBUG_VAR) -X 'bitbucket.org/kleinnic74/photos/logging.errorLog=true' \
	-X 'bitbucket.org/kleinnic74/photos/consts.GitCommit=$(GIT_COMMIT)' 
GO_ARM=CGO_ENABLED=0 GOARM=7 GOARCH=arm GOOS=linux

.PHONY: all
all: build frontend/build

$(BINDIR):
	mkdir $(BINDIR)

$(TMPDIR):
	mkdir $(TMPDIR)

.PHONY: $(BINARY_WIN) 
$(BINARY_WIN): $(BINDIR) generate
	go build -ldflags "$(GO_VARS)" -o $(BINARY_WIN) $(PKG)

.PHONY: $(BINARY_ARM) 
$(BINARY_ARM): $(BINDIR) generate
	$(GO_ARM) go build -ldflags "$(GO_VARS)" -o $(BINARY_ARM) $(PKG)

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

frontend/build: $(wildcard frontend/src/**/*) $(wildcard frontend/public/**/*)
	cd frontend && npm run build
	touch frontend/build

.PHONY: runui
runui:
	cd frontend && npm start

.PHONY: deps

deps: deptree.svg

deptree.svg:
	godepgraph -s ./cmd/photos | dot -Tsvg >deptree.svg
