package ballotbox

import (
	"encoding/json"
	"github.com/agoravoting/agora-http-go/util"
	"time"
	"math/big"
	"errors"
	"crypto/sha256"
	"io"
	"net/http"
	"fmt"
	// j "github.com/gima/jsonv/src"
	// "io/ioutil"
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

type EncryptedVote struct {
    A string `json:"a"`
    Choices []*Choice `json:"choices"`
    ElectionHash ElectionHash `json:"election_hash"`
    IssueDate time.Time `json:"issue_date"`
    Proofs []*Popk `json:"proofs"`
}

/* var	voteSchema = &j.Object{Properties:[]j.ObjectItem{
	{"vote", &j.String{}},
	{"vote_hash", &j.String{}},
}}*/

func (v *Vote) validate(electionPks []map[string]*big.Int) error {
	encryptedVote, err := ParseEncryptedVote([]byte(v.Vote))
    if err != nil {
		return err
    }
    if err := encryptedVote.validate(electionPks); err != nil {
    	return err
    }


    encryptedVoteString, err := json.Marshal(encryptedVote)
    if err != nil {
    	return err
    }

    v.Vote = string(encryptedVoteString)
	h256 := sha256.New()
   	io.WriteString(h256, v.Vote)
	_hashed := h256.Sum(nil)
	hashed := fmt.Sprintf("%x", _hashed)

	if hashed != v.VoteHash {
		return errors.New("Vote hash mismatch")
	}

    return nil
}

func (e *EncryptedVote) validate(electionPks []map[string]*big.Int) (err error) {
	for _, element := range e.Proofs {
		if err = element.validate(); err != nil {
			return err
		}
	}

	for index, element := range e.Choices {
		if err = element.validate(electionPks[index]); err != nil {
			return err
		}
	}
	if err = e.checkPopk(electionPks); err != nil {
		return err
	}

    return nil
}

func (e *EncryptedVote) checkPopk(electionPks []map[string]*big.Int) error {

	for index, proof := range e.Proofs {
    	choice := e.Choices[index]

    	h256 := sha256.New()
    	toHash := fmt.Sprintf("%s/%s", choice.Alpha.String(), proof.Commitment.String())
   		io.WriteString(h256, toHash)
   		_hashed := h256.Sum(nil)
   		hashed := fmt.Sprintf("%x", _hashed)

   		expected := big.NewInt(0)
        _, ok := expected.SetString(hashed, 16)
        if ! ok {
			return errors.New("Error calculating popk hash")
        }

        if proof.Challenge.Cmp(expected) != 0 {
			return errors.New("Popk hash mismatch")
        }

        pk := electionPks[index]

        first := big.NewInt(0)
        first.Exp(pk["g"], proof.Response, pk["p"])

        second := big.NewInt(0)
        second.Exp(choice.Alpha, proof.Challenge, pk["p"])
        second.Mul(second, proof.Commitment)
        second.Mod(second, pk["p"])

        if first.Cmp(second) != 0 {
			return errors.New("Failed verifying popk")
        }
    }

	return nil
}

type Popk struct {
    Challenge *big.Int `json:"-"`
    Commitment *big.Int `json:"-"`
    Response *big.Int `json:"-"`
    ChallengeString string `json:"challenge"`
    CommitmentString string `json:"commitment"`
    ResponseString string `json:"response"`
}

func (p *Popk) validate() error {
	p.Challenge = big.NewInt(0)
	_, ok := p.Challenge.SetString(p.ChallengeString, 10)
	if ! ok {
		return errors.New("Error parsing challenge")
	}

	p.Commitment = big.NewInt(0)
	_, ok = p.Commitment.SetString(p.CommitmentString, 10)
	if ! ok {
		return errors.New("Error parsing commitment")
	}

	p.Response = big.NewInt(0)
	_, ok = p.Response.SetString(p.ResponseString, 10)
	if ! ok {
		return errors.New("Error parsing response")
	}

	return nil
}

type Choice struct {
    Alpha *big.Int `json:"-"`
    Beta *big.Int `json:"-"`
    AlphaString string `json:"alpha"`
    BetaString string `json:"beta"`
}

func (c *Choice) validate(pk map[string]*big.Int) error {
	c.Alpha = big.NewInt(0)
	_, ok := c.Alpha.SetString(c.AlphaString, 10)
	if ! ok {
		return errors.New("Error parsing alpha")
	}
	residue := quadraticResidue(c.Alpha, pk["p"])
	if ! residue {
		return errors.New("Alpha quadratic non-residue")
	}

	c.Beta = big.NewInt(0)
	_, ok = c.Beta.SetString(c.BetaString, 10)
	if ! ok {
		return errors.New("Error parsing Beta")
	}
	residue = quadraticResidue(c.Beta, pk["p"])
	if ! residue {
		return errors.New("Beta quadratic non-residue")
	}

	return nil
}

type ElectionHash struct {
	A string `json:"a"`
	Value string `json:"value"`
}

func ParseVote(r *http.Request) (v Vote, err error) {
	if err != nil {
		return
	}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&v)

	if err != nil {
		return
	}
	/* var b []byte
	if b, err = ioutil.ReadAll(r.Body); err == nil {
    	var f interface{}
    	if err = json.Unmarshal(b, &f); err != nil {
    		return
    	}
    	var path string
    	if path, err = schema.Validate(&f); err == nil {
    		fmt.Printf(">>>>>>>>>> %s\n", b)
    		err = json.Unmarshal(b, &v)
    		fmt.Printf(">>>>>>>>>> %v\n", v)
		} else {
    		msg := fmt.Sprintf("Failed (%s). Path: %s", err, path)
    		err = errors.New(msg)
		}
	}*/
	return
}

