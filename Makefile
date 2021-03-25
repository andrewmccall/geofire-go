all: 
	
clean: 
	rm -rf vendor/
	rm -rf build/

vendor: 
	go mod tidy
	go mod vendor

build: vendor
	go build

test: build
	go test
