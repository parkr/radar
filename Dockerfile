FROM golang AS build
WORKDIR /go/src/github.com/parkr/radar
ADD . .
RUN go version
RUN CGO_ENABLED=0 go install github.com/parkr/radar/cmd/...

FROM scratch
COPY --from=build /go/bin/radar /bin/radar
COPY --from=build /go/bin/radar-healthcheck /bin/radar-healthcheck
EXPOSE 3306
HEALTHCHECK --start-period=1ms --interval=30s --retries=1 --timeout=5s CMD [ "/bin/radar-healthcheck" ]
CMD [ "/bin/radar" ]
