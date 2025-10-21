FROM golang:1.24.9-alpine
WORKDIR /app/Locator-service
COPY . .
RUN go build -o ./bin/main ./cmd/locator-service
EXPOSE 7000

CMD ["./bin/main"]