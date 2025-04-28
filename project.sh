#!/bin/bash

# run or build project only via this script

docker-compose --env-file ./cmd/.env "$@"

