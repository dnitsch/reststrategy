.PHONY: clean install codegen codegen_verify copygened bingen build

install: 
	go mod tidy
	go mod vendor

codegen:
	./hack/update-codegen.sh

copygened: codegen
	echo "Copying Generated dir and DeepCopyFunc" && cp -fr github.com/dnitsch/reststrategy/apis/reststrategy/* ./reststrategy/ && \
	echo "rm -rf github.com" && rm -rf github.com

codegen_verify:
	./hack/verify-codegen.sh
	echo "rm -rf github.com" && rm -rf github.com

clean:
	rm -rf vendor/*

bingen:
	for os in darwin linux windows; do \
	    [ $$os = "windows" ] && EXT=".exe"; \
		GOOS=$$os CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o dist/seeder-$$os$$EXT ./cmd; \
	done

build: clean install copygened bingen
	echo "Types Updated"

build_ci: clean install codegen_verify bingen 
	echo "Types Updated and verified"
