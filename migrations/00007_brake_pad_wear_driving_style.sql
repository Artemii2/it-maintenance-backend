-- +goose Up
ALTER TABLE "brake-wear"
    ALTER COLUMN driving_style TYPE VARCHAR(64);

-- +goose Down
ALTER TABLE "brake-wear"
    ALTER COLUMN driving_style TYPE VARCHAR(20);
