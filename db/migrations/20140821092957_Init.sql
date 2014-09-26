-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE votes (
	id serial PRIMARY KEY,
	vote text NOT NULL,
  vote_hash varchar(1024) NOT NULL,
	election_id varchar(1024) NOT NULL,
  voter_id varchar(1024) NOT NULL
);
CREATE INDEX vote_hash_idx ON votes(vote_hash);
CREATE INDEX voter_id_idx ON votes(voter_id);
CREATE UNIQUE INDEX voter_id_election_id ON votes(voter_id, election_id);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE votes;