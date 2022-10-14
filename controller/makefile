OWNER := dnitsch
VERSION := "v0.0.0"
REVISION := "aaaa1111"

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

bingen:
	for os in darwin linux windows; do \
	    [ $$os = "windows" ] && EXT=".exe"; \
		GOOS=$$os CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o dist/seeder-$$os$$EXT ./cmd; \
	done

# local only as the target is the docker build
build: clean install bingen

build_ci: build

docker:
	docker build --build-arg REVISION=$(REVISION) --build-arg VERSION=$(VERSION) -t ghcr.io/dnitsch/reststrategy:$(VERSION) .
