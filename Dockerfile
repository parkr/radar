FROM golang:latest AS build
WORKDIR /go/src/github.com/parkr/radar
COPY . .
RUN go version
RUN CGO_ENABLED=0 go install github.com/parkr/radar/cmd/...
RUN CGO_ENABLED=0 go test github.com/parkr/radar/...

FROM scratch
COPY --from=build /go/bin/radar /bin/radar
COPY --from=build /go/bin/radar-healthcheck /bin/radar-healthcheck
EXPOSE 3306
HEALTHCHECK --start-period=1s --interval=30s --retries=1 --timeout=5s CMD [ "/bin/radar-healthcheck" ]
CMD [ "/bin/radar" ]
