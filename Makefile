# SHELL := /usr/bin/env bash
#
# Include shared Makefiles
include project.mk
include standard.mk

DOCKERFILE = ./build/Dockerfile


# Project specific values
IMAGE_REGISTRY?=quay.io
IMAGE_REPOSITORY?=jharrington22
IMAGE_NAME?=iperf-operator

COMMIT_NUMBER=$(shell git rev-list `git rev-list --parents HEAD | egrep "^[a-f0-9]{40}$$"`..HEAD --count)
CURRENT_COMMIT=$(shell git rev-parse --short=8 HEAD)

CONTAINER_VERSION=0.1.$(COMMIT_NUMBER)-$(CURRENT_COMMIT)

IMG?=$(IMAGE_REGISTRY)/$(IMAGE_REPOSITORY)/$(IMAGE_NAME):v$(CONTAINER_VERSION)
IMAGE_URI=${IMG}
IMAGE_URI_LATEST=$(IMAGE_REGISTRY)/$(IMAGE_REPOSITORY)/$(IMAGE_NAME):latest

.PHONY: docker-build
docker-build: build

.PHONY: login
login:
	docker login -u "$$QUAY_BOT_USERNAME" --password "$$QUAY_BOT_PASSWORD" quay.io
