GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=spaceman

define SUMMARY
SpaceMan build. Available commands:

build - Build spaceman
test  - Test code

endef

export SUMMARY

all:
	@echo "$$SUMMARY"

build:
	@echo "Building binary..."
	$(GOBUILD) -o bin/spaceman spaceman.go
	strip --strip-unneeded bin/spaceman

test:
	@echo "Nothing here yet"

