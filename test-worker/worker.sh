#!/bin/bash

set -e

STEP=$1

if [ "$STEP" == "echo" ]; then
	echo "curl" # payload for next step
	exit $?
fi

if [ "$STEP" == "curl" ]; then
	RESULT=`curl -s -o /dev/null -i -w "%{http_code}" https://production--workflow-manager.int.clever.com/_health`
	if [ "$RESULT" == "200" ]; then
		echo "success" # TODO payload for next test
		exit 0
	fi

	echo "error: $RESULT"
	exit 1
fi

echo "Unknown step: $STEP"

exit 1
