all: build test

build:
	go install github.com/parkr/radar/...

test:
	go test github.com/parkr/radar/...
