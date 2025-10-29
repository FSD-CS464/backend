-- +goose Up
CREATE UNIQUE INDEX IF NOT EXISTS uid_pets_user_name    ON pets(user_id, name);
CREATE UNIQUE INDEX IF NOT EXISTS uid_habits_user_title ON habits(user_id, title);
CREATE UNIQUE INDEX IF NOT EXISTS uid_games_user_title  ON games(user_id, title);

-- +goose Down
DROP INDEX IF EXISTS uid_games_user_title;
DROP INDEX IF EXISTS uid_habits_user_title;
DROP INDEX IF EXISTS uid_pets_user_name;
