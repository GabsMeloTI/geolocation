create table offers
(
    id               bigserial
        primary key,
    advertisement_id bigint,
    price            double precision not null,
    interested_id    bigint
        references users,
    status           boolean
);
