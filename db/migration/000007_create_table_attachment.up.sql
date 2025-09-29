CREATE TABLE public.attachments (
                                      id bigserial NOT NULL,
                                      user_id int8 NOT NULL,
                                      description varchar NULL,
                                      url varchar NOT NULL,
                                      name_file varchar NULL,
                                      size_file int8 NULL,
                                      type VARCHAR(50) NOT NULL,
                                      status bool NOT NULL,
                                      created_at timestamp DEFAULT now() NOT NULL,
                                      updated_at timestamp NULL,
                                      CONSTRAINT "attachments_pkey" PRIMARY KEY (id),
                                      CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id)
);