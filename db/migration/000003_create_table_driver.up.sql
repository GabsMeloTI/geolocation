CREATE TABLE driver (
                        id BIGSERIAL PRIMARY KEY,
                        user_id BIGINT not null,
                        birth_date DATE NOT NULL,
                        cpf VARCHAR(14) NOT NULL,
                        license_number VARCHAR(20) NOT NULL,
                        license_category VARCHAR(5) NOT NULL,
                        license_expiration_date DATE NOT NULL,
                        state VARCHAR(50),
                        city VARCHAR(50),
                        neighborhood VARCHAR(50),
                        street VARCHAR(100),
                        street_number VARCHAR(10),
                        phone VARCHAR(20) NOT NULL,
                        status BOOL not null,
                        created_at timestamp not null,
                        updated_at timestamp null
);

ALTER TABLE driver
    ADD CONSTRAINT "fk_user"
    FOREIGN KEY ("user_id")
    REFERENCES users ("id");

