agora-api
=========

apt-get install golang

apt-get install git-core mercurial

echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$GOPATH/bin:$PATH' >> ~/.bashrc
echo 'export GOBIN=$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc

go get github.com/tools/godep
go get github.com/agoravoting/agora-http-go/

su - postgres
createuser -P ballotbox
createdb -O ballotbox ballotbox

go install bitbucket.org/liamstask/goose/cmd/goose

goose up

go run main.go

FIXME: the port is currently hardcoded in the go file, should be read from config.json

testing
=========
su - postgres
createdb -O ballotbox ballotbox_test

chmod u+x ballotbox/test.sh
ballotbox/test.sh

benchmarks
=========
You can benchmark a running server with

chmod u+x ballotbox/bench.sh
ballotbox/bench.sh

CAUTION: make sure the running server is connected to a test database, the benchmark will insert data