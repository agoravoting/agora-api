#!/bin/sh
BACKUP_DIR=/home/ballotbox/backup
AGORA_API=/home/ballotbox/dist/
ELECTIONS_DIR=$AGORA_API/admin/elections
NGINX_LOG=/var/log/nginx/
AGORA_API_LOG=/var/log/supervisor/

supervisorctl stop agora-api
now=`date +%Y%m%d%H%M%S`
[ -d  $BACKUP_DIR ] || mkdir $BACKUP_DIR
sudo -u postgres -H bash -l -c "pg_dump ballotbox -Fc -i -b -f /tmp/ballotbox.dump"
mv /tmp/ballotbox.dump $BACKUP_DIR
cp $AGORA_API/config.json $BACKUP_DIR
tar cf $BACKUP_DIR/elections.tar -C $ELECTIONS_DIR .
cp /etc/nginx/conf.d/agora.conf $BACKUP_DIR
tar cf $BACKUP_DIR/nginx_log.tar -C /var/log/nginx/ .
cp $AGORA_API_LOG/agora-api*.log $BACKUP_DIR

tar cfz backup_$now.tar.gz -C $BACKUP_DIR .
supervisorctl start agora-api

# ftp account info
HOST='your host'
USER='your user'
PASSWD='your password here'

ftp -n -v $HOST << EOT
bin
user $USER $PASSWD
prompt
put backup_$now.tar.gz
ls -la
bye
EOT
