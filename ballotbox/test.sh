#!/bin/bash
cd $GOPATH/src/github.com/agoravoting/agora-api
goose up
go test github.com/agoravoting/agora-api/ballotbox
goose down