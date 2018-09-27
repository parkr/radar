DOCKER_TAG:=$(shell git rev-parse HEAD)
DOCKER_IMAGE:=parkr/radar:$(DOCKER_TAG)

all: build test

build:
	go install github.com/parkr/radar/...

test:
	go test github.com/parkr/radar/...
	go vet github.com/parkr/radar/...

docker-build: all
	docker build -t $(DOCKER_IMAGE) .

docker-test: all docker-build
	docker run --rm $(DOCKER_IMAGE)

docker-release: all docker-build
	docker push $(DOCKER_IMAGE)
