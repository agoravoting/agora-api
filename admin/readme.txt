=== agora-api

 * eopeers: set public ip, private ip, and hostname at top of file

 * mkdir /srv/certs, mkdir /srv/certs/selfsigned/,

 * create self signed certificate:

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

* eopeers to install authority packages

* for an election directory, edit config.json to specify:

{
  "election-id": "1",
  "director": "wadobo-auth1",
  "authorities": "wadobo-auth2"
}

The id field is optional, it defaults to the election directory name

* nginx, if localPort is set to 8000 in eotest

server {
    listen         94.23.34.20:8000;
    server_name    vota.podemos.info;

    location / {
        proxy_pass http://localhost:8000;
    }
}
