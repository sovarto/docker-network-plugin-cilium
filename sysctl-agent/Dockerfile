# syntax=docker/dockerfile:1.3
FROM --platform=${TARGETPLATFORM:-linux/amd64} golang:1.23-alpine AS builder

RUN apk update && apk add git

COPY go.mod /src/
COPY go.sum /src/
RUN --mount=type=cache,target=/go/pkg/mod cd /src/ && go mod download

COPY . /src/
RUN --mount=type=cache,target=/root/.cache/go-build cd /src/ && CGO_ENABLED=0 GOOS=linux go build -o /docker-network-plugin-cilium-sysctl-http

FROM alpine:3.18

RUN apk add -U --no-cache tini

RUN mkdir -p /var/lib/rsyslog

WORKDIR /app

COPY --from=builder /docker-network-plugin-cilium-sysctl-http /

ENTRYPOINT ["tini", "--", "/docker-network-plugin-cilium-sysctl-http"]
