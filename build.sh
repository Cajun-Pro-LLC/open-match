#!/usr/bin/env bash

NAME=$1
REGISTRY=ghcr.io
ORG=cajun-pro-llc
REPO=open-match
TAG=${REGISTRY}/${ORG}/${REPO}-${NAME}:latest

docker build -t "${TAG}" --build-arg FUNCTION_NAME="${NAME}" .

