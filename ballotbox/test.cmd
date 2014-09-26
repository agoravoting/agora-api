cd %GOPATH%\src\github.com\agoravoting\agora-api
goose -env=test up
go test -v github.com/agoravoting/agora-api/ballotbox
goose -env=test down