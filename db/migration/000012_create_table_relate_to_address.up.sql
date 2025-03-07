CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS states (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    uf CHAR(2) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS cities (
    id SERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    state_id INT NOT NULL REFERENCES states(id) ON DELETE CASCADE,
    UNIQUE (name, state_id)
);

CREATE TABLE IF NOT EXISTS neighborhoods (
    id SERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    city_id INT NOT NULL REFERENCES cities(id) ON DELETE CASCADE,
    UNIQUE (name, city_id)
);

CREATE TABLE IF NOT EXISTS streets (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    neighborhood_id INT REFERENCES neighborhoods(id) ON DELETE SET NULL,
    UNIQUE (name, neighborhood_id)
    );

CREATE TABLE IF NOT EXISTS addresses (
    id SERIAL PRIMARY KEY,
    street_id INT NOT NULL REFERENCES streets(id) ON DELETE CASCADE,
    number VARCHAR(50),
    complement TEXT,
    cep CHAR(8) NOT NULL,
    lat DOUBLE PRECISION,
    lon DOUBLE PRECISION,
    UNIQUE (street_id, number, cep)
);

CREATE INDEX idx_streets_search_vector ON streets USING gin(search_vector);
CREATE INDEX idx_neighborhoods_search_vector ON neighborhoods USING gin(search_vector);
CREATE INDEX idx_cities_search_vector ON cities USING gin(search_vector);
CREATE INDEX idx_states_search_vector ON states USING gin(search_vector);