build/diffgpt: $(shell find . -name '*.go')
	CGO_ENABLED=0 go build -ldflags="-s -w" -o build/diffgpt .

build/diffgpt-linux-amd64: $(shell find . -name '*.go')
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build/diffgpt-linux-amd64 .

build: build/diffgpt

install:
	go install

deps:
	go mod tidy

clean:
	rm -f build/diffgpt
	rm -rf diffgpt-linux-amd64
	rm -f diffgpt-linux-amd64.tar.gz

run: build
	./build/diffgpt

release: build/diffgpt-linux-amd64
	cp build/diffgpt-linux-amd64 diffgpt
	tar czf diffgpt-linux-amd64.tar.gz -C build diffgpt
	rm -rf diffgpt
