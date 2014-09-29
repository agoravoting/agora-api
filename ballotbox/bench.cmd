cd %GOPATH%\src\github.com\agoravoting\agora-api
go test -bench=. -v -run BOGUS github.com/agoravoting/agora-api/ballotbox -host vota.podemos.info -port 80 -cpu 1