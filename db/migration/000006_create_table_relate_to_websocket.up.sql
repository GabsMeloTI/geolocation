CREATE TABLE "chat_rooms" (
                              "id" BIGSERIAL PRIMARY KEY,
                              "advertisement_id"bigint references advertisement(id) not null,
                              "advertisement_user_id" bigint references users(id) not null,
                              "interested_user_id" bigint references users(id) not null,
                              "status" bool NOT NULL,
                              "created_at" timestamp NOT NULL DEFAULT (now()),
                              "updated_at" timestamp
);


CREATE TABLE "chat_messages" (
                                 "id" BIGSERIAL PRIMARY KEY,
                                 "room_id" bigint references chat_rooms(id),
                                 "user_id" bigint references users(id),
                                 "content" text NOT NULL,
                                 "status" bool NOT NULL,
                                 "reply_id" bigint,
                                 "read_at" timestamp,
                                 "is_read"bool,
                                 "created_at" timestamp NOT NULL DEFAULT (now()),
                                 "updated_at" timestamp

);