-- +goose Up
-- +goose StatementBegin
CREATE TABLE users
(
    id            UUID PRIMARY KEY,
    name          VARCHAR(100)        NOT NULL,
    email         VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255)        NOT NULL,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX index_email ON users (email);

CREATE TABLE homes
(
    id               UUID PRIMARY KEY,
    owner_id         UUID         NOT NULL,
    name             VARCHAR(100) NOT NULL,
    street_address_1 VARCHAR(255) NOT NULL,
    street_address_2 VARCHAR(255),
    city             VARCHAR(100) NOT NULL,
    state            VARCHAR(100) NOT NULL,
    zip_code         VARCHAR(20)  NOT NULL,
    country          VARCHAR(100) NOT NULL,
    description      TEXT,
    tags             TEXT[],
    image            VARCHAR(255),
    created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES users (id)
);
CREATE INDEX index_owner_id ON homes (owner_id);
CREATE INDEX index_state ON homes (state);
CREATE INDEX index_city ON homes (city);
CREATE INDEX index_tags ON homes USING GIN (tags);

CREATE TABLE sessions
(
    session_token TEXT PRIMARY KEY,
    user_id       UUID        not null,
    expires_at    TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE admins
(
    id         UUID PRIMARY KEY,
    name       VARCHAR(100)        NOT NULL,
    email      VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE nyum_images
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID         NOT NULL,
    image_key  VARCHAR(500) NOT NULL,
    home_id    UUID,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (home_id) REFERENCES homes (id) ON DELETE CASCADE
);
CREATE INDEX index_nyum_images_user_id ON nyum_images (user_id);
CREATE INDEX index_nyum_images_home_id ON nyum_images (home_id);
CREATE INDEX index_nyum_images_key ON nyum_images (image_key);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE nyum_images ;
DROP TABLE homes;
DROP TABLE IF EXISTS sessions;
DROP TABLE users;
DROP TABLE admins;
-- +goose StatementEnd