-- +goose Up
CREATE TABLE IF NOT EXISTS services (
  id            BIGSERIAL PRIMARY KEY,
  title         VARCHAR(120) NOT NULL,
  description   TEXT NOT NULL,
  status        VARCHAR(16) NOT NULL,
  image_url     TEXT NULL,
  video_url     TEXT NULL,
  pad_type      VARCHAR(40) NOT NULL,
  base_resource INTEGER NOT NULL,
  price         INTEGER NOT NULL,
  CONSTRAINT services_status_chk CHECK (status IN ('active','deleted'))
);

CREATE INDEX IF NOT EXISTS services_title_idx ON services USING btree (title);

-- +goose Down
DROP TABLE IF EXISTS services;

