OWNER := dnitsch
NAME := reststrategy
GIT_TAG := "0.6.7-alpha2"
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
	
# echo "build controller" && cd controller && make OWNER=$(OWNER) NAME=$(NAME) VERSION=$(VERSION) REVISION=$(REVISION) build_ci

tag: 
	git tag -a $(VERSION) -m "ci tag release reststrategy" $(REVISION)
	git tag -a apis/$(VERSION) -m "ci tag release reststrategy/apis" $(REVISION)
	git tag -a seeder/$(VERSION) -m "ci tag release reststrategy/seeder" $(REVISION)
	git push origin --tags

release: 
	OWNER=$(OWNER) NAME=$(NAME) PAT=$(PAT) VERSION=$(VERSION) . hack/release.sh

docker_release:
	cd controller && \
	docker build --build-arg REVISION=$(REVISION) --build-arg VERSION=$(VERSION) -t ghcr.io/dnitsch/reststrategy:$(VERSION) . && \
	docker push ghcr.io/dnitsch/reststrategy:$(VERSION)


# for local development install all dependencies 
# in workspace
install: 
	go work sync

.PHONY: test
test:
	cd seeder && go test -cover -v ./... 
	cd controller && go test -cover -v ./... 
