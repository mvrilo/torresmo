#!/bin/sh

set -e

main() {
	lastversion=$(git tag -l | tail -n1 | sed -e 's/v//')
	nextversion=$(cat ./VERSION)

	if [ "$lastversion" = "$nextversion" ]; then
		echo 'Set the new version in the VERSION file'
		exit 1
	fi

	git tag "v${nextversion}"
	git push origin "v${nextversion}"
	exit 0
}

main "$@"
