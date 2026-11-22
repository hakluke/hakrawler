FROM golang:1.27-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /hakrawler .

FROM alpine:3.20
RUN apk --no-cache add ca-certificates
COPY --from=builder /hakrawler /usr/local/bin/hakrawler
ENTRYPOINT ["hakrawler"]
