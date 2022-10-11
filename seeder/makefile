install:
	go mod tidy
	go mod vendor

codegen:
	./hack/update-codegen.sh

copygened:
	echo "Copying Generated DeepCopyFunc" && cp -fr github.com/dnitsch/reststrategy/seeder/pkg/rest/* ./pkg/rest/ && \
	echo "rm -rf github.com" && rm -rf github.com

build: codegen copygened
	echo "Types Updated"