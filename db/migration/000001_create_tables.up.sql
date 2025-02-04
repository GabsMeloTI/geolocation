CREATE TABLE public.tolls (
                              id BIGSERIAL NOT NULL,
                              concessionaria varchar(50) NULL,
                              praca_de_pedagio varchar(50) NULL,
                              ano_do_pnv_snv int4 NULL,
                              rodovia varchar(50) NULL,
                              uf varchar(50) NULL,
                              km_m varchar(50) NULL,
                              municipio varchar(50) NULL,
                              tipo_pista varchar(50) NULL,
                              sentido varchar(50) NULL,
                              situacao varchar(50) NULL,
                              data_da_inativacao varchar(50) NULL,
                              latitude varchar(50) NULL,
                              longitude varchar(50) NULL,
                              tarifa float NULL
);


CREATE TABLE gas_station (
        id BIGSERIAL NOT NULL,
        name varchar(100) NOT NULL,
        latitude varchar(50) NOT NULL,
        longitude varchar(50) NOT NULL,
        address_name varchar(150) NOT NULL,
        municipio varchar(50) NOT NULL,
        specific_point varchar(255) NOT NULL
);

CREATE TABLE public.toll_tags (
                                  id bigserial NOT NULL,
                                  "name" varchar(50) NOT NULL,
                                  dealership_accepts text NOT  NULL,
                                  CONSTRAINT toll_tags_pkey PRIMARY KEY (id)
);


CREATE TABLE saved_routes (
                              id SERIAL PRIMARY KEY,
                              origin TEXT NOT NULL,
                              destination TEXT NOT NULL,
                              waypoints TEXT NULL,
                              response JSONB NOT NULL,
                              created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                              updated_at TIMESTAMPTZ NULL DEFAULT null,
                              favorite boolean NULL default false
);

CREATE UNIQUE INDEX idx_saved_routes_unique ON saved_routes(origin, destination, waypoints);
