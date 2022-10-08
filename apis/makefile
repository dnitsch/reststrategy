
codegen:
	./hack/update-codegen.sh

copygened:
	echo "Copying Generated dir" && cp -fr github.com/dnitsch/reststrategy/apis/generated ./ && \
	echo "Copying Generated DeepCopyFunc" && cp -fr github.com/dnitsch/reststrategy/apis/reststrategy/* ./reststrategy/ && \
	echo "rm -rf github.com" && rm -rf github.com

build: codegen copygened
	echo "Types Updated"