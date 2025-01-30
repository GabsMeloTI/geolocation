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