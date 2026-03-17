-- +goose Up
ALTER TABLE services ADD COLUMN IF NOT EXISTS driving_style_hint VARCHAR(120) NOT NULL DEFAULT '';

UPDATE services SET driving_style_hint = 'Спокойный, городской' WHERE pad_type = 'Органические';
UPDATE services SET driving_style_hint = 'Спортивный, смешанный' WHERE pad_type = 'Керамические';
UPDATE services SET driving_style_hint = 'Агрессивный, активный' WHERE pad_type = 'Полуметаллические';

-- +goose Down
ALTER TABLE services DROP COLUMN IF EXISTS driving_style_hint;
