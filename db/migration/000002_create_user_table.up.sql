CREATE TABLE profiles (
   id BIGSERIAL PRIMARY KEY,
   name VARCHAR(255) NOT NULL
);


CREATE TABLE public.users (
                              id bigserial NOT NULL,
                              "name" varchar(255) NOT NULL,
                              email varchar(255) NOT NULL,
                              "password" varchar(255) NULL,
                              created_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
                              updated_at timestamp NULL,
                              profile_id int8 NULL,
                              "document" varchar(255) NULL,
                              state varchar(255) NULL,
                              city varchar(255) NULL,
                              neighborhood varchar(255) NULL,
                              street varchar(255) NULL,
                              street_number varchar(255) NULL,
                              phone varchar(255) NULL,
                              google_id varchar(255) NULL,
                              profile_picture varchar(255) NULL,
                              status bool NOT NULL,
                              CONSTRAINT users_pkey PRIMARY KEY (id)
);