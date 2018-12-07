#!/bin/bash

COMMIT=$TRAVIS_COMMIT
TAG=$TRAVIS_BUILD_NUMBER-${COMMIT:0:7}

docker build -t pitaya-bot .

docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"

docker tag pitaya-bot:latest tfgco/pitaya-bot:$TAG
docker push tfgco/pitaya-bot:$TAG
docker tag pitaya-bot:latest tfgco/pitaya-bot:latest
docker push tfgco/pitaya-bot

DOCKERHUB_LATEST=$(python ./scripts/get_latest_tag.py)

if [ "$DOCKERHUB_LATEST" != "$TAG" ]; then
    echo "Last version is not in docker hub!"
    echo "docker hub: $DOCKERHUB_LATEST, expected: $TAG"
    exit 1
fi
