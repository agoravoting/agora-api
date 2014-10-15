package ballotbox

import (
	"github.com/agoravoting/agora-http-go/middleware"
	"github.com/agoravoting/agora-http-go/util"
	s "github.com/agoravoting/agora-http-go/server"
	"github.com/codegangsta/negroni"
	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"path"
	"os"
)

type BallotBox struct {
	router *httprouter.Router
	name   string

	insertStmt *sqlx.Stmt
	getStmt    *sqlx.Stmt
}

// TODO: move inside BallotBox
var configs = make(map[string]string)
var pubkeys = make(map[string]string)

func (bb *BallotBox) Name() string {
	return bb.name
}

func (bb *BallotBox) Init(cfg map[string]*json.RawMessage) (err error) {
	var ballotboxSessionExpire int
	json.Unmarshal(*cfg["ballotboxSessionExpire"], &ballotboxSessionExpire)

	// setup the routes
	bb.router = httprouter.New()
	bb.router.POST("/election/:election_id/vote/:voter_id", middleware.Join(
		s.Server.ErrorWrap.Do(bb.postVote),
		s.Server.CheckPerms("voter-${election_id}-${voter_id}", ballotboxSessionExpire)))
	bb.router.GET("/election/:election_id/check-hash/:voter_id/:vote_hash", middleware.Join(
		s.Server.ErrorWrap.Do(bb.checkHash),
		s.Server.CheckPerms("voter-${election_id}-${voter_id}", ballotboxSessionExpire)))
	bb.router.GET("/election/:election_id/config/:voter_id", middleware.Join(
		s.Server.ErrorWrap.Do(bb.getElectionConfig),
		s.Server.CheckPerms("voter-${election_id}-${voter_id}", ballotboxSessionExpire)))
	bb.router.GET("/election/:election_id/pubkeys", middleware.Join(
		s.Server.ErrorWrap.Do(bb.getElectionPubKeys)))

	// setup prepared sql queries
	if bb.insertStmt, err = s.Server.Db.Preparex("SELECT set_vote($1, $2, $3, $4, $5)"); err != nil {
		return
	}
	if bb.getStmt, err = s.Server.Db.Preparex("SELECT id, vote, vote_hash, election_id, voter_id FROM votes WHERE election_id = $1 and voter_id = $2 and vote_hash = $3"); err != nil {
		return
	}

	// initialize election cfgs to return in getConfig
	var electionDir string
	json.Unmarshal(*cfg["electionDir"], &electionDir)
	s.Server.Logger.Printf("Loading cfgs from %s", electionDir)

	files, err := ioutil.ReadDir(electionDir)
	if(err != nil) {
		return
	}

	for _, f := range files {
		if(f.IsDir()) {
			// read config.json
			cfgPath := path.Join(electionDir, f.Name(), "config.json")
			var cfgText string
			cfgText, err = util.Contents(cfgPath)
			if(err != nil) {
				s.Server.Logger.Printf("Could not read config.json at %s %v, skipping", cfgPath, err)
				continue
			} else {
				s.Server.Logger.Printf("Reading %s", cfgPath)
			}
			var cfg map[string]*json.RawMessage
			err = json.Unmarshal([]byte(cfgText), &cfg)
			if err != nil {
				s.Server.Logger.Printf("Error reading config file %s %v, skipping", cfgPath, err)
				continue
			}
			var electionId string
			value, ok := cfg["election-id"]
			if !ok {
				electionId = f.Name()
			} else {
				json.Unmarshal(*value, &electionId)
			}

			s.Server.Logger.Printf("Loaded config file for election %s", electionId)
			configs[electionId] = cfgText

			// read pk_<election-id>
			pkPath := path.Join(electionDir, f.Name(), "pk_" + electionId)
			var pkText string
			if _, err := os.Stat(pkPath); os.IsNotExist(err) {
				s.Server.Logger.Printf("No pubkey at %s", pkPath)
				continue
			}
			pkText, err = util.Contents(pkPath)
			if(err != nil) {
				s.Server.Logger.Printf("Could not read pubkey at %s %v, skipping", pkPath, err)
				continue
			} else {
				s.Server.Logger.Printf("Reading %s", pkPath)
			}
			var pk []interface{}
			err = json.Unmarshal([]byte(pkText), &pk)
			if err != nil {
				s.Server.Logger.Printf("Error reading pubkey file %s %v, skipping", pkPath, err)
				continue
			}
			pubkeys[electionId] = pkText
		}
	}

	// add the routes to the server
	handler := negroni.New(negroni.Wrap(bb.router))
	s.Server.Mux.OnMux("api/v1/ballotbox", handler)
	return
}

