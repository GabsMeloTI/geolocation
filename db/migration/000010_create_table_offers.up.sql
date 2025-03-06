CREATE TABLE offers (
    id BIGSERIAl PRIMARY KEY,
    advertisement_id BIGINT references advertisement(id),
    price double precision NOT NULL,
    interested_id bigint references users(id)
);