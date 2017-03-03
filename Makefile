all: build test

build:
	go install github.com/parkr/radar/...

test:
	go test github.com/parkr/radar/...
	go vet github.com/parkr/radar/...

docker-release: all
	docker build -t parkr/radar:$(shell git rev-parse HEAD) .
	docker push parkr/radar:$(shell git rev-parse HEAD)
