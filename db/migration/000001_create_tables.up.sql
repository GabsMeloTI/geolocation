CREATE TABLE public.tolls (
                              id bigserial NOT NULL,
                              concessionaria varchar(50) NULL,
                              praca_de_pedagio varchar(50) NULL,
                              ano_do_pnv_snv int4 DEFAULT 2025 NULL,
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
                              tarifa numeric NULL,
                              free_flow bool DEFAULT false NULL,
                              pay_free_flow varchar(100) DEFAULT ''::character varying NULL
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

CREATE TABLE public.balanca (
                                id bigint NOT NULL,
                                concessionaria varchar(50) NOT NULL,
                                km varchar(50) NOT NULL,
                                lat varchar(50) NOT NULL,
                                lng varchar(50) NOT NULL,
                                nome varchar(50) NOT NULL,
                                rodovia varchar(50) NOT NULL,
                                sentido varchar(50) NOT NULL,
                                uf varchar(50)NOT NULL
);

CREATE TABLE public.freight_load (
                                     type_of_load varchar(50) NULL,
                                     two_axes varchar(50) NULL,
                                     three_axes varchar(50) NULL,
                                     four_axes varchar(50) NULL,
                                     five_axes varchar(50) NULL,
                                     six_axes varchar(50) NULL,
                                     seven_axes varchar(50) NULL,
                                     nine_axes varchar(50) NULL,
                                     "name" varchar(50) NULL,
                                     description varchar(128) NULL
);



CREATE TABLE saved_routes (
                              id SERIAL PRIMARY KEY,
                              origin TEXT NOT NULL,
                              destination TEXT NOT NULL,
                              waypoints TEXT NULL,
                              request JSONB NOT NULL,
                              response JSONB NOT NULL,
                              created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                              updated_at TIMESTAMPTZ NULL DEFAULT null,
                              favorite boolean NULL default false,
                              expired_at timestamp not null
);
CREATE UNIQUE INDEX idx_saved_routes_unique ON saved_routes(origin, destination, waypoints);


CREATE TABLE public.token_hist (
                                   id bigserial NOT NULL,
                                   ip varchar(30) NOT NULL,
                                   number_request bigint NOT NULL,
                                   created_at timestamp default now() NOT NULL,
                                   exprited_at timestamp NOT NULL,
                                   valid bool DEFAULT true NULL
);

CREATE TABLE public.route_hist (
                                   id bigserial PRIMARY KEY,
                                   id_user bigint not null,
                                   origin TEXT NOT NULL,
                                   destination TEXT NOT NULL,
                                   waypoints TEXT NULL,
                                   response JSONB NOT NULL,
                                   is_public BOOL NOT NULL,
                                   number_request bigint NOT NULL,
                                   created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

create table favorite_route (
                                id BIGSERIAL PRIMARY KEY,
                                id_user BIGSERIAL NOT NULL,
                                origin TEXT NOT NULL,
                                destination TEXT NOT NULL,
                                waypoints TEXT NULL,
                                response JSONB NOT NULL,
                                created_who varchar not null,
                                created_at timestamp not null default now()
);

