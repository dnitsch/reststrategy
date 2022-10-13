OWNER := dnitsch
VERSION := "v0.0.0"
# VERSION := "$(shell git describe --tags --abbrev=0)"
REVISION := $(shell git rev-parse --short HEAD)

LDFLAGS := -ldflags="-s -w -X \"github.com/dnitsch/reststrategy/controller/cmd/controller.Version=$(VERSION)\" -X \"github.com/dnitsch/reststrategy/controller/cmd/controller.Revision=$(REVISION)\" -extldflags -static"


install:
	go mod tidy
	go mod vendor

.PHONY: clean
clean:
	rm -rf bin/*
	rm -rf dist/*
	rm -rf vendor/*
	mkdir -p dist

build: clean install
	GOOS=linux CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o dist/controller ./cmd; \

