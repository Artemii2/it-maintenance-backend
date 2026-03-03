-- +goose Up
CREATE TABLE IF NOT EXISTS application_services (
  application_id BIGINT NOT NULL REFERENCES applications(id) ON DELETE RESTRICT,
  service_id     BIGINT NOT NULL REFERENCES services(id) ON DELETE RESTRICT,

  quantity       INTEGER NOT NULL DEFAULT 1,
  position       INTEGER NOT NULL DEFAULT 1,
  is_primary     BOOLEAN NOT NULL DEFAULT FALSE,
  comment        TEXT NULL,

  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  PRIMARY KEY (application_id, service_id),
  CONSTRAINT application_services_qty_chk CHECK (quantity > 0)
);

CREATE INDEX IF NOT EXISTS application_services_app_idx ON application_services (application_id);

-- +goose Down
DROP TABLE IF EXISTS application_services;

