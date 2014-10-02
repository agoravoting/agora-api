-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE votes (
	id serial PRIMARY KEY,
	vote text NOT NULL,
  vote_hash varchar(1024) NOT NULL,
	election_id varchar(1024) NOT NULL,
  voter_id varchar(1024) NOT NULL,
  ip varchar(64) NOT NULL,
  created timestamp DEFAULT current_timestamp,
  modified timestamp DEFAULT current_timestamp,
  write_count int DEFAULT 1
);
CREATE UNIQUE INDEX voter_id_election_id ON votes(voter_id, election_id);

-- adapted from http://www.postgresql.org/docs/current/static/plpgsql-control-structures.html#PLPGSQL-UPSERT-EXAMPLE
-- previous implementation, updates first
-- RETURNS BOOLEAN AS $$ BEGIN LOOP UPDATE votes SET vote = vote, modified = current_timestamp WHERE voter_id = vid and election_id = eid; IF found THEN RETURN found; END IF; BEGIN INSERT INTO votes(vote, vote_hash, election_id, voter_id) VALUES (v, vh, eid, vid); RETURN found; EXCEPTION WHEN unique_violation THEN END; END LOOP; END; $$
-- function on one line as goose does not seem to work otherwise
CREATE FUNCTION set_vote(v TEXT, vh TEXT, eid TEXT, vid TEXT, theip TEXT)
RETURNS VOID AS $$ BEGIN BEGIN INSERT INTO votes(vote, vote_hash, election_id, voter_id, ip) VALUES (v, vh, eid, vid, theip); RETURN; EXCEPTION WHEN unique_violation THEN UPDATE votes SET vote = vote, vote_hash = vh, ip = theip, modified = current_timestamp, write_count = write_count + 1 WHERE voter_id = vid and election_id = eid; IF NOT found THEN RAISE EXCEPTION 'Could not update on election_id --> %, voter_id ---> %', eid, vid; END IF; RETURN; END; END; $$
LANGUAGE plpgsql;
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE votes;
DROP FUNCTION set_vote(vote TEXT, vote_hash TEXT, election_id TEXT, voter_id TEXT, theip TEXT);