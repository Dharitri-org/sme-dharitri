#!/usr/bin/env bash

export DHARITRITESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "$DHARITRITESTNETSCRIPTSDIR/variables.sh"

DESTINATION=mixed
if [[ -n $1 ]]; then
  DESTINATION=$1
fi

NUM_SHARDS=${SHARDCOUNT}
NUM_TXS=250
VALUE=1
GAS_PRICE=200000000000
GAS_LIMIT=100000
SLEEP_DURATION=5

echo 'Starting basic scenario cURL script...'
echo 'Will run a cURL with the following JSON payload each '$SLEEP_DURATION' seconds and each 10 cURLs will send with recalling nonce'
echo '{"value": '$VALUE', "numOfTxs": '$NUM_TXS', "numOfShards": '$NUM_SHARDS', "gasPrice": '$GAS_PRICE', "gasLimit": '$GAS_LIMIT', "destination": "'$DESTINATION'", "recallNonce": false, "scenario": "basic"}'
echo
echo 'Starting...'

let COUNT=0
for (( ; ; )); do
  curl -d '{
     "value": '$VALUE',
     "numOfTxs": '$NUM_TXS',
     "numOfShards": '$NUM_SHARDS',
     "gasPrice": '$GAS_PRICE',
     "gasLimit": '$GAS_LIMIT',
     "destination": "'$DESTINATION'",
     "recallNonce": false,
     "scenario": "basic"
   }' \
    -H "Content-Type: application/json" -X POST http://localhost:${PORT_TXGEN}/transaction/send-multiple

  let COUNT++
  sleep $SLEEP_DURATION
  if (($COUNT % 50 == 0)); then
    curl -d '{
           "value": '$VALUE',
           "numOfTxs": '$NUM_TXS',
           "numOfShards": '$NUM_SHARDS',
           "gasPrice": '$GAS_PRICE',
           "gasLimit": '$GAS_LIMIT',
           "destination": "'$DESTINATION'",
           "recallNonce": true,
           "scenario": "basic"
         }' \
      -H "Content-Type: application/json" -X POST http://localhost:${PORT_TXGEN}/transaction/send-multiple
  fi
done
