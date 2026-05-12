-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
  IF to_regclass('public.brake_pad_wear') IS NULL
     AND to_regclass('public."brake-wear"') IS NULL
     AND to_regclass('public.brake_wear') IS NOT NULL
     AND to_regclass('public."brake-pad-wear"') IS NOT NULL THEN
    ALTER TABLE brake_wear RENAME TO brake_pad_wear_tmp;
    ALTER TABLE "brake-pad-wear" RENAME TO "brake-wear";
    ALTER TABLE brake_pad_wear_tmp RENAME TO "brake-pad-wear";
  END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
  IF to_regclass('public.brake_wear') IS NULL
     AND to_regclass('public."brake-wear"') IS NOT NULL
     AND to_regclass('public."brake-pad-wear"') IS NOT NULL THEN
    ALTER TABLE "brake-pad-wear" RENAME TO brake_wear_tmp;
    ALTER TABLE "brake-wear" RENAME TO "brake-pad-wear";
    ALTER TABLE brake_wear_tmp RENAME TO brake_wear;
  END IF;
END $$;
-- +goose StatementEnd
