-- +goose Up
INSERT INTO users (id, username, password_hash, is_moderator)
VALUES
  (1, 'demo', 'demo', FALSE),
  (2, 'moderator', 'demo', TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO "brake-pad" (title, description, status, image_url, video_url, pad_type, base_resource, driving_style_hint)
VALUES
  (
    'Керамические тормозные колодки',
    'Керамические колодки обеспечивают стабильное торможение, низкий уровень шума и минимальное пылеобразование. Отличаются увеличенным сроком службы.',
    'active',
    'ceramic.jpg',
    'ceramic.MP4',
    'Керамические',
    60000,
    'спокойная'
  ),
  (
    'Органические тормозные колодки',
    'Органические колодки отличаются мягкой работой и доступной стоимостью. Подходят для спокойного городского режима.',
    'active',
    'organic.jpg',
    'organic.MP4',
    'Органические',
    35000,
    'спокойная'
  ),
  (
    'Полуметаллические тормозные колодки',
    'Полуметаллические колодки обеспечивают эффективное торможение при высоких нагрузках и подходят для активной езды.',
    'active',
    'semi_metallic.jpg',
    'semi_metallic.MP4',
    'Полуметаллические',
    45000,
    'активная'
  )
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM brake_wear;
DELETE FROM "brake-pad-wear";
DELETE FROM "brake-pad";
DELETE FROM users;
