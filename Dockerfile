FROM golang:1.11 AS build
WORKDIR /go/src/github.com/parkr/radar
ADD . .
RUN go version
RUN CGO_ENABLED=0 go install github.com/parkr/radar/cmd/...

FROM scratch
COPY --from=build /go/bin/radar /bin/radar
EXPOSE 3306
CMD [ "/bin/radar" ]
