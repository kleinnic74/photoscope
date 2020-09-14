TMPDIR=tmp
BINDIR=bin

BINARY_WIN=$(BINDIR)/win/photos.exe
BINARY_ARM=$(BINDIR)/arm/photos

TOOLS=./cmd/dbinspect

PKG=./cmd/photos

BINARIES=$(BINARY_WIN) $(BINARY_ARM) $(TOOLS)

GO_VARS=

.PHONY: all
all: build

$(BINDIR):
	mkdir $(BINDIR)

$(TMPDIR):
	mkdir $(TMPDIR)

.PHONY: $(BINARY_WIN) 
$(BINARY_WIN): $(BINDIR) generate
	go build -ldflags "$(GO_VARS)" -o $(BINARY_WIN) $(PKG)

.PHONY: $(BINARY_ARM) 
$(BINARY_ARM): $(BINDIR) generate
	CGO_ENABLED=0 GOARM=7 GOARCH=arm GOOS=linux go build $(GO_VARS) -o $(BINARY_ARM) $(PKG)

.PHONY: tools
tools:
	go build -o $(BINDIR) $(TOOLS)

.PHONY: build
build: $(BINARIES)

.PHONY: test
test:
	go test -v ./...

.PHONY: clean
clean:
	rm -fr $(BINDIR)
	go clean ./...

.PHONY: run
run: $(BINARY_WIN) $(TMPDIR)
	cd $(TMPDIR) && ../$(BINARY_WIN) -ui ../frontend/build

.PHONY: rundev
rundev: run

rundev: GO_VARS=-X 'bitbucket.org/kleinnic74/photos/consts.devmode=true'

.PHONY: generate
generate:
	go generate ./embed
