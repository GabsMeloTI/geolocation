# Testes da API Zonas de Risco

## Base URL
```
http://localhost:8080
```

## Headers
```
Authorization: Bearer {seu_token_aqui}
Content-Type: application/json
```

---

## 1. Criar Zona de Risco

**Endpoint:** `POST /zonas-risco/create`

**Request Body:**
```json
{
  "name": "Zona de Risco Centro",
  "cep": "01234567",
  "lat": -23.5505,
  "lng": -46.6333,
  "radius": 500
}
```

**Response (200):**
```json
{
  "id": 1,
  "name": "Zona de Risco Centro",
  "cep": "01234567",
  "lat": -23.5505,
  "lng": -46.6333,
  "radius": 500,
  "status": true
}
```

---

## 2. Atualizar Zona de Risco

**Endpoint:** `PUT /zonas-risco/update`

**Request Body:**
```json
{
  "id": 1,
  "name": "Zona de Risco Centro Atualizada",
  "cep": "01234567",
  "lat": -23.5505,
  "lng": -46.6333,
  "radius": 750,
  "status": true
}
```

**Response (200):**
```json
{
  "id": 1,
  "name": "Zona de Risco Centro Atualizada",
  "cep": "01234567",
  "lat": -23.5505,
  "lng": -46.6333,
  "radius": 750,
  "status": true
}
```

---

## 3. Deletar Zona de Risco (Soft Delete)

**Endpoint:** `PUT /zonas-risco/delete/{id}`

**URL Example:** `PUT /zonas-risco/delete/1`

**Response (200):**
```json
"Sucesso"
```

---

## 4. Obter Zona de Risco por ID

**Endpoint:** `GET /zonas-risco/list/{id}`

**URL Example:** `GET /zonas-risco/list/1`

**Response (200):**
```json
{
  "id": 1,
  "name": "Zona de Risco Centro",
  "cep": "01234567",
  "lat": -23.5505,
  "lng": -46.6333,
  "radius": 500,
  "status": true
}
```

---

## 5. Listar Todas as Zonas de Risco

**Endpoint:** `GET /zonas-risco/list`

**Response (200):**
```json
[
  {
    "id": 1,
    "name": "Zona de Risco Centro",
    "cep": "01234567",
    "lat": -23.5505,
    "lng": -46.6333,
    "radius": 500,
    "status": true
  },
  {
    "id": 2,
    "name": "Zona de Risco Norte",
    "cep": "01234568",
    "lat": -23.5400,
    "lng": -46.6200,
    "radius": 300,
    "status": true
  }
]
```

---

## Exemplos de Dados para Teste

### Zona de Risco São Paulo
```json
{
  "name": "Zona de Risco São Paulo",
  "cep": "01310100",
  "lat": -23.5505,
  "lng": -46.6333,
  "radius": 1000
}
```

### Zona de Risco Rio de Janeiro
```json
{
  "name": "Zona de Risco Rio de Janeiro",
  "cep": "20040020",
  "lat": -22.9068,
  "lng": -43.1729,
  "radius": 800
}
```

### Zona de Risco Belo Horizonte
```json
{
  "name": "Zona de Risco Belo Horizonte",
  "cep": "30112000",
  "lat": -19.9167,
  "lng": -43.9345,
  "radius": 600
}
```

---

## Códigos de Status HTTP

- **200 OK** - Operação realizada com sucesso
- **400 Bad Request** - Dados inválidos na requisição
- **500 Internal Server Error** - Erro interno do servidor

## Validações

- **CEP**: Deve ter exatamente 8 dígitos numéricos
- **Latitude**: Deve estar entre -90 e 90
- **Longitude**: Deve estar entre -180 e 180
- **Radius**: Deve ser um número positivo (em metros)
- **Name**: Campo obrigatório, não pode ser vazio

## Observações

- Todas as operações requerem autenticação via token Bearer
- O delete é lógico (soft delete), apenas altera o status para false
- As coordenadas são armazenadas com precisão de 6 casas decimais
- O raio é armazenado em metros
