DOCKER_TAG:=$(shell git rev-parse HEAD)
DOCKER_IMAGE:=parkr/radar:$(DOCKER_TAG)
PLATFORMS:=linux/amd64,linux/arm64

.PHONY: all build test server docker-build docker-test docker-release docker-buildx-info docker-buildx-create
all: build test
bin/%:
	$(eval BINARY_NAME := $(patsubst bin/%,%,$@))
	go build -o bin/$(BINARY_NAME) github.com/parkr/radar/cmd/$(BINARY_NAME)

build: bin/radar bin/radar-healthcheck bin/radar-poster
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
