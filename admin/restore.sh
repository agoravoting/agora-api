#!/bin/sh
AGORA_API=/home/ballotbox/dist/
ELECTIONS_DIR=$AGORA_API/admin/elections

set -e
[ -z $1 ] && { echo "No backup file $1"; exit 1; }
dir=`mktemp -d`
echo "extracting to $dir"
tar xfz $1 -C $dir
supervisorctl stop agora-api
sudo -u postgres -H bash -l -c "dropdb ballotbox; createdb -O ballotbox ballotbox; pg_restore -d ballotbox $dir/ballotbox.dump"
RESTORE_ELECTIONS_DIR="$ELECTIONS_DIR"_bak
[ -d $RESTORE_ELECTIONS_DIR ] || mkdir $RESTORE_ELECTIONS_DIR
tar xf $dir/elections.tar -C $RESTORE_ELECTIONS_DIR
cp -f $dir/config.json $AGORA_API/config.json.bak
supervisorctl start agora-api
