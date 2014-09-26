package ballotbox

import (
	"fmt"
	"github.com/agoravoting/agora-http-go/middleware"
	"github.com/agoravoting/agora-http-go/util"
	stest "github.com/agoravoting/agora-http-go/server/testing"
	"net/http"
	"testing"
	"bytes"
	"encoding/json"
	"time"
)

const (
	newVote = `{
	"vote": "0000000000000000000000000000000000000000000000000000000000000000000000000",
	"vote_hash": "hash"
}`
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

func TestAgoraApi(t *testing.T) {
	ts := stest.New(t, Config)
	defer ts.TearDown()
	voteAuth := map[string]string{"Authorization": middleware.AuthHeader("voter-1-1", SharedSecret)}

	posted := ts.RequestJson("POST", "/api/v1/ballotbox/1/1", http.StatusAccepted, voteAuth, newVote)
	fmt.Printf("new vote %v\n", posted)

	foundVote := ts.Request("GET", "/api/v1/ballotbox/1/1/hash", http.StatusOK, voteAuth, "")
	fmt.Printf("found vote %v\n", foundVote)
}

func request(method string, path string, headers map[string]string, requesTBody string, b *testing.B) *http.Response {
    client := &http.Client{}
    r, err := http.NewRequest(method, path, bytes.NewBufferString(requesTBody))
	if err != nil {
		b.Errorf("error creating request %v", err)
    }

    for key, value := range headers {
        r.Header.Set(key, value)
    }
    resp, err := client.Do(r)
    if err != nil {
		b.Errorf("error executing request %v", err)
    }

    return resp
}

func BenchmarkApi(b *testing.B) {

    confStr, err := util.Contents("../config.json")
    if err != nil {
        b.Errorf("error reading config %v", err)
        return
    }
    var s map[string]interface{}
    err = json.Unmarshal([]byte(confStr), &s)
    if err != nil {
        b.Errorf("error parsing config %v", err)
        return
    }

    secret := s["SharedSecret"].(string)
    port := s["Port"].(float64)

    // voteAuth := map[string]string{"Authorization": middleware.AuthHeader("voter-1-1", secret)}
	b.ResetTimer()
    for i := 0; i < b.N; i++ {
        now := time.Now()
        voterId := now.Nanosecond()
        header := fmt.Sprintf("voter-1-%d", voterId)
        url := fmt.Sprintf("http://localhost:%d/api/v1/ballotbox/1/%d", int(port), voterId)
        voteAuth := map[string]string{"Authorization": middleware.AuthHeader(header, secret)}
        resp := request("POST", url, voteAuth, newVote, b)
        if resp != nil && resp.StatusCode != http.StatusAccepted {
     		b.Errorf("bad status code %d", resp.StatusCode)
        }
    }
    return
}