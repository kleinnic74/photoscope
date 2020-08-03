.PHONY: all build test run

TMPDIR=tmp
BINDIR=bin

BINARY_WIN=$(BINDIR)/win/photos.exe
BINARY_ARM=$(BINDIR)/arm/photos

PKG=./cmd/photos

BINARIES=$(BINARY_WIN) $(BINARY_ARM)

GO_VARS=

all: build

$(BINDIR):
	mkdir $(BINDIR)

$(TMPDIR):
	mkdir $(TMPDIR)

.PHONY: $(BINARY_WIN)

$(BINARY_WIN): $(BINDIR)
	go build -ldflags "$(GO_VARS)" -o $(BINARY_WIN) $(PKG)

$(BINARY_ARM): $(BINDIR)
	CGO_ENABLED=0 GOARM=7 GOARCH=arm GOOS=linux go build $(GO_VARS) -o $(BINARY_ARM) $(PKG)

build: $(BINARIES)

test:
	go test -v ./...

clean:
	rm -fr $(BINDIR)
	rm -fr $(TMPDIR)
	go clean ./...

run: $(BINARY_WIN) $(TMPDIR)
	cd $(TMPDIR) && ../$(BINARY_WIN) import

.PHONY: rundev
rundev: run

rundev: GO_VARS=-X 'bitbucket.org/kleinnic74/photos/consts.devmode=true'
