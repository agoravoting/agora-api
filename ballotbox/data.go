package ballotbox

import (
	"encoding/json"
	"github.com/agoravoting/agora-http-go/util"
	"time"
)

type Vote struct {
	Id   			int64		`json:"-"`
	Vote            string		`json:"vote" db:"vote"`
	VoteHash        string  	`json:"vote_hash" db:"vote_hash"`
	ElectionId      string  	`json:"-" db:"election_id"`
	VoterId       	string  	`json:"-" db:"voter_id"`
	Ip       	    string  	`json:"-" db:"ip"`
	Created       	time.Time  	`json:"-" db:"created"`
	Modified	    time.Time  	`json:"-" db:"modified"`
	WriteCount	    int64  	    `json:"-" db:"write_count"`
}

func ParseEvent(data []byte) (v *Vote, err error) {
	v = &Vote{}
	err = json.Unmarshal(data, v)
	return
}

func (v *Vote) Marshal() ([]byte, error) {
	j, err := v.Json()
	if err != nil {
		return []byte(""), err
	}
	return util.JsonSortedMarshal(j)
}

func (v *Vote) Json() (ret map[string]interface{}, err error) {

	ret = map[string]interface{}{
		"id":               v.Id,
		"vote":        		v.Vote,
		"vote_hash":        v.VoteHash,
		"election_id":      v.ElectionId,
		"voter_id":       	v.VoterId,
	}
	return
}
