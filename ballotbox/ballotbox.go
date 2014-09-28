package ballotbox

import (
	"github.com/agoravoting/agora-http-go/middleware"
	s "github.com/agoravoting/agora-http-go/server"
	"github.com/codegangsta/negroni"
	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"
	"encoding/json"
	"net/http"
)

const (
	SESSION_EXPIRE = 3600
)

type BallotBox struct {
	router *httprouter.Router
	name   string

	insertStmt *sqlx.Stmt
	getStmt    *sqlx.Stmt
}

func (bb *BallotBox) Name() string {
	return bb.name
}

func (bb *BallotBox) Init() (err error) {
	// setup the routes
	bb.router = httprouter.New()
	bb.router.POST("/:election_id/:voter_id", middleware.Join(
		s.Server.ErrorWrap.Do(bb.post),
		s.Server.CheckPerms("voter-${election_id}-${voter_id}", SESSION_EXPIRE)))
	bb.router.GET("/:election_id/:voter_id/:vote_hash", middleware.Join(
		s.Server.ErrorWrap.Do(bb.get),
		s.Server.CheckPerms("voter-${election_id}-${voter_id}", SESSION_EXPIRE)))

	// setup prepared sql queries
	if bb.insertStmt, err = s.Server.Db.Preparex("SELECT set_vote($1, $2, $3, $4)"); err != nil {
		return
	}
	if bb.getStmt, err = s.Server.Db.Preparex("SELECT id, vote, vote_hash, election_id, voter_id FROM votes WHERE election_id = $1 and voter_id = $2 and vote_hash = $3"); err != nil {
		return
	}

	// add the routes to the server
	handler := negroni.New(negroni.Wrap(bb.router))
	s.Server.Mux.OnMux("api/v1/ballotbox", handler)
	return
}

// returns the vote corresponding to the given hash
func (bb *BallotBox) get(w http.ResponseWriter, r *http.Request, p httprouter.Params) *middleware.HandledError {
	var (
		v   []Vote
		err error
		voteHash string
	)

	electionId := p.ByName("election_id")
	voterId := p.ByName("voter_id")
	voteHash = p.ByName("vote_hash")
	if electionId == "" {
		return &middleware.HandledError{Err: err, Code: 400, Message: "No election_id", CodedMessage: "error-insert"}
	}
	if voterId == "" {
		return &middleware.HandledError{Err: err, Code: 400, Message: "No voter_id", CodedMessage: "error-insert"}
	}
	if voteHash == "" {
		return &middleware.HandledError{Err: err, Code: 400, Message: "Invalid hash format", CodedMessage: "invalid-format"}
	}

	if err = bb.getStmt.Select(&v, electionId, voterId, voteHash); err != nil {
		return &middleware.HandledError{Err: err, Code: 500, Message: "Database error", CodedMessage: "error-select"}
	}

	if len(v) == 0 {
		return &middleware.HandledError{Err: err, Code: 404, Message: "Not found", CodedMessage: "not-found"}
	}

	b, err := v[0].Marshal()
	if err != nil {
		return &middleware.HandledError{Err: err, Code: 500, Message: "Error marshalling the data", CodedMessage: "marshall-error"}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
	return nil
}

// add a new vote
func (bb *BallotBox) post(w http.ResponseWriter, r *http.Request, p httprouter.Params) *middleware.HandledError {
	var (
		tx    = s.Server.Db.MustBegin()
		vote  Vote
		err   error
	)
	vote, err = parseVote(r)
	if err != nil {
		return &middleware.HandledError{Err: err, Code: 400, Message: "Invalid json-encoded vote", CodedMessage: "invalid-json"}
	}
	vote_json, err := vote.Json()
	if err != nil {
		return &middleware.HandledError{Err: err, Code: 500, Message: "Error re-writing the data to json", CodedMessage: "error-json-encode"}
	}

	electionId := p.ByName("election_id")
	voterId := p.ByName("voter_id")
	if electionId == "" {
		return &middleware.HandledError{Err: err, Code: 400, Message: "No election_id", CodedMessage: "error-insert"}
	}
	if voterId == "" {
		return &middleware.HandledError{Err: err, Code: 400, Message: "No voter_id", CodedMessage: "error-insert"}
	}
	vote_json["election_id"] = electionId
	vote_json["voter_id"] = voterId

	var foo string
	if err = bb.insertStmt.Get(&foo, vote_json["vote"], vote_json["vote_hash"], vote_json["election_id"], vote_json["voter_id"]); err != nil {
		tx.Rollback()
		return &middleware.HandledError{Err: err, Code: 500, Message: "Error calling merge_vote", CodedMessage: "error-upsert"}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return &middleware.HandledError{Err: err, Code: 500, Message: "Error comitting the vote", CodedMessage: "error-commit"}
	}

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))

	return nil
}

// parses a vote from a request.
func parseVote(r *http.Request) (v Vote, err error) {
	// 	rb, err := httputil.DumpRequest(r, true)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&v)
	if err != nil {
		return
	}
	return
}

// add the modules to available modules on startup
func init() {
	s.Server.AvailableModules = append(s.Server.AvailableModules, &BallotBox{name: "github.com/agoravoting/agora-api/ballotbox"})
}