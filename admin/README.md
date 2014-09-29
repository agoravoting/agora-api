agora-api admin
===============

Python dependencies

    pip install SQLAlchemy, psycopg2, PrettyTable, argsparse, requests

eopeers: set public ip, private ip, hostname, port at top of file

    PUBLIC_IP_ADDRESS = "1.1.1.1"
    PRIVATE_IP_ADDRESS = "1.1.1.1"
    HOSTNAME = "foo.bar"
    PORT = 8000

Create self signed certificate

    mkdir /srv/certs, mkdir /srv/certs/selfsigned/,

    openssl req -nodes -x509 -newkey rsa:4096 -keyout key-nopass.pem -out cert.pem -days 365 <<EOF
    ${C}
    ${ST}
    ${L}
    ${O}
    ${OU}
    ${CN}
    ${EMAIL}
    EOF

  cp cert.pem calist

Make scripts executable

    chmod u+x on admin, eopeers

Use eopeers to install authority packages

    eopeers --install ..

In an election directory edit config.json. example value

    {
      "election-id": "1",
      "director": "wadobo-auth1",
      "authorities": "wadobo-auth2"
      "title": "Test election",
      "url": "https://example.com/election/url",
      "description": "election description",
      "questions_data": [{
          "question": "Who Should be President?",
          "tally_type": "ONE_CHOICE",
          "answers": [
              {"a": "ballot/answer",
              "details": "",
              "value": "Alice"},
              {"a": "ballot/answer",
              "details": "",
              "value": "Bob"}
          ],
          "max": 1, "min": 0
      }],
      "voting_start_date": "2013-12-06T18:17:14.457000",
      "voting_end_date": "2013-12-09T18:17:14.457000",
      "is_recurring": false,
      "extra": []
    }

The id field is optional, it defaults to the election directory name

Forward port in nginx, for example

    server {
        listen         1.1.1.1:8000;
        server_name    foo.bar;

        location / {
            proxy_pass http://localhost:8000;
        }
    }