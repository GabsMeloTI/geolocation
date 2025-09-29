CREATE TABLE locations (
                           id               BIGSERIAL PRIMARY KEY,
                           type             VARCHAR(50) NOT NULL,
                           address          VARCHAR(50) NULL,
                           id_provider_info BIGINT NOT NULL,
                           color            TEXT NOT NULL,
                           created_at       TIMESTAMP NOT NULL DEFAULT now(),
                           updated_at       TIMESTAMP  NULL,
                           access_id        BIGINT NOT NULL,
                           tenant_id        UUID NOT NULL
);


CREATE TABLE areas (
                      id               BIGSERIAL PRIMARY KEY,
                      locations_id     BIGINT NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
                      latitude         NUMERIC(9,6) NOT NULL,
                      longitude        NUMERIC(9,6) NOT NULL,
                      description      VARCHAR(20) NOT NULL,
                      created_at       TIMESTAMP NOT NULL DEFAULT now(),
                      updated_at       TIMESTAMP  NULL
);
