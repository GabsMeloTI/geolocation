CREATE TABLE user_request_hist (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NULL REFERENCES users(id),
    token TEXT NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    status_code INT NOT NULL,
    execution_time_ms INT NOT NULL,
    created_at TIMESTAMP DEFAULT now() NOT NULL
);
