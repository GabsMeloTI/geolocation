CREATE TABLE profiles (
                          id BIGSERIAL PRIMARY KEY,
                          name VARCHAR(255) NOT NULL
);


create table users
(
    id            bigserial primary key,
    name          varchar(255) not null,
    email         varchar(255) not null,
    password      varchar(255),
    created_at    timestamp default CURRENT_TIMESTAMP,
    updated_at    timestamp,
    profile_id    bigint references profiles,
    document      varchar(255),
    state         varchar(255),
    city          varchar(255),
    neighborhood  varchar(255),
    street        varchar(255),
    street_number varchar(255),
    phone         varchar(255),
    google_id     varchar(255),
    profile_picture varchar(255)
);