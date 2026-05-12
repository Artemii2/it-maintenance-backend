-- +goose Up
CREATE TABLE IF NOT EXISTS "brake-pad-wear" (
  application_id BIGINT NOT NULL REFERENCES "brake-wear"(id) ON DELETE RESTRICT,
  service_id     BIGINT NOT NULL REFERENCES "brake-pad"(id) ON DELETE RESTRICT,

  quantity       INTEGER NOT NULL DEFAULT 1,
  position       INTEGER NOT NULL DEFAULT 1,
  is_primary     BOOLEAN NOT NULL DEFAULT FALSE,

  remaining_km      DOUBLE PRECISION NULL,
  remaining_percent INTEGER NULL,

  PRIMARY KEY (application_id, service_id),
  CONSTRAINT brake_pad_wear_qty_chk CHECK (quantity > 0)
);

CREATE INDEX IF NOT EXISTS brake_pad_wear_app_idx ON "brake-pad-wear" (application_id);

-- +goose Down
DROP TABLE IF EXISTS "brake-pad-wear";
