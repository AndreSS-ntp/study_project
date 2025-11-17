FROM golang:1.24.9-alpine
WORKDIR /app/Storage-service
COPY . .
RUN go build -o ./bin/main ./cmd/storage-service
EXPOSE 7003

CMD ["./bin/main"]
