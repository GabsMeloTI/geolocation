# GO Geolocation

## Visão Geral
Serviço destinado ao gerenciamento de usuários e suas relações com organizações. Oferece APIs para operações individuais e em lote, incluindo criação, atualização, associação e consulta de dados de usuários e organizações.


## Funcionalidades
1. Gerenciamento de Usuários: Cadastro, atualização e remoção de usuários.
2. Autenticação: Login e geração de Bearer token.
3. Gestão de Endereços: Atualização de informações residenciais dos usuários.
4. Gestão de Planos: Atribuição de planos aos usuários.
5. Gerenciamento de Motoristas: Cadastro, edição e remoção de motoristas vinculados a usuários.
6. Gerenciamento de Carrocerias: Cadastro, edição e remoção de trailers.
7. Gerenciamento de Caminhões: Cadastro, edição e remoção de unidades de trator.
8. Gerenciamento de Anúncios: Cadastro, edição e remoção de anúncios de transporte.


## Tecnologias utilizadas


- **[Golang](https://golang.org/):** Linguagem principal para implementação.
- **[PostgreSQL](https://www.postgresql.org/):** Banco de dados relacional para metadados.
- **[Docker & Docker Compose](https://www.docker.com/get-started/):** Containerização e orquestração.
- **[SQLC](https://docs.sqlc.dev/en/stable/tutorials/getting-started-postgresql.html):** Gerador de queries que faz a conversão de SQL em código Go tipado.
- **[Echo](https://echo.labstack.com/):** Framework de desenvolvimento web para a linguagem de programação GoLang.
- **[AWS](https://aws.amazon.com/pt/free/?gclid=CjwKCAiArKW-BhAzEiwAZhWsIBv7tTDV_dj41cYBblkPVNVlXSQLdNwLY9WzMJJLEo4AUbJIXmbMrBoClNYQAvD_BwE&all-free-tier.sort-by=item.additionalFields.SortRank&all-free-tier.sort-order=asc&awsf.Free%20Tier%20Types=*all&awsf.Free%20Tier%20Categories=categories%23compute&trk=d0b462ed-a9ff-4714-8a75-634758c49d4c&sc_channel=ps&ef_id=CjwKCAiArKW-BhAzEiwAZhWsIBv7tTDV_dj41cYBblkPVNVlXSQLdNwLY9WzMJJLEo4AUbJIXmbMrBoClNYQAvD_BwE:G:s&s_kwcid=AL!4422!3!490489331981!e!!g!!aws%20cloud!12024810921!121376982652):** Serviço de cloud altamente confiável.
- **[MeiliSearch](https://www.meilisearch.com/docs/home):** Serviço de documento para pesquisa de palavras.


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
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
    - `200`: Sucesso. Retorna o token jwt do usuário registrado.
      
    ```json
    {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
    ```

**POST** `/v2/login`  
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
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
    - `200`: Sucesso. Retorna os dados do usuário.
          
    ```json
    {
        "id": 1,
        "name": "User Name",
        "email": "user@example.com",
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
    ```


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
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
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

**PUT** `/user/personal/update`  
Atualiza as informações pessoais do usuário logado.

- **Body**:
    - JSON no formato do schema `model.UserUpdatePasswordRequest`.
  ```json
  {
    "id": 15,
    "name": "John Doe",
    "document": "123.456.789-00",
    "email": "johndoe@example.com",
    "phone": "+55 11 91234-5678",
    "cep": "12345-678"
  }
  ```
- **Responses**:
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
    - `200`: Sucesso. Retorna um JSON `GetUserResponse`.
      ````json
      {
        "id": 15,
        "name": "John Doe",
        "email": "johndoe@example.com",
        "document": "123.456.789-00",
        "phone": "+55 11 91234-5678"
      }
      ````

**PUT** `/user/address/update`  
Atualiza as informações de endereço do usuário logado.

- **Body**:
    - JSON no formato do schema `model.UserUpdatePasswordRequest`.
  ```json
  {
    "id": 15,
    "state": "São Paulo",
    "city": "São Paulo",
    "neighborhood": "Jardins",
    "street": "Av. Paulista",
    "street_number": "1234",
    "complement": "Apto 101",
    "cep": "01310100"
  }
  ```
- **Responses**:
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
    - `200`: Sucesso. Retorna um JSON `GetUserResponse`.
      ````json
      {
        "id": 15,
        "state": "São Paulo",
        "city": "São Paulo",
        "neighborhood": "Jardins",
        "street": "Av. Paulista",
        "street_number": "1234",
        "cep": "01310100",
        "complement": "Apto 101"
      }
      ````

**POST** `/user/plan`  
Atribui um plano para o usuário logado.

- **Body**:
    - JSON no formato do schema `model.UserUpdatePasswordRequest`.
  ```json
  {
    "id_plan": 2,
    "annual": false
  }
  ```
- **Responses**:
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
    - `200`: Sucesso. Retorna um JSON `GetUserResponse`.
      ````json
      {
        "id": 7,
        "id_user": 25,
        "id_plan": 1,
        "annual": false,
        "active": true,
        "active_date": "2025-03-06T16:43:02.185332Z",
        "expiration_date": "2025-04-05T13:43:01.52508Z",
        "price": 0
      }
      ````

---

### Driver

**POST** `/driver/create`  
Cria um novo mostorista para o usuário logado.

- **Body**:
    - JSON no formato do schema `CreateDriverDto`.
  ```json
  {
    "user_id": 10,
    "birth_date": "1990-05-10T00:00:00Z",
    "cpf": "12345678909",
    "license_number": "92524554641",
    "license_category": "B",
    "license_expiration_date": "2025-12-31T00:00:00Z",
    "state": "SP",
    "city": "São Paulo",
    "neighborhood": "Centro",
    "street": "Rua das Flores",
    "street_number": "123",
    "phone": "+5511999999999"
  }
  ```
- **Responses**:
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
    - `200`: Sucesso. Retorna um JSON `DriverResponse`.
    ```json
    {
      "id": 10,
      "user_id": 25,
      "name": "",
      "birth_date": "1990-05-10T00:00:00Z",
      "cpf": "12345678909",
      "license_number": "92524554641",
      "license_category": "B",
      "license_expiration_date": "2025-12-31T00:00:00Z",
      "cep": "",
      "state": "SP",
      "city": "São Paulo",
      "neighborhood": "Centro",
      "street": "Rua das Flores",
      "street_number": "123",
      "phone": "+5511999999999",
      "complement": "",
      "status": true,
      "created_at": "2025-03-06T16:49:58.752604Z",
      "updated_at": null
    }
    ```

**PUT** `/driver/update`  
Altera o mostorista do usuário logado.

- **Body**:
    - JSON no formato do schema `UpdateDriverDto`.
  ```json
  {
    "id": 10, 
    "birth_date": "1990-05-10T00:00:00Z",
    "license_category": "B",
    "license_expiration_date": "2025-12-31T00:00:00Z",
    "state": "SP",
    "city": "AAAAAAAAAA Paulo",
    "neighborhood": "Centro",
    "street": "Rua das Flores",
    "street_number": "123",
    "phone": "+55 (11) 99999-9999"
  }
  ```
- **Responses**:
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
    - `200`: Sucesso. Retorna um JSON `DriverResponse`.
    ```json
    {
      "id": 10,
      "user_id": 25,
      "name": "",
      "birth_date": "1990-05-10T00:00:00Z",
      "cpf": "12345678909",
      "license_number": "92524554641",
      "license_category": "B",
      "license_expiration_date": "2025-12-31T00:00:00Z",
      "cep": "",
      "state": "SP",
      "city": "AAAAAAAAAA Paulo",
      "neighborhood": "Centro",
      "street": "Rua das Flores",
      "street_number": "123",
      "phone": "+55 (11) 99999-9999",
      "complement": "",
      "status": true,
      "created_at": "2025-03-06T16:49:58.752604Z",
      "updated_at": "2025-03-06T16:51:53.349841Z"
    }
    ```
    
**DELETE** `/driver/delete/{id}`  
Deleta o mostorista do usuário logado.
- **Path Parameter**:
    - `id` (ID do usuário)
- **Responses**:
    - `200`: Sucesso. Retorna um "Success"`.
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.


**GET** `/driver/list`  
Retorna uma lista dos motoristas associados ao usuário autenticado (baseado no token).

- **Responses**:
    - `200`: Sucesso. Retorna um JSON `[]DriverResponse`.
      ````json
      [
        {
            "id": 4,
            "user_id": 10,
            "name": "op",
            "birth_date": "2025-03-05T00:00:00Z",
            "cpf": "55873283842",
            "license_number": "5587",
            "license_category": "5587",
            "license_expiration_date": "2025-03-05T00:00:00Z",
            "cep": "78",
            "state": "5587",
            "city": "5587",
            "neighborhood": "5587",
            "street": "5587",
            "street_number": "5587",
            "phone": "5587",
            "complement": "jk",
            "status": true,
            "created_at": "2025-03-05T13:18:32.836262Z",
            "updated_at": null
        }
      ]
      ````
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.

---

### Trailer

**POST** `/trailer/create`  
Vincular um usuário a varias organizações.

- **Body**:
  - JSON no formato do schema `CreateTrailerDto`.
  ```json
  {
    "license_plate": "XYZ9876",
    "chassis": "CHS123456789",
    "body_type": "tank",
    "load_capacity": 15.5,
    "length": 13.2,
    "width": 2.5,
    "height": 3.8,
    "state": "sp",
    "renavan": "233434",
    "axles": 1
  }
  ```
- **Responses**:
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
    - `200`: Sucesso. Retorna um JSON `TrailerResponse`.
    ```json
    {
      "id": 19,
      "user_id": 25,
      "license_plate": "XYZ9876",
      "chassis": "CHS123456789",
      "body_type": "tank",
      "load_capacity": 15.5,
      "length": 13.2,
      "width": 3.8,
      "height": 2.5,
      "state": "sp",
      "renavan": "233434",
      "axles": 1,
      "status": true,
      "created_at": "2025-03-06T16:55:48.559127Z",
      "updated_at": null
    }
    ```

**PUT** `/trailer/update`  
Vincular um usuário a varias organizações.

- **Body**:
    - JSON no formato do schema `UpdateTrailerDto`.
  ```json
  {
    "id": 19,
    "license_plate": "jjjjjj",
    "chassis": "jkj",
    "body_type": "tank",
    "load_capacity": 15.5,
    "length": 13.2,
    "width": 2.5,
    "height": 3.8,
    "state": "op",
    "renavan": "op",
    "axles": 1
  }
  ```
- **Responses**:
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
    - `200`: Sucesso. Retorna um JSON `TrailerResponse`.
    ```json
    {
      "id": 19,
      "user_id": 25,
      "license_plate": "jjjjjj",
      "chassis": "jkj",
      "body_type": "tank",
      "load_capacity": 15.5,
      "length": 13.2,
      "width": 3.8,
      "height": 2.5,
      "state": "op",
      "renavan": "op",
      "axles": 1,
      "status": true,
      "created_at": "2025-03-06T16:55:48.559127Z",
      "updated_at": "2025-03-06T16:57:15.281931Z"
    }
    ```
    
**DELETE** `/trailer/delete/{id}`  
Retorna uma lista das organizações associadas a um usuário específico.
- **Path Parameter**:
    - `id` (ID do usuário)
- **Responses**:
    - `200`: Sucesso. Retorna um "Success"`.
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.


**GET** `/trailer/list`  
Retorna uma lista daS carroceirias associadas ao usuário autenticado (baseado no token).

- **Responses**:
    - `200`: Sucesso. Retorna um JSON `[]TrailerResponse`.
      ````json
      [
        {
            "id": 10,
            "user_id": 25,
            "license_plate": "ABC-123",
            "chassis": "chassi",
            "body_type": "tipo_1",
            "load_capacity": 1,
            "length": 1,
            "width": 1,
            "height": 1,
            "state": "SP",
            "renavan": "renavan",
            "axles": 1,
            "status": true,
            "created_at": "2025-03-06T13:30:40.213952Z",
            "updated_at": null
        },
        {
            "id": 11,
            "user_id": 25,
            "license_plate": "ABC-123",
            "chassis": "chassi",
            "body_type": "tipo_1",
            "load_capacity": 1,
            "length": 1,
            "width": 1,
            "height": 1,
            "state": "SP",
            "renavan": "renavan",
            "axles": 1,
            "status": true,
            "created_at": "2025-03-06T13:34:07.998224Z",
            "updated_at": null
        }
      ]
      ````
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.

---

### Tractor Unit

**POST** `/tractor-unit/create`  
Vincular um usuário a varias organizações.

- **Body**:
  - JSON no formato do schema `CreateTractorUnitDto`.
  ```json
  {
    "driver_id": 12,
    "model": "FH16",
    "brand": "Volvo",
    "license_plate": "ABC1234",
    "state": "SP",
    "chassis": "CHS987654321",
    "manufacture_year": 2015,
    "engine_power": "400 horses",
    "unit_type": "tractor unit",
    "can_couple": true,
    "height": 3.75,
    "renavan": "op",
    "capacity": "op",
    "width": 32,
    "length": 32,
    "color": "opa"
  }
  ```
- **Responses**:
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
    - `200`: Sucesso. Retorna um JSON `TractorUnitResponse`.
    ```json
    {
      "id": 6,
      "user_id": 25,
      "license_plate": "ABC1234",
      "driver_id": 12,
      "chassis": "CHS987654321",
      "brand": "Volvo",
      "model": "FH16",
      "manufacture_year": 2015,
      "engine_power": "400 horses",
      "unit_type": "tractor unit",
      "height": 3.75,
      "state": "SP",
      "renavan": "op",
      "capacity": "op",
      "width": 32,
      "length": 32,
      "color": "opa",
      "status": true,
      "created_at": "2025-03-06T17:02:12.888301Z",
      "updated_at": null
    }
    ```

**PUT** `/tractor-unit/update`  
Vincular um usuário a varias organizações.

- **Body**:
    - JSON no formato do schema `UpdateTractorUnitDto`.
  ```json
  {
    "id": 19,
    "license_plate": "jjjjjj",
    "chassis": "jkj",
    "body_type": "tank",
    "load_capacity": 15.5,
    "length": 13.2,
    "width": 2.5,
    "height": 3.8,
    "state": "op",
    "renavan": "op",
    "axles": 1
  }
  ```
- **Responses**:
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
    - `200`: Sucesso. Retorna um JSON `TractorUnitResponse`.
    ```json
    {
      "id": 6,
      "user_id": 25,
      "license_plate": "ABC1234",
      "driver_id": 12,
      "chassis": "CHS987654321",
      "brand": "Volvo",
      "model": "FH16",
      "manufacture_year": 2015,
      "engine_power": "400 horses",
      "unit_type": "tractor unit",
      "height": 3.75,
      "state": "SP",
      "renavan": "op",
      "capacity": "op",
      "width": 32,
      "length": 32,
      "color": "opa",
      "status": true,
      "created_at": "2025-03-06T17:02:12.888301Z",
      "updated_at": "2025-03-06T17:03:25.756224Z"
    }
    ```
    
**DELETE** `/tractor-unit/delete/{id}`  
Retorna uma lista das organizações associadas a um usuário específico.
- **Path Parameter**:
    - `id` (ID do usuário)
- **Responses**:
    - `200`: Sucesso. Retorna um "Success"`.
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.


**GET** `/tractor-unit/list`  
Retorna uma lista dos cavalos associados ao usuário autenticado (baseado no token).

- **Responses**:
    - `200`: Sucesso. Retorna um JSON `TractorUnitResponse`.
      ````json
      [
        {
            "id": 5,
            "user_id": 25,
            "license_plate": "ABC1234",
            "driver_id": 12,
            "chassis": "CHS987654321",
            "brand": "Volvo",
            "model": "FH16",
            "manufacture_year": 2015,
            "engine_power": "400 horses",
            "unit_type": "tractor unit",
            "height": 3.75,
            "state": "SP",
            "renavan": "",
            "capacity": "",
            "width": 0,
            "length": 0,
            "color": "",
            "status": true,
            "created_at": "2025-03-06T17:01:35.544646Z",
            "updated_at": null
        }
      ]
      ````
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.

---

### Address

**GET** `/address/find`  
Busca um endereço correspondente com a query digitada.

- **Query**:
  - String com o endereço a ser buscado `q`.
  
- **Responses**:
  - `400`: Requisição inválida. Retorna uma mensagem de erro.
  - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
  - `200`: Sucesso. Retorna uma lista JSON `AddressResponse`.
  ````json
  {
  "id": 1,
  "street": "Rua Exemplo",
  "neighborhood": "Bairro Exemplo",
  "city": "Cidade Exemplo",
  "state": "SP",
  "number": "123",
  "cep": "12345-678",
  "latitude": -23.55052,
  "longitude": -46.633308
  }
  ````

### Advertisement

**POST** `/advertisement/create`  
Vincular um usuário a varias organizações.

- **Body**:
  - JSON no formato do schema `CreateAdvertisementDto`.
  ```json
  {
    "destination": "São Paulo",
    "origin": "Rio de Janeiro",
    "destination_lat": -23.55052,
    "destination_lng": -46.6333,
    "origin_lat": -22.9068,
    "origin_lng": -43.1728,
    "distance": 90,
    "pickup_date": "2025-04-06T10:00:00Z",
    "delivery_date": "2025-04-07T18:00:00Z",
    "expiration_date": "2025-04-08T10:00:00Z",
    "title": "Transporte de carga",
    "cargo_type": "Geral",
    "cargo_species": "Sólida",
    "cargo_weight": 50,
    "vehicles_accepted": "Caminhão, Van",
    "trailer": "Refrigerado",
    "requires_tarp": true,
    "tracking": true,
    "agency": false,
    "description": "Transporte de produtos perecíveis.",
    "payment_type": "Antecipado",
    "advance": "20%",
    "toll": true,
    "situation": "Ativo",
    "price": 500,
    "state": "sp",
    "city": "sp",
    "complement": "sp",
    "neighborhood": "sp",
    "street": "sp",
    "street_number": "90",
    "cep": "09405905"
  }
  ```
- **Responses**:
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
    - `200`: Sucesso. Retorna um JSON `AdvertisementResponse`.
    ```json
    {
      "id": 3,
      "user_id": 25,
      "destination": "São Paulo",
      "origin": "Rio de Janeiro",
      "destination_lat": -23.55052,
      "destination_lng": -46.6333,
      "origin_lat": -22.9068,
      "origin_lng": -43.1728,
      "distance": 90,
      "pickup_date": "2025-04-06T10:00:00Z",
      "delivery_date": "2025-04-07T18:00:00Z",
      "expiration_date": "2025-04-08T10:00:00Z",
      "title": "Transporte de carga",
      "cargo_type": "Geral",
      "cargo_species": "Sólida",
      "cargo_weight": 50,
      "vehicles_accepted": "Caminhão, Van",
      "trailer": "Refrigerado",
      "requires_tarp": true,
      "tracking": true,
      "agency": false,
      "description": "Transporte de produtos perecíveis.",
      "payment_type": "Antecipado",
      "advance": "20%",
      "toll": true,
      "situation": "ativo",
      "price": 500,
      "state": "sp",
      "city": "sp",
      "complement": "sp",
      "neighborhood": "sp",
      "street": "sp",
      "street_number": "90",
      "cep": "09405905",
      "status": true,
      "created_at": "2025-03-06T17:44:15.014667Z",
      "updated_at": null,
      "created_who": "Joao Victor",
      "updated_who": null
    }
    ```

**PUT** `/advertisement/update`  
Vincular um usuário a varias organizações.

- **Body**:
    - JSON no formato do schema `UpdateAdvertisementDto`.
  ```json
  {
    "id": 2,
    "destination": "São Paulo",
    "origin": "Rio de Janeiro",
    "destination_lat": -23.55052,
    "destination_lng": -46.6333,
    "origin_lat": -22.9068,
    "origin_lng": -43.1728,
    "distance": 90,
    "pickup_date": "2025-04-06T10:00:00Z",
    "delivery_date": "2025-04-07T18:00:00Z",
    "expiration_date": "2025-04-08T10:00:00Z",
    "title": "Transporte de carga",
    "cargo_type": "Geral",
    "cargo_species": "Sólida",
    "cargo_weight": 50,
    "vehicles_accepted": "Caminhão, Van",
    "trailer": "Refrigerado",
    "requires_tarp": true,
    "tracking": true,
    "agency": false,
    "description": "Transporte de produtos perecíveis.",
    "payment_type": "Antecipado",
    "advance": "20%",
    "toll": true,
    "situation": "Ativo",
    "price": 500,
    "state": "sp",
    "city": "sp",
    "complement": "sp",
    "neighborhood": "sp",
    "street": "sp",
    "street_number": "90",
    "cep": "09405905"
  }
  ```
- **Responses**:
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
    - `200`: Sucesso. Retorna um JSON `AdvertisementResponse`.
    ```json
    {
      "id": 2,
      "user_id": 25,
      "destination": "São Paulo",
      "origin": "Rio de Janeiro",
      "destination_lat": -23.55052,
      "destination_lng": -46.6333,
      "origin_lat": -22.9068,
      "origin_lng": -43.1728,
      "distance": 90,
      "pickup_date": "2025-04-06T10:00:00Z",
      "delivery_date": "2025-04-07T18:00:00Z",
      "expiration_date": "2025-04-08T10:00:00Z",
      "title": "Transporte de carga",
      "cargo_type": "Geral",
      "cargo_species": "Sólida",
      "cargo_weight": 50,
      "vehicles_accepted": "Caminhão, Van",
      "trailer": "Refrigerado",
      "requires_tarp": true,
      "tracking": true,
      "agency": false,
      "description": "Transporte de produtos perecíveis.",
      "payment_type": "Antecipado",
      "advance": "20%",
      "toll": true,
      "situation": "ativo",
      "price": 500,
      "state": "sp",
      "city": "sp",
      "complement": "sp",
      "neighborhood": "sp",
      "street": "sp",
      "street_number": "90",
      "cep": "09405905",
      "status": true,
      "created_at": "2025-03-06T17:43:34.167385Z",
      "updated_at": "2025-03-06T17:45:10.212302Z",
      "created_who": "Joao Victor",
      "updated_who": "Joao Victor"
    }
    ```
    
**DELETE** `/advertisement/delete/{id}`  
Retorna uma lista das organizações associadas a um usuário específico.
- **Path Parameter**:
    - `id` (ID do usuário)
- **Responses**:
    - `200`: Sucesso. Retorna um "Success"`.
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.


**GET** `/advertisement/list`  
Retorna uma lista de todos os anúncios (com todas informações).

- **Responses**:
    - `200`: Sucesso. Retorna um JSON `[]AdvertisementResponseAll`.
      ````json
      [
        {
            "id": 1,
            "user_id": 25,
            "user_name": "Joao Victor",
            "active_there": "2025-03-05T19:57:21.339004Z",
            "active_duration": "Ativo a: 2024 anos e 2 meses",
            "user_city": "",
            "user_state": "",
            "user_phone": "11939101265",
            "user_email": "jhoni@gmail.com",
            "destination": "São Paulo",
            "origin": "Rio de Janeiro",
            "destination_lat": -23.55052,
            "destination_lng": -46.6333,
            "origin_lat": -22.9068,
            "origin_lng": -43.1728,
            "distance": 90,
            "pickup_date": "2025-04-06T10:00:00Z",
            "delivery_date": "2025-04-07T18:00:00Z",
            "expiration_date": "2025-04-08T10:00:00Z",
            "title": "Transporte de carga",
            "cargo_type": "Geral",
            "cargo_species": "Sólida",
            "cargo_weight": 50,
            "vehicles_accepted": "Caminhão, Van",
            "trailer": "Refrigerado",
            "requires_tarp": true,
            "tracking": true,
            "agency": false,
            "description": "Transporte de produtos perecíveis.",
            "payment_type": "Antecipado",
            "advance": "20%",
            "toll": true,
            "situation": "ativo",
            "active_freight": 2,
            "price": 500,
            "state": "",
            "city": "",
            "complement": "",
            "neighborhood": "",
            "street": "",
            "street_number": "",
            "cep": "",
            "created_at": "2025-03-06T17:43:00.167197Z",
            "created_who": "Joao Victor"
        }
      ]
      ````
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
 

**GET** `public/advertisement/list`  
Retorna uma lista de todos os anúncios (com menos informações).

- **Responses**:
    - `200`: Sucesso. Retorna um JSON `[]AdvertisementResponseNoUser`.
      ````json
      [
        {
            "id": 1,
            "user_id": 25,
            "user_name": "Joao Victor",
            "active_there": "2025-03-05T19:57:21.339004Z",
            "active_duration": "Ativo a: 2024 anos e 2 meses",
            "user_city": "",
            "user_state": "",
            "user_phone": "11939101265",
            "user_email": "jhoni@gmail.com",
            "destination": "São Paulo",
            "origin": "Rio de Janeiro",
            "destination_lat": -23.55052,
            "destination_lng": -46.6333,
            "origin_lat": -22.9068,
            "origin_lng": -43.1728,
            "distance": 90,
            "pickup_date": "2025-04-06T10:00:00Z",
            "delivery_date": "2025-04-07T18:00:00Z",
            "expiration_date": "2025-04-08T10:00:00Z",
            "title": "Transporte de carga",
            "cargo_type": "Geral",
            "cargo_species": "Sólida",
            "cargo_weight": 50,
            "vehicles_accepted": "Caminhão, Van",
            "trailer": "Refrigerado",
            "requires_tarp": true,
            "tracking": true,
            "agency": false,
            "description": "Transporte de produtos perecíveis.",
            "payment_type": "Antecipado",
            "advance": "20%",
            "toll": true,
            "situation": "ativo",
            "active_freight": 2,
            "price": 500,
            "state": "",
            "city": "",
            "complement": "",
            "neighborhood": "",
            "street": "",
            "street_number": "",
            "cep": "",
            "created_at": "2025-03-06T17:43:00.167197Z",
            "created_who": "Joao Victor"
        }
      ]
      ````
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.

---

### Attachment

**POST** `/attach/upload`  
Realiza o carregamento de um arquivo para o bucket S3.
- **Body**:
  - Form-data no formato do schema `*multipart.Form`.
  ```form-data
  files_input: archive.png - FILE
  description: Teste upload - TEXT
  ```
- **Responses**:
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
    - `200`: Sucesso. Retorna um "Success".
    
    
**PUT** `/attach/delete/{id}`  
Realiza o delete de um arquivo para o bucket S3.
- **Path Parameter**:
    - `id` (ID do usuário)
- **Responses**:
    - `200`: Sucesso. Retorna um "Success"`.
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.


---

### Route

**GET** `/route/favorite/list`  
Lista as rotas que foram salvas pelo usuário logado.

- **Responses**:
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
    - `200`: Sucesso. Retorna um JSON `FavoriteRouteResponse`.
    ```json
    [
      {
          "id": 5,
          "id_user": 10,
          "origin": "Praça da Sé - Sé, São Paulo - SP, 01001-000, Brazil",
          "destination": "Campina Grande do Sul, State of Paraná, 83430-000, Brazil",
          "waypoints": "",
          "response": {
              "routes": [
                  {
                      "costs": {
                          "tag": 26.409999999999997,
                          "cash": 27.799999999999997,
                          "axles": 2,
                          "tagAndCash": 27.799999999999997,
                          "prepaidCard": 27.799999999999997,
                          "fuel_in_the_hwy": 192,
                          "maximumTollCost": 27.799999999999997,
                          "minimumTollCost": 27.799999999999997,
                          "fuel_in_the_city": 192
                      },
                      "tolls": [
                          {
                              "id": 114,
                              "lat": -23.600039,
                              "lng": -46.812842,
                              "name": "P13 - Régis Bittencourt",
                              "road": "SP-021",
                              "type": "Pedágio",
                              "state": "SP",
                              "tagImg": [
                                  "https://tags-tolls.s3.us-east-1.amazonaws.com/veloe.png",
                                  "https://tags-tolls.s3.us-east-1.amazonaws.com/semparar.png",
                                  "https://tags-tolls.s3.us-east-1.amazonaws.com/moveMais.png",
                                  "https://tags-tolls.s3.us-east-1.amazonaws.com/taggy.png",
                                  "https://tags-tolls.s3.us-east-1.amazonaws.com/conectcar.png"
                              ],
                              "arrival": {
                                  "time": "19m1s",
                                  "distance": "19.02 km"
                              },
                              "country": "Brasil",
                              "tagCost": 3,
                              "cashCost": 3.2,
                              "currency": "BRL",
                              "free_flow": false,
                              "concession": "RODOANEL OESTE",
                              "tagPrimary": [
                                  "veloe",
                                  "semParar",
                                  "moveMais",
                                  "taggy",
                                  "conectCar"
                              ],
                              "pay_free_flow": "",
                              "concession_img": "https://dealership-routes.s3.us-east-1.amazonaws.com/ccr_rodoanel.png",
                              "prepaidCardCost": 3
                          },
                          .....
    ]
    ```
    

**PUT** `/route/favorite/remove/{id}`  
Remove a rota dos favoritos. 
- **Path Parameter**:
    - `id` (ID do usuário)
    
- **Responses**:
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
    - `200`: Sucesso. Retorna um "success".
 

**POST** `/check-route-tolls-easy`  
Funciona como roteirizados, calcula rota de uma origem a um destino (e opcionalmente, paradas)
- **Body**:
    - JSON no formato do schema `FrontInfo`.
  ```json
  {
    "origin": "são paulo",
    "destination": "rio grande do sul",
    "consumptionCity": 12,
    "consumptionHwy": 12,
    "price": 6.0,
    "axles": 2,
    "type": "Auto",
    "waypoints": [],
    "typeRoute": "EFICIENTE",
    "public_or_private": "private",
    "favorite":true
  }
  ```
- **Responses**:
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
    - `200`: Sucesso. Retorna um  JSON `FinalOutput`.
    ```json
    {
      "summary": {
          "location_origin": {
              "location": {
                  "latitude": -23.5503099,
                  "longitude": -46.6342009
              },
              "address": "Praça da Sé - Sé, São Paulo - SP, 01001-000, Brazil"
          },
          "location_destination": {
              "location": {
                  "latitude": -30.0368176,
                  "longitude": -51.2089887
              },
              "address": "Porto Alegre, RS, Brazil"
          },
          "all_stopping_points": null,
          "fuel_price": {
              "price": 6,
              "currency": "BRL",
              "units": "km",
              "fuel_unit": "liter"
          },
          "fuel_efficiency": {
              "city": 12,
              "hwy": 12,
              "units": "km",
              "fuel_unit": "liter"
          }
      },
      "routes": [
          {
              "summary": {
                  "route_type": "fatest",
                  "hasTolls": true,
                  "distance": {
                      "text": "1132 km",
                      "value": 1132273.1
                  },
                  "duration": {
                      "text": "14h50m21s",
                      "value": 53421.5
                  },
                  "url": "https://www.google.com/maps/dir/?api=1&origin=Pra%C3%A7a+da+S%C3%A9+-+S%C3%A9%2C+S%C3%A3o+Paulo+-+SP%2C+01001-000%2C+Brazil&destination=Porto+Alegre%2C+RS%2C+Brazil",
                  "url_waze": "https://www.waze.com/pt-BR/live-map/directions/br?to=place.ChIJHctqVtKcGZURH-mHn6gRMWA&from=place.EjdQcmHDp2EgZGEgU8OpIC0gU8OpLCBTw6NvIFBhdWxvIC0gU1AsIDAxMDAxLTAwMCwgQnJhemlsIi4qLAoUChIJM0KuqqtZzpQRscVLca9vGNkSFAoSCb10CyKqWc6UERlMYGPSqhJk&time=1741255655835&reverse=yes"
              },
              "costs": {
                  "tagAndCash": 79.9,
                  "fuel_in_the_city": 566,
                  "fuel_in_the_hwy": 566,
                  "tag": 75.905,
                  "cash": 79.9,
                  "prepaidCard": 79.9,
                  "maximumTollCost": 79.9,
                  "minimumTollCost": 79.9,
                  "axles": 2
              },
              "tolls": [
                  {
                      "id": 114,
                      "lat": -23.600039,
                      "lng": -46.812842,
                      "name": "P13 - Régis Bittencourt",
                      "concession": "RODOANEL OESTE",
                      "concession_img": "https://dealership-routes.s3.us-east-1.amazonaws.com/ccr_rodoanel.png",
                      "road": "SP-021",
                      "state": "SP",
                      "country": "Brasil",
                      "type": "Pedágio",
                      "tagCost": 3,
                      "cashCost": 3.2,
                      "currency": "BRL",
                      "prepaidCardCost": 3,
                      "arrival": {
                          "distance": "19.02 km",
                          "time": "19m1s"
                      },
                      "tagPrimary": [
                          "veloe",
                          "semParar",
                          "moveMais",
                          "taggy",
                          "conectCar"
                      ],
                      "tagImg": [
                          "https://tags-tolls.s3.us-east-1.amazonaws.com/veloe.png",
                          "https://tags-tolls.s3.us-east-1.amazonaws.com/semparar.png",
                          "https://tags-tolls.s3.us-east-1.amazonaws.com/moveMais.png",
                          "https://tags-tolls.s3.us-east-1.amazonaws.com/taggy.png",
                          "https://tags-tolls.s3.us-east-1.amazonaws.com/conectcar.png"
                      ],
                      "free_flow": false,
                      "pay_free_flow": ""
                  },
                  .....
    ]
    ```  

**POST** `/check-route-tolls-simpplify`  
Funciona como roteirizados, calcula rota de uma origem a um destino (e opcionalmente, paradas)

- **Body**:
    - JSON no formato do schema `FrontInfo`.
  ```json
  {
    "origin": "são paulo",
    "destination": "rio grande do sul",
    "consumptionCity": 12,
    "consumptionHwy": 12,
    "price": 6.0,
    "axles": 2,
    "type": "Auto",
    "waypoints": [],
    "typeRoute": "EFICIENTE",
    "public_or_private": "private",
    "favorite":true
  }
  ```
- **Responses**:
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
    - `200`: Sucesso. Retorna um JSON `FinalOutput`.
    ```json
    {
      "summary": {
          "location_origin": {
              "location": {
                  "latitude": -23.5503099,
                  "longitude": -46.6342009
              },
              "address": "Praça da Sé - Sé, São Paulo - SP, 01001-000, Brazil"
          },
          "location_destination": {
              "location": {
                  "latitude": -30.0368176,
                  "longitude": -51.2089887
              },
              "address": "Porto Alegre, RS, Brazil"
          },
          "all_stopping_points": null,
          "fuel_price": {
              "price": 6,
              "currency": "BRL",
              "units": "km",
              "fuel_unit": "liter"
          },
          "fuel_efficiency": {
              "city": 12,
              "hwy": 12,
              "units": "km",
              "fuel_unit": "liter"
          }
      },
      "routes": [
          {
              "summary": {
                  "route_type": "fatest",
                  "hasTolls": true,
                  "distance": {
                      "text": "1132 km",
                      "value": 1132273.1
                  },
                  "duration": {
                      "text": "14h50m21s",
                      "value": 53421.5
                  },
                  "url": "https://www.google.com/maps/dir/?api=1&origin=Pra%C3%A7a+da+S%C3%A9+-+S%C3%A9%2C+S%C3%A3o+Paulo+-+SP%2C+01001-000%2C+Brazil&destination=Porto+Alegre%2C+RS%2C+Brazil",
                  "url_waze": "https://www.waze.com/pt-BR/live-map/directions/br?to=place.ChIJHctqVtKcGZURH-mHn6gRMWA&from=place.EjdQcmHDp2EgZGEgU8OpIC0gU8OpLCBTw6NvIFBhdWxvIC0gU1AsIDAxMDAxLTAwMCwgQnJhemlsIi4qLAoUChIJM0KuqqtZzpQRscVLca9vGNkSFAoSCb10CyKqWc6UERlMYGPSqhJk&time=1741255655835&reverse=yes"
              },
              "costs": {
                  "tagAndCash": 79.9,
                  "fuel_in_the_city": 566,
                  "fuel_in_the_hwy": 566,
                  "tag": 75.905,
                  "cash": 79.9,
                  "prepaidCard": 79.9,
                  "maximumTollCost": 79.9,
                  "minimumTollCost": 79.9,
                  "axles": 2
              },
              "tolls": [
                  {
                      "id": 114,
                      "lat": -23.600039,
                      "lng": -46.812842,
                      "name": "P13 - Régis Bittencourt",
                      "concession": "RODOANEL OESTE",
                      "concession_img": "https://dealership-routes.s3.us-east-1.amazonaws.com/ccr_rodoanel.png",
                      "road": "SP-021",
                      "state": "SP",
                      "country": "Brasil",
                      "type": "Pedágio",
                      "tagCost": 3,
                      "cashCost": 3.2,
                      "currency": "BRL",
                      "prepaidCardCost": 3,
                      "arrival": {
                          "distance": "19.02 km",
                          "time": "19m1s"
                      },
                      "tagPrimary": [
                          "veloe",
                          "semParar",
                          "moveMais",
                          "taggy",
                          "conectCar"
                      ],
                      "tagImg": [
                          "https://tags-tolls.s3.us-east-1.amazonaws.com/veloe.png",
                          "https://tags-tolls.s3.us-east-1.amazonaws.com/semparar.png",
                          "https://tags-tolls.s3.us-east-1.amazonaws.com/moveMais.png",
                          "https://tags-tolls.s3.us-east-1.amazonaws.com/taggy.png",
                          "https://tags-tolls.s3.us-east-1.amazonaws.com/conectcar.png"
                      ],
                      "free_flow": false,
                      "pay_free_flow": ""
                  },
                  .....
    ]
    ```


**POST** `/public/check-route-tolls`  
Funciona como roteirizados, calcula rota de uma origem a um destino (e opcionalmente, paradas)
OBS: Só é possível fazer duas requisições.

- **Body**:
    - JSON no formato do schema `FrontInfo`.
  ```json
  {
    "origin": "são paulo",
    "destination": "rio grande do sul",
    "consumptionCity": 12,
    "consumptionHwy": 12,
    "price": 6.0,
    "axles": 2,
    "type": "Auto",
    "waypoints": [],
    "typeRoute": "EFICIENTE",
    "public_or_private": "private",
    "favorite":true
  }
  ```
- **Responses**:
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
    - `200`: Sucesso. Retorna um JSON `FinalOutput`.
    ```json
    {
      "summary": {
          "location_origin": {
              "location": {
                  "latitude": -23.5503099,
                  "longitude": -46.6342009
              },
              "address": "Praça da Sé - Sé, São Paulo - SP, 01001-000, Brazil"
          },
          "location_destination": {
              "location": {
                  "latitude": -30.0368176,
                  "longitude": -51.2089887
              },
              "address": "Porto Alegre, RS, Brazil"
          },
          "all_stopping_points": null,
          "fuel_price": {
              "price": 6,
              "currency": "BRL",
              "units": "km",
              "fuel_unit": "liter"
          },
          "fuel_efficiency": {
              "city": 12,
              "hwy": 12,
              "units": "km",
              "fuel_unit": "liter"
          }
      },
      "routes": [
          {
              "summary": {
                  "route_type": "fatest",
                  "hasTolls": true,
                  "distance": {
                      "text": "1132 km",
                      "value": 1132273.1
                  },
                  "duration": {
                      "text": "14h50m21s",
                      "value": 53421.5
                  },
                  "url": "https://www.google.com/maps/dir/?api=1&origin=Pra%C3%A7a+da+S%C3%A9+-+S%C3%A9%2C+S%C3%A3o+Paulo+-+SP%2C+01001-000%2C+Brazil&destination=Porto+Alegre%2C+RS%2C+Brazil",
                  "url_waze": "https://www.waze.com/pt-BR/live-map/directions/br?to=place.ChIJHctqVtKcGZURH-mHn6gRMWA&from=place.EjdQcmHDp2EgZGEgU8OpIC0gU8OpLCBTw6NvIFBhdWxvIC0gU1AsIDAxMDAxLTAwMCwgQnJhemlsIi4qLAoUChIJM0KuqqtZzpQRscVLca9vGNkSFAoSCb10CyKqWc6UERlMYGPSqhJk&time=1741255655835&reverse=yes"
              },
              "costs": {
                  "tagAndCash": 79.9,
                  "fuel_in_the_city": 566,
                  "fuel_in_the_hwy": 566,
                  "tag": 75.905,
                  "cash": 79.9,
                  "prepaidCard": 79.9,
                  "maximumTollCost": 79.9,
                  "minimumTollCost": 79.9,
                  "axles": 2
              },
              "tolls": [
                  {
                      "id": 114,
                      "lat": -23.600039,
                      "lng": -46.812842,
                      "name": "P13 - Régis Bittencourt",
                      "concession": "RODOANEL OESTE",
                      "concession_img": "https://dealership-routes.s3.us-east-1.amazonaws.com/ccr_rodoanel.png",
                      "road": "SP-021",
                      "state": "SP",
                      "country": "Brasil",
                      "type": "Pedágio",
                      "tagCost": 3,
                      "cashCost": 3.2,
                      "currency": "BRL",
                      "prepaidCardCost": 3,
                      "arrival": {
                          "distance": "19.02 km",
                          "time": "19m1s"
                      },
                      "tagPrimary": [
                          "veloe",
                          "semParar",
                          "moveMais",
                          "taggy",
                          "conectCar"
                      ],
                      "tagImg": [
                          "https://tags-tolls.s3.us-east-1.amazonaws.com/veloe.png",
                          "https://tags-tolls.s3.us-east-1.amazonaws.com/semparar.png",
                          "https://tags-tolls.s3.us-east-1.amazonaws.com/moveMais.png",
                          "https://tags-tolls.s3.us-east-1.amazonaws.com/taggy.png",
                          "https://tags-tolls.s3.us-east-1.amazonaws.com/conectcar.png"
                      ],
                      "free_flow": false,
                      "pay_free_flow": ""
                  },
                  .....
    ]
    ``` 

---

### Chat

**POST** `/chat/create-room`  
Realiza a criação de um sala de bate-papo.

- **Body**:
  - JSON no formato do schema `CreateChatRoomRequest`.
  ```json
  {
    "advertisement_id": 1
  }
  ```
- **Responses**:
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
    - `200`: Sucesso. Retorna um JSON `CreateChatRoomResponse`.
    ```json
    {
       "id": 3,
       "advertisement_id": 1,
       "advertisement_user_id": 10,
       "interested_user_id": 15,
       "created_at": "2025-02-28T18:47:38.10766Z"
    }
    ```

**PUT** `/chat/messages/{room_id}`  
Lista as mensagens de um sala de bate-papo.

- **Path Parameter**:
    - `id` (ID do usuário)
- **Responses**:
    - `400`: Requisição inválida. Retorna uma mensagem de erro.
    - `500`: Erro interno do servidor. Retorna uma mensagem de erro.
    - `200`: Sucesso. Retorna um JSON `MessageResponse`.
    ```json
    [
      {
        "message_id": 3,
        "room_id": 3,
        "user_id": 15,
        "content": "Opa e ai tudo bem",
        "name": "João Dev",
        "profile_picture": "",
        "created_at": "2025-02-28T19:03:37.299482Z"
      }
    ]
    ```
    
 
---

## Instruções para Execução

### Execução Local

1. **Clone o repositório:**
   ```bash
   git clone https://github.com/GabsMeloTI/geolocation.git
   cd geolocation
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
   git clone https://github.com/GabsMeloTI/geolocation.git
   cd geolocation
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
---
## Informações MeiliSearch

Este documento fornece os comandos essenciais para gerenciar e entender o índice addresses no MeiliSearch, uma solução de busca rápida e leve, ideal para substituir consultas complexas em PostgreSQL.

Casos de Uso:

- Busca por endereços com correção de erros (ex.: "av. paulsta" → "Avenida Paulista")
- Ordenação por relevância ou atributos específicos

### 1. **Buscas simples**
  ````bash
    curl -X GET 'http://localhost:7700/indexes/streets/search?q=paulista'
  ````
### 2. **Buscas com filtros**
Configure quais atributos são filtráveis e aplique a busca com filtros
  ````bash
    curl -X GET 'http://localhost:7700/indexes/streets/search' \
    -H 'Content-Type: application/json' \
    --data-binary '{
      "q": "paulista",
      "filter": "uf = SP AND city = 'SAO PAULO'"
    }'
  ````
### 3. **Estatísticas do índice**
  ````bash
    curl -X GET 'http://localhost:7700/indexes/streets/stats'
  ````
  
  Saída 
````json
{
    "numberOfDocuments": 8964424,
    "rawDocumentDbSize": 5971550208,
    "avgDocumentSize": 658,
    "isIndexing": true,
    "numberOfEmbeddings": 0,
    "numberOfEmbeddedDocuments": 0,
    "fieldDistribution": {
        "city": 8964424,
        "city_lat": 8964424,
        "city_lon": 8964424,
        "neighborhood": 8964424,
        "neighborhood_lat": 8964424,
        "neighborhood_lon": 8964424,
        "number": 8964424,
        "state": 8964424,
        "state_lat": 8964424,
        "state_lon": 8964424,
        "street": 8964424,
        "street_id": 8964424,
        "uf": 8964424
    }
}
````

### 4. **Configuração Avançada de Busca**
As rankingRules definem a ordem de prioridade dos resultados.
  ````bash
    curl -X PATCH 'http://localhost:7700/indexes/streets/settings' \
  -H 'Content-Type: application/json' \
  --data-binary '{
    "rankingRules": [
      "words",           # Prioriza documentos com todas as palavras da query
      "typo",            # Penaliza documentos com erros de digitação
      "proximity",       # Prioriza palavras próximas no texto
      "attribute",       # Considera a importância dos campos (searchableAttributes)
      "sort",            # Respeita ordenação explícita (ex.: "sort": ["city_name:asc"])
      "exactness"        # Prioriza correspondências exatas (ex.: "SP" vs "São Paulo")
    ]
  }'
  ````
### 5. **Configuração de Typo Tolerance**
Controla como o MeiliSearch lida com erros de digitação.
  ````bash
   curl -X PATCH 'http://localhost:7700/indexes/streets/settings' \
  -H 'Content-Type: application/json' \
  --data-binary '{
    "typoTolerance": {
      "enabled": true,     # Ativa a correção de erros
      "minWordSizeForTypos": {
        "oneTypo": 5,      # Tamanho mínimo para permitir 1 erro (default: 4)
        "twoTypos": 8      # Tamanho mínimo para permitir 2 erros (default: 8)
      },
      "disableOnWords": ["SP", "RJ"],  # Desativa correção para estas palavras
      "disableOnAttributes": ["cep"]   # Desativa correção em campos específicos
    }
  }'
  ````

#### Explicação dos Parâmetros:

- `oneTypo`: Palavras com 5+ letras podem ter 1 erro ("paulsta" → "paulista").
- `twoTypos`: Palavras com 8+ letras podem ter 2 erros ("avenuda paulsta" → "avenida paulista").
- `disableOnWords`: Ignora correção em siglas (ex.: "SP").

## **Tasks do MeiliSearch**
Tasks são operações assíncronas no MeiliSearch (ex.: criação de índice, atualização de documentos). Cada task tem um UID e status:

- `enqueued`: Na fila para processamento.
- `processing`: Em execução.
- `succeeded`: Concluída com sucesso.
- `failed`: Falhou (ver detalhes no erro).

### 1. **Listar Todas Tasks**
  ````bash
    curl -X GET 'http://localhost:7700/tasks'
  ````
### 2. **Filtrar Tasks por Status**
  ````bash
    curl -X GET 'http://localhost:7700/tasks?statuses=succeeded,enqueued'
  ````
### 3. **Ver Detalhes de uma Task**
  ````bash
    curl -X GET 'http://localhost:7700/tasks/12345'  # Substitua 12345 pelo UID da task'
  ````

## Observações

- **Variáveis de Ambiente:** As variáveis de ambiente são essenciais para o funcionamento correto da aplicação.
- **Portas:** Ajuste as portas no `docker-compose.yml` caso necessário.
- **Banco de Dados:** Certifique-se de que o banco esteja acessível e configurado corretamente.

---

Seguindo estas instruções, você conseguirá executar o projeto localmente ou em contêineres Docker, com suporte completo para upload de arquivos e armazenamento de metadados.

