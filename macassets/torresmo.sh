#!/bin/sh
now=$(date +%s)
appdir="$(dirname "$0")"
$appdir/../Resources/torresmo server --gui --watch=downloads --out=$HOME/Downloads --addr=:8000 2>>/tmp/torresmo-${now}.log
