#!/usr/bin/env bash

export DHARITRITESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

$DHARITRITESTNETSCRIPTSDIR/stop.sh
$DHARITRITESTNETSCRIPTSDIR/clean.sh
$DHARITRITESTNETSCRIPTSDIR/config.sh
$DHARITRITESTNETSCRIPTSDIR/start.sh $1
