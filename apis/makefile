.PHONY: clean install codegen codegen_verify copygened build

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

build: clean install copygened
	echo "Types Updated"

build_ci: clean install codegen_verify 
	echo "Types Updated and verified"
