FROM golang:1.17-alpine AS builder

RUN apk --no-cache add ca-certificates git
WORKDIR /build
# Fetch dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY hakrawler.go .
RUN CGO_ENABLED=0 go build -o hackrawler .

FROM alpine:latest

RUN apk --no-cache upgrade
RUN mkdir /hackrawler  \
    && adduser -D hackrawler --shell /usr/sbin/nologin \
    && chown -R hackrawler:hackrawler /hackrawler
WORKDIR /hackrawler

COPY --from=builder /build/hackrawler .

USER hackrawler

ENTRYPOINT ["/hackrawler/hackrawler"]
