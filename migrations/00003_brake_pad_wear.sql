-- +goose Up
CREATE TABLE IF NOT EXISTS "brake-pad-wear" (
  id             BIGSERIAL PRIMARY KEY,
  status         VARCHAR(16) NOT NULL,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_by_id  BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

  formed_at      TIMESTAMPTZ NULL,
  completed_at   TIMESTAMPTZ NULL,
  moderator_id   BIGINT NULL REFERENCES users(id) ON DELETE RESTRICT,

  driving_style  VARCHAR(20) NULL,
  mileage        INTEGER NULL,

  CONSTRAINT brake_pad_wear_status_chk CHECK (status IN ('draft','deleted','formed','completed','rejected'))
);

CREATE UNIQUE INDEX IF NOT EXISTS brake_pad_wear_one_draft_per_user
  ON "brake-pad-wear" (created_by_id)
  WHERE status = 'draft';

CREATE INDEX IF NOT EXISTS brake_pad_wear_status_idx ON "brake-pad-wear" USING btree (status);

-- +goose Down
DROP TABLE IF EXISTS "brake-pad-wear";
