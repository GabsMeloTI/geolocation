FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod tidy

COPY . .

RUN go build -o main

RUN whoami
RUN id

EXPOSE 8080

CMD ["./main"]