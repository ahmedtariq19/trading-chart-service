FROM golang:1.18-alpine

WORKDIR /app

COPY . .

RUN go build -o trading-chart-service .

CMD ["./trading-chart-service"]