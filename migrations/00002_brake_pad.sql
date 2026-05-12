-- +goose Up
CREATE TABLE IF NOT EXISTS "brake-pad" (
  id            BIGSERIAL PRIMARY KEY,
  title         VARCHAR(120) NOT NULL,
  description   TEXT NOT NULL,
  status        VARCHAR(16) NOT NULL,
  image_url     TEXT NULL,
  video_url     TEXT NULL,
  pad_type      VARCHAR(40) NOT NULL,
  base_resource INTEGER NOT NULL,
  driving_style_hint VARCHAR(120) NULL,
  CONSTRAINT brake_pad_status_chk CHECK (status IN ('active','deleted'))
);

CREATE INDEX IF NOT EXISTS brake_pad_title_idx ON "brake-pad" USING btree (title);

-- +goose Down
DROP TABLE IF EXISTS "brake-pad";
