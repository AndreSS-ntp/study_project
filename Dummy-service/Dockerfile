FROM golang:1.24.9-alpine
WORKDIR /app/Dummy-service
COPY . .
RUN go build -o ./bin/main ./cmd/dummy-service
EXPOSE 7001

CMD ["./bin/main"]
