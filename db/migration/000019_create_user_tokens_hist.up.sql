CREATE TABLE user_tokens_hist (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    token TEXT NOT NULL,
    payload JSONB NOT NULL,
    expired_at TIMESTAMP NOT NULL,
    origin VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT now() NOT NULL
);
