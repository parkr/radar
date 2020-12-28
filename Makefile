DOCKER_TAG:=$(shell git rev-parse HEAD)
DOCKER_IMAGE:=parkr/radar:$(DOCKER_TAG)

all: build test

build:
	go install github.com/parkr/radar/...

test:
	go test github.com/parkr/radar/...
	go vet github.com/parkr/radar/...

server: build
	$(shell radar -http="localhost:8291" -debug)

docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-test: docker-build
	docker run --rm \
	  --name radar \
	  -e RADAR_HEALTHCHECK_URL=http://0.0.0.0:8291/health \
	  $(DOCKER_IMAGE) \
	  radar -http=":8291" -debug


docker-release: docker-build
	docker push $(DOCKER_IMAGE)
