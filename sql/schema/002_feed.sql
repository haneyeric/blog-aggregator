-- +goose Up
CREATE TABLE feeds (
    id uuid PRIMARY KEY,
    created_at timestamp not null,
    updated_at timestamp not null,
    name text unique not null,
    url text unique not null,
    user_id uuid not null
);
ALTER TABLE IF EXISTS feeds
ADD CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- +goose Down
DROP TABLE IF EXISTS feeds;


