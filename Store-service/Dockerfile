FROM golang:1.24.9-alpine
WORKDIR /app/Store-service
COPY . .
RUN go build -o ./bin/main ./cmd/store-service
EXPOSE 7002

CMD ["./bin/main"]
