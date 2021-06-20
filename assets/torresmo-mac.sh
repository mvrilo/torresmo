#!/bin/sh
now=$(date +%s)
appdir="$(dirname "$0")"
$appdir/../Resources/torresmo server --gui --watch=$HOME/Downloads/torresmo --out=$HOME/Downloads/torresmo --addr=:8000 2>>/tmp/torresmo-$now.log
