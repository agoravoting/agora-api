package ballotbox

import (
	"fmt"
	"github.com/agoravoting/agora-http-go/middleware"
	stest "github.com/agoravoting/agora-http-go/server/testing"
	"net/http"
	"testing"
)

const (
	newVote = `{
	"vote": "yeeehaaa",
	"vote_hash": "wowowowowo"
}`
	secret = "somesecret"
)

var (
	SharedSecret = "somesecret"
	Config       = `{
	"Debug": true,
	"DbMaxIddleConnections": 5,
	"DbConnectString": "user=ballotbox password=ballotbox dbname=ballotbox_test sslmode=disable",

	"SharedSecret": "somesecret",
	"Admins": ["test@example.com"],
	"ActiveModules": [
		"github.com/agoravoting/agora-api/ballotbox"
	],
	"RavenDSN": ""
}`
)

func TestEventApi(t *testing.T) {
	ts := stest.New(t, Config)
	defer ts.TearDown()
	voteAuth := map[string]string{"Authorization": middleware.AuthHeader("voter-1-1", SharedSecret)}

	newVote := ts.RequestJson("POST", "/api/v1/ballotbox/1/1", http.StatusAccepted, voteAuth, newVote).(map[string]interface{})
	fmt.Printf("new vote %v\n", newVote)

	foundVote := ts.Request("GET", "/api/v1/ballotbox/1/1/wowowowowo", http.StatusOK, voteAuth, "")
	fmt.Printf("found vote %v\n", foundVote)
}