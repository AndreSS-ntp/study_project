FROM golang:1.23.2-alpine
WORKDIR /app/Dummy-service
COPY . .
RUN go build -o main .
EXPOSE 7001

CMD ["./main"]