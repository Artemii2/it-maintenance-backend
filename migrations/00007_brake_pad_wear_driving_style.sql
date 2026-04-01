-- +goose Up
ALTER TABLE "brake-pad-wear"
    ALTER COLUMN driving_style TYPE VARCHAR(64);

-- +goose Down
ALTER TABLE "brake-pad-wear"
    ALTER COLUMN driving_style TYPE VARCHAR(20);
