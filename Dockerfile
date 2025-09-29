FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copia só os manifests primeiro → cache
COPY go.mod go.sum ./
RUN go mod download

# Copia o restante do código
COPY . .

# Compila
RUN go build -o main

EXPOSE 8080
CMD ["./main"]
