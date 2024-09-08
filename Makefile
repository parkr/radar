DOCKER_TAG:=$(shell git rev-parse HEAD)
DOCKER_IMAGE:=parkr/radar:$(DOCKER_TAG)
PLATFORMS:=linux/amd64,linux/arm64

all: build test

build:
	go install github.com/parkr/radar/...

test:
	go test github.com/parkr/radar/...
	go vet github.com/parkr/radar/...

server: build
	$(shell radar -http="localhost:8291" -debug)

docker-build:
	docker buildx build \
        --platform $(PLATFORMS) \
        --load \
        -t $(DOCKER_IMAGE) \
		.

docker-test: docker-build
	docker run --rm \
	  --name radar \
	  -e RADAR_HEALTHCHECK_URL=http://0.0.0.0:8291/health \
	  -e RADAR_REPO=parkr/radar \
	  $(DOCKER_IMAGE) \
	  radar -http=":8291" -debug


docker-release: docker-buildx-info
	docker buildx build \
        --platform $(PLATFORMS) \
        --push \
        -t $(DOCKER_IMAGE) \
		.

docker-buildx-info:
	docker buildx version
	docker buildx ls

docker-buildx-create: docker-buildx-info
	docker buildx create --use --bootstrap --platform $(PLATFORMS)
