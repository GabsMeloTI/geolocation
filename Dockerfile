# Stage 1: build
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copia apenas arquivos de dependências primeiro
COPY go.mod go.sum ./
RUN go mod download

# Copia o restante do código
COPY . .

# Usa cache do Go build (novo recurso do BuildKit)
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o main .

# Stage 2: runtime
FROM alpine:3.20

WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 8080
CMD ["./main"]
