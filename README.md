# GO Geolocation

// TODO:
Serviço destinado ao gerenciamento de usuários e suas relações com organizações.
Oferece APIs para operações individuais e em lote, incluindo criação, atualização, associação e consulta de dados de usuários e organizações.

---

## Funcionalidades
1. **Gerenciamento de Usuários:** Cadastro, atualização de usuário e alteração de senha.
2. **Operações em Lote:** Criação de múltiplos usuários a partir de arquivos Excel.
3. **Associação de Organizações:** Atribuição e consulta de acessos de usuários às organizações.
4. **Geração de Tokens:** Criação de tokens para acesso a plataforma.


## Tecnologias utilizadas


- **[Golang](https://golang.org/):** Linguagem principal para implementação.
- **[PostgreSQL](https://www.postgresql.org/):** Banco de dados relacional para metadados.
- **[Docker & Docker Compose](https://www.docker.com/get-started/):** Containerização e orquestração.
- **[SQLC](https://docs.sqlc.dev/en/stable/tutorials/getting-started-postgresql.html):** Gerador de queries que faz a conversão de SQL em código Go tipado.
- **[Echo](https://echo.labstack.com/):** Framework de desenvolvimento web para a linguagem de programação GoLang.


## Endpoints

---

### Grupo User
**POST** `/v2/create`  
Cria um novo usuário no sistema.

- **Body**:
    - JSON no formato do schema `RequestCreateUser`.
  ```json
    {
    "email": "example@example.com",
    "name": "John Doe",
    "telephone": "+5511999999999",
    "document": "123.456.789-00",
    "password": "@Pass123",
    "confirm_password": "@Pass123",
    "type_person": 1
    //"token": "apenas se for create com login google"
    }
  ```
- **Responses**:
    - `200`: Sucesso. Retorna o token jwt do usuário registrado.
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.



**POST** `/login`  
Loga o usuário no sistema.

- **Body**:
    - JSON no formato do schema `RequestLogin`.
  ```json
  {
  "username": "seu_usuario",
  "password": "sua_senha",
  "token": "seu_token"
  }

  ```
  
- **Responses**:
    - `200`: Sucesso. Retorna os dados do usuário.
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.

**GET** `/user/info`  
Retorna todas as informações do usuário.


- **Responses**:
  - `200`: Sucesso. Retorna um JSON `GetUserResponse`.
      ````json
      {
      "id": 1,
      "name": "João Silva",
      "email": "joao.silva@example.com",
      "created_at": "2025-03-06T12:00:00Z",
      "updated_at": "2025-03-06T12:30:00Z",
      "profile_id": 2,
      "document": "123.456.789-00",
      "state": "SP",
      "city": "São Paulo",
      "neighborhood": "Centro",
      "street": "Rua Exemplo",
      "street_number": "123",
      "phone": "+55 11 91234-5678",
      "profile_picture": "https://example.com/profile.jpg",
      "cep": "01000-000",
      "complement": "Apto 101"
      }
      ````
  - `400`: Requisição inválida. Retorna uma mensagem de erro.
  - `500`: Erro interno do servidor. Retorna uma mensagem de erro.


**PUT** `/user/delete`  
Deleta o usuário logado

- **Responses**:
    - `200`: Sucesso. Retorna uma mensagem de "Success".
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.

**PUT** `/user/update`  
Atualiza a senha de um usuário no sistema.

- **Body**:
    - JSON no formato do schema `UpdateUserRequest`.
  ```json
  {
  "name": "John Doe",
  "profile_picture": "https://example.com/path/to/profile.jpg",
  "state": "SP",
  "city": "São Paulo",
  "neighborhood": "Centro",
  "street": "Rua Exemplo",
  "street_number": "123",
  "phone": "+5511999999999"
  }
  ```
- **Responses**:
    - `200`: Sucesso. Retorna um JSON `GetUserResponse`.
      ````json
      {
       "id": 123456,
       "name": "John Doe",
       "email": "john.doe@example.com",
       "profile_id": 78910,
       "document": "123.456.789-00",
       "state": "SP",
       "city": "São Paulo",
       "neighborhood": "Centro",
       "street": "Rua Exemplo",
       "street_number": "123",
       "phone": "+5511999999999",
       "google_id": "google123456",
       "profile_picture": "https://example.com/path/to/profile.jpg"
      }
      ````
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.

**PUT** `/user/update/password/logged`  
Atualiza a senha de um usuário no sistema e valida se é o usuário logado.

- **Body**:
    - JSON no formato do schema `model.UserUpdatePasswordRequest`.
  ```json
  {
    "password": "newPassword123",
    "confirmPassword": "newPassword123",
    "id": 12345
  }
  ```
- **Responses**:
    - `200`: Sucesso. Retorna uma mensagem de "Success".
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.

**PUT** `/user/delete/{id}`  
Exclui um usuário do sistema.

- **Path Parameters**:
    - `id` (integer): ID do usuário. **Obrigatório**.

- **Responses**:
    - `200`: Sucesso. Retorna uma mensagem de "Success".
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.

---
### Grupo Link

**POST** `/user/link/`  
Vincular um usuário a varias organizações.

- **Body**:
    - JSON no formato do schema `model.LinkUserRequest`.
  ```json
  {
    "id_user": 1,
    "ids_org": [1, 2, 3]
  }
  ```
- **Responses**:
    - `200`: Sucesso. Retorna uma mensagem de "Success".
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.

**GET** `/user/link/list`  
Retorna uma lista das organizações associadas ao usuário autenticado (baseado no token).

- **Responses**:
    - `200`: Sucesso. Retorna um JSON `[]model.LinkUserListResponse`.
      ````json
      [
        {
            "id": 1,
            "ranker": 10,
            "fantasy_name": "João da Silva",
            "company_name": "Empresa Exemplo A",
            "link": "https://example.com/a"
        },
        {
            "id": 2,
            "ranker": 9,
            "fantasy_name": "Maria da Silva",
            "company_name": "Empresa Exemplo B",
            "link": "https://example.com/b"
        }
      ]
      ````
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.

**GET** `/user/link/list/{id}`  
Retorna uma lista das organizações associadas a um usuário específico.
- **Path Parameter**:
    - `id` (ID do usuário)
- **Responses**:
    - `200`: Sucesso. Retorna uma lista JSON `[]model.LinkUserListIDResponse`.
      ````json
      {
          "id": [1, 2, 3]
      }
      ````
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
___

### Grupo: User Organization
**GET** `/user/organization/list`  
Retorna uma lista das organizações às quais o usuário autenticado tem acesso.

- **Responses**:
    - `200`: Sucesso. Retorna uma lista JSON `[]model.UserOrganizationListResponse`.
      ````json
       [
        {
          "id_front": 101,
          "id": 1,
          "cnpj": "12.345.678/0001-99",
          "fantasy_name": "Tech Solutions",
          "logo_url": "https://example.com/logos/tech_solutions.png",
          "logged": true,
          "city": "São Paulo",
          "state": "SP",
          "company_name": "Tech Solutions Ltda."
        }
       ]
      ````
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.

**POST** `/user/organization/access/{id}`  
Gera um token específico para a organização informada, permitindo que o usuário selecione o tenant correspondente.
- **Path Parameter**:
    - `{id}` (ID da organização).
- **Responses**:
    - `200`: Sucesso. Retorna uma string do token `TOKEN`.
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
---

## Instruções para Execução

### Execução Local

1. **Clone o repositório:**
   ```bash
   git clone https://github.com/simpplify-org/GO-user-service.git
   cd GO-user-service
   ```

2. **Instale as dependências:**
   ```bash
   go mod tidy
   ```

3. **Configure as variáveis de ambiente:**
   ```txt
   Crie um arquivo .env com as variáveis de ambiente seguindo o arquivo .env_example.
   ```

4. **Configurar o Banco de Dados PostgreSQL**

   Para inicializar e configurar o banco de dados PostgreSQL com Docker:
   ```bash
   make postgres-setup
   ```

   Para iniciar ou parar o container do PostgreSQL posteriormente:
   ```bash
   make start-postgres
   make stop-postgres
   ```

   Para criar o banco de dados configurado no `.env`(seguindo arquivo .env_example):
   ```bash
   make createdb
   ```

5. **Inicie a aplicação:**
   ```bash
   make run
   ```

### Execução via Docker

1. **Clone o repositório:**
   ```bash
   git clone https://github.com/simpplify-org/GO-user-service.git
   cd GO-user-service
   ```

2. **Configure o arquivo `.env`:**
   ```txt
   Crie um arquivo .env com as variáveis de ambiente seguindo o arquivo .env_example.
   ```

3. **Inicie os contêineres:**
   ```bash
   docker-compose up --build
   ```


## Comandos adicionais do Makefile
### swag
- Gera a documentação Swagger da API, com saída para o diretório docs/.
  ````bash
  make swag
  ````

### sqlc
- Gera os arquivos sqlc do projeto.
  ````bash
  make sqlc
  ````

## Observações

- **Variáveis de Ambiente:** As variáveis de ambiente são essenciais para o funcionamento correto da aplicação.
- **Portas:** Ajuste as portas no `docker-compose.yml` caso necessário.
- **Banco de Dados:** Certifique-se de que o banco esteja acessível e configurado corretamente.

---

Seguindo estas instruções, você conseguirá executar o projeto localmente ou em contêineres Docker, com suporte completo para upload de arquivos e armazenamento de metadados.

