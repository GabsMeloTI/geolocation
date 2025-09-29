CREATE TABLE "chat_rooms" (
                              "id" BIGSERIAL PRIMARY KEY,
                              "advertisement_id"bigint references advertisement(id) not null,
                              "advertisement_user_id" bigint references users(id) not null,
                              "interested_user_id" bigint references users(id) not null,
                              "status" bool NOT NULL,
                              "created_at" timestamp NOT NULL DEFAULT (now()),
                              "updated_at" timestamp
);


create table chat_messages
(
    id           bigserial
        primary key,
    room_id      bigint
        references chat_rooms,
    user_id      bigint
        references users,
    content      text                    not null,
    status       boolean                 not null,
    reply_id     bigint,
    read_at      timestamp,
    is_read      boolean,
    created_at   timestamp default now() not null,
    updated_at   timestamp,
    type_message varchar,
    is_accepted  boolean
);