func ParseEncryptedVote(data []byte) (v *EncryptedVote, err error) {
	v = &EncryptedVote{}
	err = json.Unmarshal(data, v)
	return
}

// re encode to map
func (v *Vote) Map() (ret map[string]interface{}, err error) {

	ret = map[string]interface{}{
		"id":               v.Id,
		"vote":        		v.Vote,
		"vote_hash":        v.VoteHash,
		"election_id":      v.ElectionId,
		"voter_id":       	v.VoterId,
	}
	return
}

func (v *Vote) Marshal() ([]byte, error) {
	j, err := v.Map()
	if err != nil {
		return []byte(""), err
	}
	return util.JsonSortedMarshal(j)
}

func quadraticResidue(value *big.Int, modulus *big.Int) bool {
	// clone values
	val := big.NewInt(0)
	val.SetBytes(value.Bytes())
	mod := big.NewInt(0)
	mod.SetBytes(modulus.Bytes())

	return legendre(val, mod) == 1
}

// http://programmingpraxis.com/2012/05/01/legendres-symbol/
func legendre(value *big.Int, modulus *big.Int) int64 {
	zero := big.NewInt(0)
	two := big.NewInt(2)
	three := big.NewInt(3)
	four := big.NewInt(4)
	five := big.NewInt(5)
	eight := big.NewInt(8)
	c1 := big.NewInt(0)

	a := big.NewInt(0)
	a.Mod(value, modulus)
	t := big.NewInt(1)

	for a.Cmp(zero) != 0 {
		for c1.Mod(a, two).Cmp(zero) == 0 {
			a.Div(a, two)

			if c1.Mod(modulus, eight).Cmp(three) == 0 || c1.Mod(modulus, eight).Cmp(five) == 0 {
				t.Neg(t)
			}
		}

		// swap
		tmp := big.NewInt(0)
		tmp.SetBytes(a.Bytes())
		a.SetBytes(modulus.Bytes())
		modulus.SetBytes(tmp.Bytes())

		if c1.Mod(a, four).Cmp(three) == 0 && c1.Mod(modulus, four).Cmp(three) == 0 {
			t.Neg(t)
		}
		a.Mod(a, modulus)
	}

	if modulus.Cmp(big.NewInt(1)) == 0 {
		return t.Int64()
	}
	return 0
}