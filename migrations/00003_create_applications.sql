-- +goose Up
CREATE TABLE IF NOT EXISTS applications (
  id             BIGSERIAL PRIMARY KEY,
  status         VARCHAR(16) NOT NULL,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_by_id  BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

  formed_at      TIMESTAMPTZ NULL,
  completed_at   TIMESTAMPTZ NULL,
  moderator_id   BIGINT NULL REFERENCES users(id) ON DELETE RESTRICT,

  driving_style  VARCHAR(20) NULL,
  mileage        INTEGER NULL,

  total_price    INTEGER NULL,

  CONSTRAINT applications_status_chk CHECK (status IN ('draft','deleted','formed','completed','rejected'))
);

-- не более одной заявки-черновика на пользователя
CREATE UNIQUE INDEX IF NOT EXISTS applications_one_draft_per_user
  ON applications (created_by_id)
  WHERE status = 'draft';

CREATE INDEX IF NOT EXISTS applications_status_idx ON applications USING btree (status);

-- +goose Down
DROP TABLE IF EXISTS applications;

