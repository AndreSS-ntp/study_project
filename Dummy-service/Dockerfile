FROM golang:1.23.2-alpine
WORKDIR /app/Dummy-service
COPY . .
RUN mkdir -p /bin && go build -o /bin/main .
EXPOSE 7001

CMD ["/bin/main"]
