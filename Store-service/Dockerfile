FROM golang:1.23.6-alpine
WORKDIR /app/Store-service
COPY . .
RUN go build -o ./bin/main ./cmd/store-service
EXPOSE 7002

CMD ["./bin/main"]
