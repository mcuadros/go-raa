# Package configuration
PROJECT = raa
CMD = raa

# Environment
BASE_PATH := $(shell pwd)
BUILD_PATH := $(BASE_PATH)/build
VERSION ?= $(shell git branch 2> /dev/null | sed -e '/^[^*]/d' -e 's/* \(.*\)/\1/')
BUILD ?= $(shell date)
ASSETS := static

# PACKAGES
PKG_OS = darwin linux
PKG_ARCH = amd64
PKG_CONTENT = ../../README.md ../../LICENSE

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOGET = $(GOCMD) get
GOTEST = $(GOCMD) test

# Rules
all: test build

build: clean
	$(GOCMD) build -ldflags "-X main.version $(VERSION) -X main.build \"$(BUILD)\"" .; \

test:
	$(GOTEST) -v ../.././...

install:
	cp -rf $(CMD) /usr/bin/; \

packages: clean
	for os in $(PKG_OS); do \
		for arch in $(PKG_ARCH); do \
			cd $(BASE_PATH); \
			mkdir -p $(BUILD_PATH)/$(PROJECT)_$(VERSION)_$${os}_$${arch}; \
			GOOS=$${os} GOARCH=$${arch} $(GOCMD) build -ldflags "-X main.version $(VERSION) -X main.build \"$(BUILD)\"" -o $(BUILD_PATH)/$(PROJECT)_$(VERSION)_$${os}_$${arch}/$(CMD) .; \
			for content in $(PKG_CONTENT); do \
				cp -rf $${content} $(BUILD_PATH)/$(PROJECT)_$(VERSION)_$${os}_$${arch}/; \
			done; \
			cd  $(BUILD_PATH) && tar -cvzf $(BUILD_PATH)/$(PROJECT)_$(VERSION)_$${os}_$${arch}.tar.gz $(PROJECT)_$(VERSION)_$${os}_$${arch}/; \
		done; \
	done;

clean:
	rm -rf $(BUILD_PATH)
	$(GOCLEAN) .
