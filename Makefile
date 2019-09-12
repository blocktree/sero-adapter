.PHONY: all clean
.PHONY: openw-sero
.PHONY: deps

# Check for required command tools to build or stop immediately
EXECUTABLES = git go find pwd
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH)))

GO ?= latest

# openw-sero
OPENWSEROVERSION = $(shell git describe --tags `git rev-list --tags --max-count=1`)
OPENWSEROBINARY = openw-sero
OPENWSEROMAIN = serocli/main.go

BUILDDIR = build
GITREV = $(shell git rev-parse --short HEAD)
BUILDTIME = $(shell date +'%Y-%m-%d_%T')

OPENWSEROLDFLAGS="-X github.com/blocktree/sero-adapter/serocli/commands.Version=${OPENWSEROVERSION} \
	-X github.com/blocktree/sero-adapter/serocli/commands.GitRev=${GITREV} \
	-X github.com/blocktree/sero-adapter/serocli/commands.BuildTime=${BUILDTIME} \
	-X github.com/blocktree/go-openw-cli/openwcli.FixAppID=${APPID} \
	-X github.com/blocktree/go-openw-cli/openwcli.FixAppKey=${APPKEY}"

# OS platfom
# options: windows-6.0/*,darwin-10.10/amd64,linux/amd64,linux/386,linux/arm64,linux/mips64, linux/mips64le
TARGETS="darwin-10.10/amd64,linux/amd64,windows-6.0/*"

deps:
	go get -u github.com/gythialy/xgo

build:
	GO111MODULE=on go build -ldflags $(OPENWSEROLDFLAGS) -i -o $(shell pwd)/$(BUILDDIR)/$(OPENWSEROBINARY) $(shell pwd)/$(OPENWSEROMAIN)
	@echo "Build $(OPENWSEROBINARY) done."

all: openw-sero

clean:
	rm -rf $(shell pwd)/$(BUILDDIR)/

openw-sero:
	xgo --dest=$(BUILDDIR) --ldflags=$(OPENWSEROLDFLAGS) --out=$(OPENWSEROBINARY)-$(OPENWSEROVERSION)-$(GITREV) --targets=$(TARGETS) \
	--pkg=$(OPENWSEROMAIN) .
