# Image: golang:1.16.7-alpine3.14
FROM golang@sha256:7e31a85c5b182e446c9e0e6fba57c522902f281a6a5a6cbd25afa17ac48a6b85 as build
RUN GO111MODULE=on go get -v github.com/hakluke/hakrawler

# Image: alpine:3.14.1
FROM alpine@sha256:be9bdc0ef8e96dbc428dc189b31e2e3b05523d96d12ed627c37aa2936653258c
RUN apk -U upgrade --no-cache
COPY --from=build /go/bin/hakrawler /usr/local/bin/hakrawler

RUN adduser \
    --gecos "" \
    --disabled-password \
    hakrawler

USER hakrawler
ENTRYPOINT ["hakrawler"]
