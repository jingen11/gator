-- +goose Up
ALTER TABLE feeds ADD last_fetched_at TIMESTAMP;

-- +goose Down
DROP TABLE feeds;
