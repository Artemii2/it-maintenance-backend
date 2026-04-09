-- +goose Up
INSERT INTO users (id, username, password_hash, is_moderator)
VALUES
  (1, 'demo', 'demo', FALSE),
  (2, 'moderator', 'demo', TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO services (title, description, status, image_url, video_url, pad_type, base_resource, price)
VALUES
  (
    'Керамические тормозные колодки',
    'Керамические колодки обеспечивают стабильное торможение, низкий уровень шума и минимальное пылеобразование. Отличаются увеличенным сроком службы.',
    'active',
    'ceramic.jpg',
    'ceramic.MP4',
    'Керамические',
    60000,
    8500
  ),
  (
    'Органические тормозные колодки',
    'Органические колодки отличаются мягкой работой и доступной стоимостью. Подходят для спокойного городского режима.',
    'active',
    'organic.jpg',
    'organic.MP4',
    'Органические',
    35000,
    4500
  ),
  (
    'Полуметаллические тормозные колодки',
    'Полуметаллические колодки обеспечивают эффективное торможение при высоких нагрузках и подходят для активной езды.',
    'active',
    'semi_metallic.jpg',
    'semi_metallic.MP4',
    'Полуметаллические',
    45000,
    6500
  )
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM application_services;
DELETE FROM applications;
DELETE FROM services;
DELETE FROM users;

