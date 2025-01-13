-- +goose Up
CREATE TABLE feed_follows (
    id uuid PRIMARY KEY,
    created_at timestamp not null,
    updated_at timestamp not null,
    user_id uuid not null REFERENCES users(id) ON DELETE CASCADE,
    feed_id uuid not null REFERENCES feeds(id) ON DELETE CASCADE,
    UNIQUE(user_id, feed_id)
);

-- +goose Down
DROP TABLE IF EXISTS feed_follows;


