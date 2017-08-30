#!/usr/bin/env bash

set -e

PROBABILITY=50
RANDOM_100=$((RANDOM % 100))

echo "$RANDOM_100 < $PROBABILITY"
if [ "$RANDOM_100" -lt "$PROBABILITY" ]; then
    echo failed
    exit 1
else
    echo succeeded
    exit 0
fi
