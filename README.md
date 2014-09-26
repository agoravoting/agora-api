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
