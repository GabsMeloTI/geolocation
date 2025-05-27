CREATE TABLE public.route_enterprise (
                                         id bigserial NOT NULL,
                                         origin text NOT NULL,
                                         destination text NOT NULL,
                                         waypoints text NULL,
                                         response jsonb NOT NULL,
                                         status bool DEFAULT true NULL,
                                         created_at timestamp DEFAULT now() NOT NULL,
                                         created_who varchar not null,
                                         tenant_id UUID not null,
                                         access_id bigint not null,
                                         CONSTRAINT route_enterprise_pkey PRIMARY KEY (id)
);


CREATE TABLE public."Organizations" (
                                        id bigserial NOT NULL,
                                        email varchar NOT NULL,
                                        phone varchar NOT NULL,
                                        cnpj varchar NOT NULL,
                                        logo_url varchar NULL,
                                        fantasy_name varchar NOT NULL,
                                        company_name varchar NULL,
                                        state varchar NULL,
                                        city varchar NULL,
                                        status bool NOT NULL,
                                        created_at timestamp DEFAULT now() NOT NULL,
                                        updated_at timestamp NULL,
                                        access_id int8 NULL,
                                        tenant_id uuid NULL,
                                        capa_url varchar NULL,
                                        neighborhood varchar NULL,
                                        "number" varchar NULL,
                                        street varchar NULL,
                                        postcode varchar NULL,
                                        state_registration varchar NULL,
                                        municipal_registration varchar NULL,
                                        rntrc varchar NULL,
                                        complement varchar NULL,
                                        control_portal bool DEFAULT false NULL,
                                        contract_term_signed bool DEFAULT false NULL,
                                        automatic_email varchar NULL,
                                        CONSTRAINT "Organizations_pkey" PRIMARY KEY (id)
);