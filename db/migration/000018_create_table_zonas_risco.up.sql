CREATE TABLE IF NOT EXISTS zonas_risco (
  id BIGSERIAL PRIMARY KEY,
  name        VARCHAR(100) NOT NULL,
  cep         VARCHAR(8)      NOT NULL,
  lat         FLOAT NOT NULL,
  lng         FLOAT NOT NULL,
  radius      BIGINT      NOT NULL,
  type BIGINT null,
  status      BOOLEAN      NOT NULL DEFAULT TRUE
);