// returns the vote corresponding to the given hash
func (bb *BallotBox) checkHash(w http.ResponseWriter, r *http.Request, p httprouter.Params) *middleware.HandledError {
	var (
		v   []Vote
		err error
		voteHash string
	)

	electionId := p.ByName("election_id")
	voterId := p.ByName("voter_id")
	voteHash = p.ByName("vote_hash")
	if electionId == "" {
		return &middleware.HandledError{Err: err, Code: 400, Message: "No election_id", CodedMessage: "empty-election-id"}
	}
	if voterId == "" {
		return &middleware.HandledError{Err: err, Code: 400, Message: "No voter_id", CodedMessage: "empty-voter-id"}
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

func (bb *BallotBox) getElectionConfig(w http.ResponseWriter, r *http.Request, p httprouter.Params) *middleware.HandledError {
	var err error
	s.Server.Logger.Printf("getElectionConfig")

	electionId := p.ByName("election_id")
	voterId := p.ByName("voter_id")
	if electionId == "" {
		return &middleware.HandledError{Err: err, Code: 400, Message: "No election_id", CodedMessage: "empty-election-id"}
	}
	if voterId == "" {
		return &middleware.HandledError{Err: err, Code: 400, Message: "No voter_id", CodedMessage: "empty-voter-id"}
	}

	w.Header().Set("Content-Type", "application/json")
	config, ok := configs[electionId]
	if !ok {
		return &middleware.HandledError{Err: err, Code: 404, Message: "Not found", CodedMessage: "not-found"}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(config))
	return nil
}

func (bb *BallotBox) getElectionPubKeys(w http.ResponseWriter, r *http.Request, p httprouter.Params) *middleware.HandledError {
	var err error
	s.Server.Logger.Printf("getElectionConfig")

	electionId := p.ByName("election_id")
	if electionId == "" {
		return &middleware.HandledError{Err: err, Code: 400, Message: "No election_id", CodedMessage: "empty-election-id"}
	}

	w.Header().Set("Content-Type", "application/json")
	pubkeys, ok := pubkeys[electionId]
	if !ok {
		return &middleware.HandledError{Err: err, Code: 404, Message: "Not found", CodedMessage: "not-found"}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pubkeys))
	return nil
}

// add a new vote
func (bb *BallotBox) postVote(w http.ResponseWriter, r *http.Request, p httprouter.Params) *middleware.HandledError {
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
	// ip := r.RemoteAddr
	ip := r.Header.Get("X-Forwarded-For")
	if len(ip) == 0 {
		ip = r.RemoteAddr
	}

	if electionId == "" {
		return &middleware.HandledError{Err: err, Code: 400, Message: "No election_id", CodedMessage: "empty-election-id"}
	}
	if voterId == "" {
		return &middleware.HandledError{Err: err, Code: 400, Message: "No voter_id", CodedMessage: "empty-voter-id"}
	}
	vote_json["election_id"] = electionId
	vote_json["voter_id"] = voterId

	var foo string
	if err = bb.insertStmt.Get(&foo, vote_json["vote"], vote_json["vote_hash"], vote_json["election_id"], vote_json["voter_id"], ip); err != nil {
		tx.Rollback()
		return &middleware.HandledError{Err: err, Code: 500, Message: "Error calling set_vote", CodedMessage: "error-upsert"}
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
	// rb, err := httputil.DumpRequest(r, true)
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