# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Instala dependências do sistema necessárias para building CGO se precisar, tzdata, etc.
RUN apk add --no-cache git tzdata

# Copia os manifests e instala dependências Go (evita re-download desnecessário no cache)
COPY go.mod go.sum ./
RUN go mod download

# Copia todo o código-fonte restante
COPY . .

# Compila a aplicação estática
# CGO_ENABLED=0 garante um binário estático que pode rodar em container scratch ou alpine limpo sem dependências C
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main .

# Run stage
FROM alpine:3.19

# Adiciona certificados e fuso horário para a aplicação funcionar corretamente
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copia o binário construído no stage de builder
COPY --from=builder /app/main .

# Copia os arquivos de migration necessários em tempo de execução
COPY --from=builder /app/db/migration ./db/migration

EXPOSE 8080

CMD ["./main"]
