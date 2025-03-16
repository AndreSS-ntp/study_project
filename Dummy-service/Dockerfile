FROM golang:1.23.6-alpine
WORKDIR /app/Dummy-service
COPY . .
RUN go build -o ./bin/main ./cmd/dummy-service
EXPOSE 7001

CMD ["./bin/main"]
