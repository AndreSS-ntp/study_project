FROM golang:1.23.6-alpine AS builder
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

FROM alpine:3.18
RUN apk add --no-cache postgresql-client
COPY --from=builder /go/bin/goose /usr/bin/goose
COPY ./migrations /migrations
WORKDIR /migrations