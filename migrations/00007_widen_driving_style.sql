-- +goose Up
ALTER TABLE applications
    ALTER COLUMN driving_style TYPE VARCHAR(64);

-- +goose Down
ALTER TABLE applications
    ALTER COLUMN driving_style TYPE VARCHAR(20);

