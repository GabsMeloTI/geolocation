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

