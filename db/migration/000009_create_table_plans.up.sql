CREATE TABLE plans (
                        id BIGSERIAL PRIMARY KEY,
                        name VARCHAR(20) NOT NULL,
                        price VARCHAR(20) NOT NULL,
                        duration VARCHAR(20) NOT NULL,
                        annual BOOLEAN NOT NULL
);

CREATE TABLE user_plan (
                       id BIGSERIAL PRIMARY KEY,
                       id_user BIGINT references users(id),
                       id_plan BIGINT references plans(id),
                       active BOOLEAN NOT NULL,
                       active_date TIMESTAMP NOT NULL default now(),
                       expiration_date TIMESTAMP NOT NULL
);
