
OWNER := dnitsch
NAME := reststrategy
GIT_TAG := "0.6.2"
VERSION := "v$(GIT_TAG)"
REVISION := $(shell git rev-parse --short HEAD)

build: 
	echo "build seeder first as it contains nested types for APIs" && cd seeder && make OWNER=$(OWNER) NAME=$(NAME) VERSION=$(VERSION) REVISION=$(REVISION) build
	echo "build apis" && cd apis && make OWNER=$(OWNER) NAME=$(NAME) VERSION=$(VERSION) REVISION=$(REVISION) build
	echo "build controller" && cd controller && make OWNER=$(OWNER) NAME=$(NAME) VERSION=$(VERSION) REVISION=$(REVISION) build

build_ci: 
	echo "build seeder first as it contains nested types for APIs" && cd seeder && make OWNER=$(OWNER) NAME=$(NAME) VERSION=$(VERSION) REVISION=$(REVISION) build_ci
	echo "build apis" && cd apis && make OWNER=$(OWNER) NAME=$(NAME) VERSION=$(VERSION) REVISION=$(REVISION) build_ci
	echo "build controller" && cd controller && make OWNER=$(OWNER) NAME=$(NAME) VERSION=$(VERSION) REVISION=$(REVISION) build_ci


tag: 
	git tag "v$(GIT_TAG)"
	git tag "apis/v$(GIT_TAG)"
	git tag "seeder/v$(GIT_TAG)"
	git tag "controller/v$(GIT_TAG)"
	git push --tags

release: 
	OWNER=$(OWNER) NAME=$(NAME) PAT=$(PAT) VERSION=$(VERSION) . hack/release.sh

# for local development install all dependencies 
# in workspace
install: 
	go work sync
