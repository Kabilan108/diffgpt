build/diffgpt: $(shell find . -name '*.go')
	GO111MODULE=on go build -o build/diffgpt .

build: build/diffgpt

install:
	GO111MODULE=on go install

deps:
	GO111MODULE=on go mod tidy

clean:
	rm -f build/diffgpt

run: build
	./build/diffgpt