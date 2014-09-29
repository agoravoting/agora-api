#!/bin/bash
cd $GOPATH/src/github.com/agoravoting/agora-api
go test -bench=. -v -run BOGUS github.com/agoravoting/agora-api/ballotbox -addr 3000 -cpu 1