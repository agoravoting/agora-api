# Installation procedure

You need to install some dependencies and setup the environment first. Basically go, godep, git and mercurial:

    apt-get install golang
    apt-get install git-core mercurial

    echo 'export GOPATH=$HOME/go' >> ~/.bashrc
    echo 'export PATH=$GOPATH/bin:$PATH' >> ~/.bashrc
    echo 'export GOBIN=$GOPATH/bin' >> ~/.bashrc
    source ~/.bashrc

    go get github.com/tools/godep

Then, from the agora-api directory:

    godep restore

After that, set up the database:

    su - postgres
    createuser -P ballotbox
    createdb -O ballotbox ballotbox

    go install bitbucket.org/liamstask/goose/cmd/goose

You have to configure the db password both in config.json and in db/dbconf.yml.
Then, you can migrate database:

    goose up

And now you are good to go and run the backend:

    go run main.go

# Testing

First you need to create the test database:

    su - postgres
    createdb -O ballotbox ballotbox_test

To run the tests you can execute:

    chmod u+x ballotbox/test.sh
    ballotbox/test.sh

# Benchmarks

You can benchmark a running server with:

    chmod u+x ballotbox/bench.sh
    ballotbox/bench.sh

CAUTION: make sure the running server is connected to a test database, the benchmark will insert data

