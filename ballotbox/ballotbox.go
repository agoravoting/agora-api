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
	"math/big"
	"fmt"
)

type BallotBox struct {
	router *httprouter.Router
	name   string

	insertStmt *sqlx.Stmt
	getStmt    *sqlx.Stmt
	maxWrites  int

	configs map[string]string
	pubkeys map[string]string
	pubkeyObjects map[string][]map[string]*big.Int
	checkResidues bool
}

func (bb *BallotBox) Name() string {
	return bb.name
}

func (bb *BallotBox) Init(cfg map[string]*json.RawMessage) (err error) {
	bb.configs = make(map[string]string)
	bb.pubkeys = make(map[string]string)
	bb.pubkeyObjects = make(map[string][]map[string]*big.Int)

	var ballotboxSessionExpire int
	json.Unmarshal(*cfg["ballotboxSessionExpire"], &ballotboxSessionExpire)
	var maxWrites int
	json.Unmarshal(*cfg["maxWrites"], &maxWrites)
	bb.maxWrites = maxWrites

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
	if bb.insertStmt, err = s.Server.Db.Preparex("SELECT set_vote($1, $2, $3, $4, $5, $6)"); err != nil {
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
			bb.configs[electionId] = cfgText

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
			bb.pubkeys[electionId] = pkText

			var pksDecoded []interface{}
			if err := json.Unmarshal([]byte(pkText), &pksDecoded); err != nil {
    			s.Server.Logger.Printf("Error reading pubkey file %s %v, skipping", pkPath, err)
				continue
    		}

    		keys := make([]map[string]*big.Int, len(pksDecoded))

    		for index,element := range pksDecoded {
    			next := element.(map[string]interface{})
    			_modulus := next["p"].(string)
    			_generator := next["g"].(string)

    			modulus := big.NewInt(0)
        		_, ok := modulus.SetString(_modulus, 10)
        		if ! ok {
        			s.Server.Logger.Printf("Error reading pubkey(p) file %s, skipping", pkPath)
					continue
        		}
        		generator := big.NewInt(0)
        		_, ok = generator.SetString(_generator, 10)
        		if ! ok {
        			s.Server.Logger.Printf("Error reading pubkey(g) file %s, skipping", pkPath)
					continue
        		}
        		key := make(map[string]*big.Int)
        		key["p"] = modulus
        		key["g"] = generator

        		keys[index] = key
    		}
    		bb.pubkeyObjects[electionId] = keys
		}
	}

	json.Unmarshal(*cfg["checkResidues"], &bb.checkResidues)

	// add the routes to the server
	handler := negroni.New(negroni.Wrap(bb.router))
	s.Server.Mux.OnMux("api/v1/ballotbox", handler)
	return
}

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
	config, ok := bb.configs[electionId]
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
	pubkeys, ok := bb.pubkeys[electionId]
	if !ok {
		return &middleware.HandledError{Err: err, Code: 404, Message: "Not found", CodedMessage: "not-found"}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pubkeys))
	return nil
}

func (bb *BallotBox) postVote(w http.ResponseWriter, r *http.Request, p httprouter.Params) *middleware.HandledError {
	var (
		tx    = s.Server.Db.MustBegin()
		vote  Vote
		err   error
	)
	vote, err = ParseVote(r)
	if err != nil {
		return &middleware.HandledError{Err: err, Code: 400, Message: "Invalid json-encoded vote", CodedMessage: "invalid-json"}
	}

	electionId := p.ByName("election_id")
	voterId := p.ByName("voter_id")
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
	pks, ok := bb.pubkeyObjects[electionId]
    if ! ok {
    	return &middleware.HandledError{Err: err, Code: 400, Message: "Pks not found for election", CodedMessage: "vote-pks-not-found"}
    }
    if err := vote.validate(pks, bb.checkResidues); err != nil {
    	return &middleware.HandledError{Err: err, Code: 400, Message: "Vote validation failed", CodedMessage: "vote-validation-failedj"}
    }

	encryptedVoteString := vote.Vote

	var updated string
	if err = bb.insertStmt.Get(&updated, encryptedVoteString, vote.VoteHash, electionId, voterId, ip, bb.maxWrites); err != nil {
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
	var json = fmt.Sprintf("{\"updated\": \"%s\"}", updated)
	w.Write([]byte(json))

	return nil
}

func init() {
	s.Server.AvailableModules = append(s.Server.AvailableModules, &BallotBox{name: "github.com/agoravoting/agora-api/ballotbox"})
}