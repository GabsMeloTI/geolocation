CREATE TABLE announcement (
                              id BIGSERIAL PRIMARY KEY,
                              user_id BIGINT not null,
                              destination VARCHAR(255) NOT NULL,
                              origin VARCHAR(255) NOT NULL,
                              destination_lat DECIMAL(10,6) NOT NULL,
                              destination_lng DECIMAL(10,6) NOT NULL,
                              origin_lat DECIMAL(10,6) NOT NULL,
                              origin_lng DECIMAL(10,6) NOT NULL,
                              distance BIGINT NOT NULL,
                              pickup_date DATE NOT NULL,
                              delivery_date DATE NOT NULL,
                              expiration_date DATE NOT NULL,
                              title VARCHAR NOT NUL,
                              cargo_type VARCHAR(100) NOT NULL,
                              cargo_species VARCHAR NOT NULL,
                              cargo_volume VARCHAR NOT NULL,
                              cargo_weight DECIMAL(10,2) NOT NULL,
                              vehicles_accepted VARCHAR(100) NOT NULL,
                              trailer VARCHAR NOT NULL,
                              requires_tarp BOOLEAN NOT NULL,
                              tracking BOOLEAN NOT NULL,
                              agency BOOLEAN  NOT NULL,
                              description TEXT NOT NULL,
                              payment_type TEXT NOT NULL,
                              advance VARCHAR NOT NULL,
                              toll BOOLEAN NOT NULL,
                              situation VARCHAR NOT NULL,
                              status BOOL NOT NULL,
                              created_at TIMESTAMP NOT NULL,
                              created_who VARCHAR NOT NULL,
                              updated_at TIMESTAMP NULL,
                              updated_who VARCHAR NULL
);

ALTER TABLE announcement
    ADD CONSTRAINT "fk_user"
    FOREIGN KEY ("user_id")
    REFERENCES users ("id");

