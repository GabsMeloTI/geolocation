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

CREATE TABLE public.unique_ceps (
 id serial PRIMARY KEY,
 street_id int4 NOT NULL,
  "number" varchar(50) NULL,
 complement text NULL,
 cep char(8) NOT NULL,
 lat float8 NULL,
 lon float8 NULL,
 street_name varchar NULL,
 neighborhood_name varchar NULL,
 neighborhood_lat float8 NULL,
 neighborhood_lon float8 NULL,
 city_name varchar NULL,
 city_lat float8 NULL,
 city_lon float8 NULL,
 state_uf varchar(2) NULL,
 state_lat float8 NULL,
 state_lon float8 NULL
);

CREATE INDEX idx_streets_search_vector ON streets USING gin(search_vector);
CREATE INDEX idx_neighborhoods_search_vector ON neighborhoods USING gin(search_vector);
CREATE INDEX idx_cities_search_vector ON cities USING gin(search_vector);
CREATE INDEX idx_states_search_vector ON states USING gin(search_vector);

CREATE TABLE public.unique_ceps (
    id serial4 NOT NULL,
    street_id int4 NOT NULL,
    "number" varchar(50) NULL,
    complement text NULL,
    cep bpchar(8) NOT NULL,
    lat float8 NULL,
    lon float8 NULL,
    street_name varchar NULL,
    neighborhood_name varchar NULL,
    neighborhood_lat float8 NULL,
    neighborhood_lon float8 NULL,
    city_name varchar NULL,
    city_lat float8 NULL,
    city_lon float8 NULL,
    state_uf varchar(2) NULL,
    state_lat float8 NULL,
    state_lon float8 NULL,
    CONSTRAINT unique_ceps_pkey PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS unique_ceps ON ceps (cep);
