NAME := reststrategyseeder
VERSION := "v0.0.0"
REVISION := $(shell git rev-parse --short HEAD)

LDFLAGS := -ldflags="-s -w -X \"github.com/$(OWNER)/$(NAME)/seeder/cmd/strategyrestseeder.Version=$(VERSION)\" -X \"github.com/$(OWNER)/$(NAME)/seeder/cmd/strategyrestseeder.Revision=$(REVISION)\" -extldflags -static"

install:
	go mod tidy
	go mod vendor
	
codegen:
	./hack/update-codegen.sh

copygened: codegen
	echo "Copying Generated DeepCopyFunc" && cp -fr github.com/dnitsch/reststrategy/seeder/pkg/rest/* ./pkg/rest/ && \
	echo "rm -rf github.com" && rm -rf github.com
	echo "Types Updated"

.PHONY: clean
clean:
	rm -rf bin/*
	rm -rf dist/*
	rm -rf vendor/*
	mkdir -p dist

build: clean install copygened
	for os in darwin linux windows; do \
	    [ $$os = "windows" ] && EXT=".exe"; \
		GOOS=$$os CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o dist/$(NAME)-$$os$$EXT ./cmd; \
	done
