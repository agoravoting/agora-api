-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE votes (
	id serial PRIMARY KEY,
	vote text NOT NULL,
  vote_hash varchar(1024) NOT NULL,
	election_id varchar(1024) NOT NULL,
  voter_id varchar(1024) NOT NULL,
  created timestamp DEFAULT current_timestamp,
  modified timestamp DEFAULT current_timestamp
);
CREATE INDEX vote_hash_idx ON votes(vote_hash);
CREATE INDEX voter_id_idx ON votes(voter_id);
CREATE UNIQUE INDEX voter_id_election_id ON votes(voter_id, election_id);

CREATE OR REPLACE FUNCTION update_modified_column()
  RETURNS TRIGGER AS $$ BEGIN NEW.modified = now(); RETURN NEW; END; $$
  language 'plpgsql';

CREATE TRIGGER update_modtime BEFORE UPDATE
  ON votes FOR EACH ROW EXECUTE PROCEDURE
  update_modified_column();
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TRIGGER update_modtime on votes;
DROP TABLE votes;
DROP FUNCTION update_modified_column();
