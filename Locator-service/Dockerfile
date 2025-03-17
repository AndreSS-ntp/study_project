FROM golang:1.23.6-alpine
WORKDIR /app/Locator-service
COPY . .
RUN go build -o ./bin/main ./cmd/store-service
EXPOSE 7000

CMD ["./bin/main"]