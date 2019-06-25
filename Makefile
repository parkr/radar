DOCKER_TAG:=$(shell git rev-parse HEAD)
DOCKER_IMAGE:=parkr/radar:$(DOCKER_TAG)

all: build test

build:
	go install github.com/parkr/radar/...

test:
	go test github.com/parkr/radar/...
	go vet github.com/parkr/radar/...

server: build
	$(shell RADAR_MYSQL_URL='root@/radar_development?parseTime=true' radar -http="localhost:8291" -debug)

docker-build: all
	docker build -t $(DOCKER_IMAGE) .

docker-test: all docker-build
	docker run --rm \
	  --name radar \
	  -e RADAR_HEALTHCHECK_URL=http://0.0.0.0:8291/health \
	  -e RADAR_MYSQL_URL=root@/radar_development?parseTime=true \
	  $(DOCKER_IMAGE) \
	  radar -http=":8291" -debug


docker-release: all docker-build
	docker push $(DOCKER_IMAGE)
