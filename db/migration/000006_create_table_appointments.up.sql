CREATE TABLE appointments (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT not null,
    truck_id BIGINT not null,
    advertisement_id BIGINT not null,
    situation VARCHAR NOT NULL,
    status BOOL NOT NULL,
    created_who VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_who VARCHAR NULL,
    updated_at TIMESTAMP NULL
);