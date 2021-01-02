# Makefile for releasing authentication_handler_service
#
# The release version is controlled from pkg/version

TAG?=latest
NAME:=authentication_handler_service
DOCKER_REPOSITORY:=blackspaceinc
DOCKER_IMAGE_NAME:=$(DOCKER_REPOSITORY)/$(NAME)
GIT_COMMIT:=$(shell git describe --dirty --always)
VERSION:=$(shell grep 'VERSION' pkg/version/version.go | awk '{ print $$4 }' | tr -d '"')
EXTRA_RUN_ARGS?=

run:
	go run -ldflags "-s -w -X github.com/blackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/pkg/version.REVISION=$(GIT_COMMIT)" cmd/authentication_handler_service/* \
	--level=debug --grpc-port=9999 --backend-url=https://httpbin.org/status/401 --backend-url=https://httpbin.org/status/500 \
	--ui-logo=https://raw.githubusercontent.com/github.com/blackspaceInc/BlackspacePlatform/authentication_handler_service/gh-pages/cuddle_clap.gif $(EXTRA_RUN_ARGS)

test:
	sleep 10s
	go test -v -race ./...

build:
	GIT_COMMIT=$$(git rev-list -1 HEAD) && CGO_ENABLED=0 go build  -ldflags "-s -w -X github.com/blackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/pkg/version.REVISION=$(GIT_COMMIT)" -a -o ./bin/authentication_handler_service ./cmd/authentication_handler_service/*
	GIT_COMMIT=$$(git rev-list -1 HEAD) && CGO_ENABLED=0 go build  -ldflags "-s -w -X github.com/blackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/pkg/version.REVISION=$(GIT_COMMIT)" -a -o ./bin/podcli ./cmd/podcli/*

fmt:
	gofmt -l -s -w ./
	goimports -l -w ./

build-charts:
	helm lint charts/*
	helm package charts/*

build-container:
	docker build -t $(DOCKER_IMAGE_NAME):$(VERSION) .

build-base:
	docker build -f Dockerfile.base -t $(DOCKER_REPOSITORY)/authentication_handler_service-base:latest .

push-base: build-base
	docker push $(DOCKER_REPOSITORY)/authentication_handler_service-base:latest

test-container:
	@docker rm -f authentication_handler_service || true
	@docker run -dp 9898:9898 --name=authentication_handler_service $(DOCKER_IMAGE_NAME):$(VERSION)
	@docker ps
	sudo curl -L GET --url http://localhost:9898/healthz

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
	sed -i '' "s/tag: $$current/tag: $$next/g" charts/authentication_handler_service/values.yaml && \
	sed -i '' "s/tag: $$current/tag: $$next/g" charts/authentication_handler_service/values-prod.yaml && \
	sed -i '' "s/appVersion: $$current/appVersion: $$next/g" charts/authentication_handler_service/Chart.yaml && \
	sed -i '' "s/version: $$current/version: $$next/g" charts/authentication_handler_service/Chart.yaml && \
	sed -i '' "s/authentication_handler_service:$$current/authentication_handler_service:$$next/g" kustomize/deployment.yaml && \
	sed -i '' "s/authentication_handler_service:$$current/authentication_handler_service:$$next/g" deploy/webapp/frontend/deployment.yaml && \
	sed -i '' "s/authentication_handler_service:$$current/authentication_handler_service:$$next/g" deploy/webapp/backend/deployment.yaml && \
	sed -i '' "s/authentication_handler_service:$$current/authentication_handler_service:$$next/g" deploy/bases/frontend/deployment.yaml && \
	sed -i '' "s/authentication_handler_service:$$current/authentication_handler_service:$$next/g" deploy/bases/backend/deployment.yaml && \
	echo "Version $$next set in code, deployment, chart and kustomize"

release:
	git tag $(VERSION)
	git push origin $(VERSION)

swagger:
	go get github.com/swaggo/swag/cmd/swag
	cd pkg/api && $$(go env GOPATH)/bin/swag init -g server.gogqlproto:
	protoc -I. \
		-I$(GOPATH)/src \
		-I=$(GOPATH)/src/github.com/infobloxopen/protoc-gen-gorm \
		-I=$(GOPATH)/src/github.com/infobloxopen/atlas-app-toolkit \
		-I=$(GOPATH)/src/github.com/lyft/protoc-gen-validate/validate/validate.proto \
		-I=$(GOPATH)/src/github.com/infobloxopen/protoc-gen-gorm/options \
		-I=$(GOPATH)/src/github.com/protobuf/src/google/protobuf/timestamp.proto \
		--gogoopsee_out=plugins=grpc+graphql,Mopsee/protobuf/opsee.proto=github.com/opsee/protobuf/opseeproto,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:./graphql/ --proto_path=$(GOPATH)/src:. ./models/*.proto

protogen: gqlproto
	@echo "setting up service schema definition via protobuf"
	protoc -I. \
			-I$(GOPATH)/src \
			-I=$(GOPATH)/src/github.com/infobloxopen/protoc-gen-gorm \
			-I=$(GOPATH)/src/github.com/infobloxopen/atlas-app-toolkit \
			-I=$(GOPATH)/src/github.com/lyft/protoc-gen-validate/validate/validate.proto \
			-I=$(GOPATH)/src/github.com/infobloxopen/protoc-gen-gorm/options \
			-I=$(GOPATH)/src/github.com/protobuf/src/google/protobuf/timestamp.proto \
			--proto_path=${GOPATH}/src/github.com/gogo/protobuf/protobuf \
			--govalidators_out=./graphql/ \
			--go_out=plugins=grpc:./graphql/ --gorm_out="engine=postgres:./graphql/" ./models/schema.proto

updateSchema: protogen gqlproto
	go run github.com/99designs/gqlgen

down:
	docker-compose down
	./scripts/clean.sh

up:
	docker-compose -f docker-compose.dev.yml -f docker-compose.monitoring.dev.yml config
	docker-compose -f docker-compose.dev.yml -f docker-compose.monitoring.dev.yml up --remove-orphans

kube:
	kompose convert -f docker-compose.prod.yml -o ./deployments/kubemanifests.yml
	kubectl apply -f kubemanifests.yml

deploy-service:
	kubectl apply -f ./deployments/kubemanifests.yml
	kubectl apply -f ./deployments/service.yml

deploy-dev:
	@echo "deploying service to development environment First we attach the linkerd proxy sidecar to the service then we deploy"
	kustomize build ./deploy/overlays/dev | kubectl apply -f-

deploy-staging:
	@echo "deploying service to staging environment. First we attach the linkerd proxy sidecar to the service then we deploy"
	kustomize build ./deploy/overlays/staging | kubectl apply -f-

deploy-production:
	@echo "deploying service to production environment First we attach the linkerd proxy sidecar to the service then we deploy"
	kustomize build ./overlays/production | kubectl apply -f-

deploy: deploy-dev deploy-staging
	@echo "deploying"

start:
	# docker-compose -f docker-compose.dev.yaml build
	docker-compose -f docker-compose.jaeger.yml -f docker-compose.dev.yaml up --remove-orphans

stop:
	docker-compose -f docker-compose.jaeger.yml -f docker-compose.dev.yaml down

start-e2e-dependencies:
	# starts the necessary services in order to permit e2e testing
	@echo starts the required support container services to allow e2e testing
	docker-compose -f docker-compose.jaeger.yml -f docker-compose-authentication_service.yml up --remove-orphans -d

stop-e2e-dependencies:
	@echo stopping the container services used for testing
	docker-compose -f docker-compose.jaeger.yml -f docker-compose-authentication_service.yml down

test-suite: start-e2e-dependencies test stop-e2e-dependencies
