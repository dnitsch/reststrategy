install: 
	go mod tidy
	go mod vendor

codegen:
	./hack/update-codegen.sh

copygened:
	echo "Copying Generated dir" && cp -fr github.com/dnitsch/reststrategy/apis/generated ./
	echo "Copying Generated DeepCopyFunc" && cp -fr github.com/dnitsch/reststrategy/apis/reststrategy/* ./reststrategy/ && \
	echo "rm -rf github.com" && rm -rf github.com

.PHONY: clean
clean:
	rm -rf vendor/*

build: clean install codegen copygened 
	echo "Types Updated"
