# Makefile for releasing podinfo
#
# The release version is controlled from pkg/version

TAG?=latest
NAME:=email-service
DOCKER_REPOSITORY:=blackspaceinc
DOCKER_IMAGE_NAME:=$(DOCKER_REPOSITORY)/$(NAME)
GIT_COMMIT:=$(shell git describe --dirty --always)
VERSION:=$(shell grep 'VERSION' pkg/version/version.go | awk '{ print $$4 }' | tr -d '"')
EXTRA_RUN_ARGS?=

# CI Specific Parameters
PROJECT_NAME := "email-service"
PKG := "github.com/BlackspaceInc/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)

.PHONY: run
run:
	@echo "run server"
	go run -ldflags "-s -w -X github.com/BlackspaceInc/email-service/version.REVISION=$(GIT_COMMIT)" cmd/podinfo/* \
	--level=debug --grpc-port=9999 --backend-url=https://httpbin.org/status/401 --backend-url=https://httpbin.org/status/500 \
	--ui-logo=https://raw.githubusercontent.com/stefanprodan/podinfo/gh-pages/cuddle_clap.gif $(EXTRA_RUN_ARGS)

test:
	go test -v -race ./...

build:
	GIT_COMMIT=$$(git rev-list -1 HEAD) && CGO_ENABLED=0 go build  -ldflags "-s -w -X github.com/BlackspaceInc/email-service/version.REVISION=$(GIT_COMMIT)" -a -o ./bin/podinfo ./cmd/podinfo/*
	GIT_COMMIT=$$(git rev-list -1 HEAD) && CGO_ENABLED=0 go build  -ldflags "-s -w -X github.com/BlackspaceInc/email-service/version.REVISION=$(GIT_COMMIT)" -a -o ./bin/podcli ./cmd/podcli/*

fmt:
	gofmt -l -s -w ./
	goimports -l -w ./

build-charts:
	helm lint charts/*
	helm package charts/*

build-container:
	docker build -t $(DOCKER_IMAGE_NAME):$(VERSION) .

build-base:
	docker build -f Dockerfile.base -t $(DOCKER_REPOSITORY)/podinfo-base:latest .

push-base: build-base
	docker push $(DOCKER_REPOSITORY)/podinfo-base:latest

test-container:
	@docker rm -f podinfo || true
	@docker run -dp 9898:9898 --name=podinfo $(DOCKER_IMAGE_NAME):$(VERSION)
	@docker ps
	@TOKEN=$$(curl -sd 'test' localhost:9898/token | jq -r .token) && \
	curl -sH "Authorization: Bearer $${TOKEN}" localhost:9898/token/validate | grep test

push-container:
	docker tag $(DOCKER_IMAGE_NAME):$(VERSION) $(DOCKER_IMAGE_NAME):latest
	docker push $(DOCKER_IMAGE_NAME):$(VERSION)
	docker push $(DOCKER_IMAGE_NAME):latest
	docker tag $(DOCKER_IMAGE_NAME):$(VERSION) quay.io/$(DOCKER_IMAGE_NAME):$(VERSION)
	docker tag $(DOCKER_IMAGE_NAME):$(VERSION) quay.io/$(DOCKER_IMAGE_NAME):latest
	docker push quay.io/$(DOCKER_IMAGE_NAME):$(VERSION)
	docker push quay.io/$(DOCKER_IMAGE_NAME):latest

version-set:
	@next="$(TAG)" && \
	current="$(VERSION)" && \
	sed -i '' "s/$$current/$$next/g" pkg/version/version.go && \
	sed -i '' "s/tag: $$current/tag: $$next/g" charts/podinfo/values.yaml && \
	sed -i '' "s/appVersion: $$current/appVersion: $$next/g" charts/podinfo/Chart.yaml && \
	sed -i '' "s/version: $$current/version: $$next/g" charts/podinfo/Chart.yaml && \
	sed -i '' "s/podinfo:$$current/podinfo:$$next/g" kustomize/deployment.yaml && \
	sed -i '' "s/podinfo:$$current/podinfo:$$next/g" deploy/webapp/frontend/deployment.yaml && \
	sed -i '' "s/podinfo:$$current/podinfo:$$next/g" deploy/webapp/backend/deployment.yaml && \
	sed -i '' "s/podinfo:$$current/podinfo:$$next/g" deploy/bases/frontend/deployment.yaml && \
	sed -i '' "s/podinfo:$$current/podinfo:$$next/g" deploy/bases/backend/deployment.yaml && \
	echo "Version $$next set in code, deployment, chart and kustomize"

release:
	git tag $(VERSION)
	git push origin $(VERSION)

swagger:
	go get github.com/swaggo/swag/cmd/swag
	cd pkg/api && $$(go env GOPATH)/bin/swag init -g server.go

.PHONY: zipkin
zipkin:
	@echo 'zipkin server up'
	GO111MODULE=off docker run -it -d --rm -p 9411:9411 openzipkin/zipkin

.PHONY: modd
modd: clean queue-up zipkin
	@echo "development server up with configured auto-reload"
	GO111MODULE=on modd

.PHONY: queue-up
queue-up:
	@echo "spinning up rabbitMQ instance"
	docker run --detach --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management

.PHONY: clean
clean:
	@echo 'running cleanup scrip to remove all docker containers, volumes, networks, cached values'
	./scripts/cleanup.sh