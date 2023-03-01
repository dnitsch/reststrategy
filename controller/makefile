OWNER := dnitsch
VERSION := "v0.0.0"
REVISION := "aaaa1111"

LDFLAGS := -ldflags="-s -w -X \"github.com/dnitsch/reststrategy/controller/cmd/controller.Version=$(VERSION)\" -X \"github.com/dnitsch/reststrategy/controller/cmd/controller.Revision=$(REVISION)\" -extldflags -static"

install:
	go mod tidy
	go mod vendor

install_ci:
	go mod vendor


.PHONY: clean
clean:
	rm -rf bin/*
	rm -rf dist/*
	rm -rf vendor/*
	mkdir -p dist

bingen:
	GOOS=$$os CGO_ENABLED=0 go build -mod=readonly $(LDFLAGS) -o dist/seeder-$$os ./cmd; 

# for os in darwin linux windows; do \
#     [ $$os = "windows" ] && EXT=".exe"; \
# done

# local only as the target is the docker build
build: clean install bingen

build_ci: clean install_ci bingen
