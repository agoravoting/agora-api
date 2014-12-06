
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE votes ADD CONSTRAINT uniquevotehash UNIQUE (vote_hash);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE votes DROP CONSTRAINT uniquevotehash
