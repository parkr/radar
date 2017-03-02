FROM golang

WORKDIR /go/src/github.com/parkr/radar

EXPOSE 3306

ADD . .

RUN go version

# Compile a standalone executable
RUN CGO_ENABLED=0 go install github.com/parkr/radar

# Make `radar` available to the `Dockerfile.release` build
CMD [ "./radar" ]
