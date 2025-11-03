-- +goose Up

-- CockroachDB supports gen_random_uuid() without extensions.

CREATE TABLE IF NOT EXISTS users (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email         STRING NOT NULL UNIQUE,
  display_name  STRING NOT NULL,
  attrs         JSONB NOT NULL DEFAULT '{}'::JSONB,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS pets (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name          STRING NOT NULL,
  species       STRING NOT NULL,
  attrs         JSONB NOT NULL DEFAULT '{}'::JSONB,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_pets_user_id ON pets(user_id);

CREATE TABLE IF NOT EXISTS habits (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title         STRING NOT NULL,
  done          BOOLEAN NOT NULL DEFAULT false,
  icons         STRING NOT NULL DEFAULT '💡',
  cadence       STRING NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_habits_user_id ON habits(user_id);

CREATE TABLE IF NOT EXISTS games (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title         STRING NOT NULL,
  status        STRING NOT NULL DEFAULT 'backlog',
  attrs         JSONB NOT NULL DEFAULT '{}'::JSONB,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_games_user_id ON games(user_id);

-- +goose Down
DROP TABLE IF EXISTS games;
DROP TABLE IF EXISTS habits;
DROP TABLE IF EXISTS pets;
DROP TABLE IF EXISTS users;
