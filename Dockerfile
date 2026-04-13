FROM golang:1.23-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o backlog-mcp .

FROM alpine:3.21

RUN adduser -D -u 1000 backlog && \
    mkdir -p /data/requirements && \
    chown -R backlog:backlog /data

USER backlog

COPY --from=builder /build/backlog-mcp /usr/local/bin/backlog-mcp

ENV BACKLOG_ROOT=/data/requirements

ENTRYPOINT ["backlog-mcp"]


