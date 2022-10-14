OWNER := dnitsch
NAME := reststrategy
GIT_TAG := "0.6.7-alpha"
VERSION := "v$(GIT_TAG)"
REVISION := $(shell git rev-parse --short HEAD)

build_seeder: 
	echo "build seeder first as it contains nested types for APIs" 
	cd seeder && make OWNER=$(OWNER) NAME=$(NAME) VERSION=$(VERSION) REVISION=$(REVISION) build

build_apis: 
	echo "build apis" 
	cd apis && make OWNER=$(OWNER) NAME=$(NAME) VERSION=$(VERSION) REVISION=$(REVISION) build

build_controller: 
	echo "build controller"
	cd controller && make OWNER=$(OWNER) NAME=$(NAME) VERSION=$(VERSION) REVISION=$(REVISION) build

buidl: build_seeder build_apis build_controller

build_ci: 
	echo "build seeder first as it contains nested types for APIs" && cd seeder && make OWNER=$(OWNER) NAME=$(NAME) VERSION=$(VERSION) REVISION=$(REVISION) build_ci
	echo "build apis" && cd apis && make OWNER=$(OWNER) NAME=$(NAME) VERSION=$(VERSION) REVISION=$(REVISION) build_ci
	echo "build controller" && cd controller && make OWNER=$(OWNER) NAME=$(NAME) VERSION=$(VERSION) REVISION=$(REVISION) build_ci

tag: 
	git tag "v$(GIT_TAG)"
	git tag "apis/v$(GIT_TAG)"
	git tag "seeder/v$(GIT_TAG)"
	git push --tags

release: 
	OWNER=$(OWNER) NAME=$(NAME) PAT=$(PAT) VERSION=$(VERSION) . hack/release.sh
	cd controller && make VERSION=$(VERSION) REVISION=$(REVISION) docker

# for local development install all dependencies 
# in workspace
install: 
	go work sync

.PHONY: test
test:
	cd seeder && go test -cover -v ./... 
	cd controller && go test -cover -v ./... 
