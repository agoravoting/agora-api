#!/bin/bash
cd $GOPATH/src/github.com/agoravoting/agora-api
goose -env=test up
go test -bench=. -v -run BOGUS github.com/agoravoting/agora-api/ballotbox -addr 3000
goose -env=test down