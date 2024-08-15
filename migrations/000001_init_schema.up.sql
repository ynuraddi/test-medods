CREATE TABLE IF NOT EXISTS "users" (
    id SERIAL PRIMARY KEY,
    email VARCHAR UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS "sessions" (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    access_token_id VARCHAR NOT NULL,
    refresh_token_hash VARCHAR NOT NULL,
    ip VARCHAR NOT NULL,
    created_at BIGINT NOT NULL,
    version BIGINT NOT NULL DEFAULT 1,

    FOREIGN KEY(user_id) REFERENCES users(id)
);

INSERT INTO users(email) VALUES('mock@gmail.com');