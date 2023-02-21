NAME := reststrategy
VERSION := "v0.0.0"
REVISION := $(shell git rev-parse --short HEAD)

LDFLAGS := -ldflags="-s -w -X \"github.com/$(OWNER)/$(NAME)/seeder/cmd/strategyrestseeder.Version=$(VERSION)\" -X \"github.com/$(OWNER)/$(NAME)/seeder/cmd/strategyrestseeder.Revision=$(REVISION)\" -extldflags -static"

install:
	go mod tidy
	go mod vendor
	
install_ci:
	go mod vendor

codegen:
	./hack/update-codegen.sh

copygened: codegen
	echo "Copying Generated DeepCopyFunc" && cp -fr github.com/dnitsch/reststrategy/seeder/* ./ && \
	echo "rm -rf github.com" && rm -rf github.com
	echo "Types Updated"

codegen_verify:
	./hack/verify-codegen.sh
	echo "rm -rf github.com" && rm -rf github.com

.PHONY: clean
clean:
	rm -rf bin/*
	rm -rf dist/*
	rm -rf vendor/*
	mkdir -p dist

bingen:
	for os in darwin linux windows; do \
		GOOS=$$os CGO_ENABLED=0 go build -mod=readonly -buildvcs=false $(LDFLAGS) -o dist/seeder-$$os ./cmd; \
	done

build: clean install copygened bingen

build_ci: clean install_ci codegen_verify bingen
