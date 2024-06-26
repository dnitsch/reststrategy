OWNER := dnitsch
NAME := reststrategy
GIT_TAG := "0.9.0"
VERSION := "v$(GIT_TAG)"
REVISION := "aaaa1111-always-overwrite-in-CI"

build_seeder: 
	echo "build seeder first as it contains nested types for APIs" 
	cd seeder && make OWNER=$(OWNER) NAME=$(NAME) VERSION=$(VERSION) REVISION=$(REVISION) build

build_apis: 
	echo "build apis" 
	cd apis && make OWNER=$(OWNER) NAME=$(NAME) VERSION=$(VERSION) REVISION=$(REVISION) build

build_controller: 
	echo "build controller"
	cd controller && make OWNER=$(OWNER) NAME=$(NAME) VERSION=$(VERSION) REVISION=$(REVISION) build

build: build_seeder # build_apis build_controller

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

docker_build_raw_controller:
	cd controller && \
	docker build --build-arg REVISION=$(REVISION) --build-arg VERSION=$(VERSION) -t ghcr.io/dnitsch/reststrategy:raw-$(VERSION) .

docker_build_kubebuilder_controller:
	cd kubebuilder-controller && \
	docker build --build-arg REVISION=$(REVISION) --build-arg VERSION=$(VERSION) -t ghcr.io/dnitsch/reststrategy:kubebuilder-$(VERSION) .

docker_release: docker_build_kubebuilder_controller docker_build_raw_controller
	docker push ghcr.io/dnitsch/reststrategy:$(VERSION)

# for local development install all dependencies 
# in workspace
install: 
	go work sync

.PHONY: test test_seeder test_controller
test_seeder:
	go test ./seeder/... -v -mod=readonly -race -coverprofile=seeder/.coverage/out | go-junit-report > seeder/.coverage/report-junit.xml && \
	gocov convert seeder/.coverage/out | gocov-xml > seeder/.coverage/report-cobertura.xml

test_controller:
	go test ./controller/... -v -mod=readonly -race -timeout 10m0s -coverprofile=controller/.coverage/out | go-junit-report > controller/.coverage/report-junit.xml && \
	gocov convert controller/.coverage/out | gocov-xml > controller/.coverage/report-cobertura.xml

test_kubebuilder_controller:
	go test ./kubebuilder-controller/... -v -mod=readonly -timeout 10m0s -race -coverprofile=kubebuilder-controller/.coverage/out | go-junit-report > kubebuilder-controller/.coverage/report-junit.xml && \
	gocov convert kubebuilder-controller/.coverage/out | gocov-xml > kubebuilder-controller/.coverage/report-cobertura.xml

# running in CI on an alpine container without gcc so only running -race on local set up
test: test_prereq test_seeder test_controller test_kubebuilder_controller

test_prereq: 
	mkdir -p seeder/.coverage controller/.coverage kubebuilder-controller/.coverage
	go install github.com/jstemmer/go-junit-report/v2@latest && \
	go install github.com/axw/gocov/gocov@latest && \
	go install github.com/AlekSi/gocov-xml@latest

coverage: test
	go tool cover -html=seeder/.coverage/out
	go tool cover -html=controller/.coverage/out

coverage_seeder: test_seeder
	go tool cover -html=seeder/.coverage/out

coverage_controller: test_controller
	go tool cover -html=controller/.coverage/out

coverage_kubebuilder_controller: test_kubebuilder_controller
	go tool cover -html=kubebuilder-controller/.coverage/out