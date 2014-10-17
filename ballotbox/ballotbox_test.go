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

	"flag"
	"strconv"
	"os"
    "crypto/tls"
)

const (
	newVote = `{
	"vote": "0000000000000000000000000000000000000000000000000000000000000000000000000",
	"vote_hash": "hash"
}`
)

var (
	port int
    host *string
	// sharedsecret duplicated here, once used in below test, the other in config passed to server, must match.
	SharedSecret = "somesecret"
	Config       = `{
	"Debug": true,
	"DbMaxIddleConnections": 5,
	"DbConnectString": "user=ballotbox password=ballotbox dbname=ballotbox_test sslmode=disable",
    "ballotboxSessionExpire": 1000000,
	"maxWrites": 2,
    "SharedSecret": "somesecret",
	"Admins": ["test@example.com"],
	"ActiveModules": [
		"github.com/agoravoting/agora-api/ballotbox"
	],
	"RavenDSN": "",
    "electionDir": "../admin/elections"
}`
)

func TestAgoraApi(t *testing.T) {
	ts := stest.New(t, Config)
	defer ts.TearDown()
	voteAuth := map[string]string{"Authorization": middleware.AuthHeader("voter-1-1", SharedSecret)}

	posted := ts.RequestJson("POST", "/api/v1/ballotbox/election/1/vote/1", http.StatusAccepted, voteAuth, newVote)
	fmt.Printf("new vote %v\n", posted)
	posted = ts.RequestJson("POST", "/api/v1/ballotbox/election/1/vote/1", http.StatusAccepted, voteAuth, newVote)
	fmt.Printf("new vote %v\n", posted)

	foundVote := ts.Request("GET", "/api/v1/ballotbox/election/1/check-hash/1/hash", http.StatusOK, voteAuth, "")
	fmt.Printf("found vote %v\n", foundVote)

    // will only work if there is an election (config.json) with election-id 1 in admin/elections
    // cfg := ts.RequestJson("GET", "/api/v1/ballotbox/get_election_config/1/1", http.StatusOK, voteAuth, newVote)
    // fmt.Printf("cfg %v\n", cfg)
}

func request(method string, path string, headers map[string]string, requesTBody string, b *testing.B) *http.Response {
    // skip certificate validation
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }

    client := &http.Client{Transport: tr}
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

// reads config from config.json
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

	b.ResetTimer()
    b.SetParallelism(10)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
    		now := time.Now()
	    	voterId := now.Nanosecond()
            // posting
            // header := fmt.Sprintf("voter-1-%d", voterId)
	    	// url := fmt.Sprintf("http://%s:%d/api/v1/ballotbox/election/1/vote/%d", *host, port, voterId)

            // get config
            header := fmt.Sprintf("voter-1306-%d", voterId)
            url := fmt.Sprintf("https://%s:%d/api/v1/ballotbox/election/1306/config/%d", *host, port, voterId)

            voteAuth := map[string]string{"Authorization": middleware.AuthHeader(header, secret)}
	    	// resp := request("POST", url, voteAuth, newVote, b)
            resp := request("GET", url, voteAuth, "", b)
	    	// for post
            // if resp != nil && resp.StatusCode != http.StatusAccepted {
            // for get config
            if resp != nil && resp.StatusCode != http.StatusOK {
	 			b.Errorf("bad status code %d", resp.StatusCode)
	    	}
	    }
	})

    /*c := make(chan string)
    start := time.Now()

    routines := 30

    for j:= 0; j < routines; j++ {
    	go func(){
    		for i := 0; i < 10; i++ {
	    		now := time.Now()
	        	voterId := now.Nanosecond()
	        	header := fmt.Sprintf("voter-1306-%d", voterId)
                url := fmt.Sprintf("https://%s:%d/api/v1/ballotbox/election/1306/config/%d", *host, port, voterId)
                voteAuth := map[string]string{"Authorization": middleware.AuthHeader(header, secret)}
                // resp := request("POST", url, voteAuth, newVote, b)
                fmt.Printf("=>")
                resp := request("GET", url, voteAuth, "", b)
                fmt.Printf("|")
                // if resp != nil && resp.StatusCode != http.StatusAccepted {
                if resp != nil && resp.StatusCode != http.StatusOK {
                    b.Errorf("bad status code %d", resp.StatusCode)
                }
	        }
        	c <- "ok"
		}()
    }

    for j:= 0; j < routines; j++ {
    	<- c
    }
    delta := time.Now().Sub(start)
    fmt.Printf("elapsed %f", float64(delta) / 1000000000)*/

    return
}

// used to parse command line arguments (see http://golang.org/pkg/flag/ example)
func init() {
	var err error

	addr := flag.String("port", "3000", "http port")
    host = flag.String("host", "localhost", "http host address")
	flag.Parse()
	fmt.Printf("running against %s port %v..\n", *host, *addr)
	port, err = strconv.Atoi(*addr); if err != nil {
    	fmt.Printf("*** error parsing port %v\n", err)
        os.Exit(1)
    }
}