cd %GOPATH%\src\github.com\agoravoting\agora-api
go test -bench=. -v -run BOGUS github.com/agoravoting/agora-api/ballotbox -host vota.podemos.info -port 443 -cpu 4 -benchtime 10s
rem go test -bench=. -v -run BOGUS github.com/agoravoting/agora-api/ballotbox -port 3000 -cpu 4