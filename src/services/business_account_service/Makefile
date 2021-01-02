# Makefile for releasing business_account_service
#
# The release version is controlled from pkg/version

TAG?=latest
NAME:=business_account_service
DOCKER_REPOSITORY:=blackspaceinc
DOCKER_IMAGE_NAME:=$(DOCKER_REPOSITORY)/$(NAME)
GIT_COMMIT:=$(shell git describe --dirty --always)
VERSION:=$(shell grep 'VERSION' pkg/version/version.go | awk '{ print $$4 }' | tr -d '"')
EXTRA_RUN_ARGS?=

run:
	go run -ldflags "-s -w -X github.com/blackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/pkg/version.REVISION=$(GIT_COMMIT)" cmd/$(NAME)//* \
	--level=debug --grpc-port=9999 --backend-url=https://httpbin.org/status/401 --backend-url=https://httpbin.org/status/500 \
	--ui-logo=https://raw.githubusercontent.com/stefanprodan/$(NAME)//gh-pages/cuddle_clap.gif $(EXTRA_RUN_ARGS)

build-cli:
	GIT_COMMIT=$$(git rev-list -1 HEAD) && CGO_ENABLED=0 go build  -ldflags "-s -w -X github.com/blackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/version.REVISION=$(GIT_COMMIT)" -a -o ./bin/business_account_service_cli ./cmd/business_account_service_cli/*
build: build-cli
	GIT_COMMIT=$$(git rev-list -1 HEAD) && CGO_ENABLED=0 go build  -ldflags "-s -w -X github.com/blackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/version.REVISION=$(GIT_COMMIT)" -a -o ./bin/business_account_service ./cmd/business_account_service/*

fmt:
	gofmt -l -s -w ./
	goimports -l -w ./

build-charts:
	helm lint charts/*
	helm package charts/*

build-container:
	docker build -t $(DOCKER_IMAGE_NAME):$(VERSION) .

build-base:
	docker build -f Dockerfile.base -t $(DOCKER_REPOSITORY)/$(NAME)-base:latest .

push-base: build-base
	docker push $(DOCKER_REPOSITORY)/$(NAME)-base:latest

test-container:
	@docker rm -f $(NAME) || true
	@docker run -dp 9898:9898 --name=$(NAME) $(DOCKER_IMAGE_NAME):$(VERSION)
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
	sed -i '' "s/tag: $$current/tag: $$next/g" charts/$(NAME)/values.yaml && \
	sed -i '' "s/tag: $$current/tag: $$next/g" charts/$(NAME)//values-prod.yaml && \
	sed -i '' "s/appVersion: $$current/appVersion: $$next/g" charts/$(NAME)//Chart.yaml && \
	sed -i '' "s/version: $$current/version: $$next/g" charts/$(NAME)//Chart.yaml && \
	sed -i '' "s/$(NAME)/:$$current/$(NAME)/:$$next/g" kustomize/deployment.yaml && \
	sed -i '' "s/$(NAME)/:$$current/$(NAME)/:$$next/g" deploy/webapp/frontend/deployment.yaml && \
	sed -i '' "s/$(NAME)/:$$current/$(NAME)/:$$next/g" deploy/webapp/backend/deployment.yaml && \
	sed -i '' "s/$(NAME)/:$$current/$(NAME)/:$$next/g" deploy/bases/frontend/deployment.yaml && \
	sed -i '' "s/$(NAME)/:$$current/$(NAME)/:$$next/g" deploy/bases/backend/deployment.yaml && \
	echo "Version $$next set in code, deployment, chart and kustomize"

release:
	git tag $(VERSION)
	git push origin $(VERSION)

download_proto_dependencies:
	@echo "downloading protoc-gen tool"
	go install github.com/danielvladco/go-proto-gql/protoc-gen-gql && go install github.com/danielvladco/go-proto-gql/protoc-gen-gogql && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc

generate_graphql: download_proto_dependencies
	protoc -I. \
		-I$(GOPATH)/src \
		-I=$(GOPATH)/src/github.com/infobloxopen/protoc-gen-gorm \
		-I=$(GOPATH)/src/github.com/infobloxopen/atlas-app-toolkit \
		-I=$(GOPATH)/src/github.com/lyft/protoc-gen-validate/validate/validate.proto \
		-I=$(GOPATH)/src/github.com/infobloxopen/protoc-gen-gorm/options \
		-I=$(GOPATH)/src/github.com/protobuf/src/google/protobuf/timestamp.proto \
		--gogoopsee_out=plugins=grpc+graphql,Mopsee/protobuf/opsee.proto=github.com/opsee/protobuf/opseeproto,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:./pkg/graphql_api/ --proto_path=$(GOPATH)/src:. ./proto/*.proto

generate_grpc_definitions: download_proto_dependencies
	@echo "setting up grpc service schema definition via protobuf"
	protoc -I. \
			-I$(GOPATH)/src \
			-I=$(GOPATH)/src/github.com/infobloxopen/protoc-gen-gorm \
			-I=$(GOPATH)/src/github.com/infobloxopen/atlas-app-toolkit \
			-I=$(GOPATH)/src/github.com/lyft/protoc-gen-validate/validate/validate.proto \
			-I=$(GOPATH)/src/github.com/infobloxopen/protoc-gen-gorm/options \
			-I=$(GOPATH)/src/github.com/protobuf/src/google/protobuf/timestamp.proto \
			--proto_path=${GOPATH}/src/github.com/gogo/protobuf/protobuf \
			--govalidators_out=./pkg/graphql_api/ \
			--go_out=./pkg/graphql_api --gorm_out="engine=postgres:./pkg/graphql_api/" ./model/schema.proto

generate_graphql_schema_from_proto_definitions: generate_grpc_definitions
	proto2graphql --input ./model/schema.proto --output ./pkg/graphql_api/schema/schema.graphql

graphql: generate_graphql_schema_from_proto_definitions
	go run github.com/99designs/gqlgen

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

build_and_start_services_locally:
	# starts the necessary services in order to permit e2e testing
	@echo starts the required support container services to allow e2e testing
	docker-compose -f compose/docker-compose.jaeger.yaml \
				   -f compose/docker-compose.postgres.yaml \
				   -f compose/docker-compose.authentication_handler_service.yaml \
				   -f compose/docker-compose.dev.yaml up --force-recreate --build --remove-orphans --always-recreate-deps --renew-anon-volumes -d

start_services_locally:
	# starts the necessary services in order to permit e2e testing
	@echo starts the required support container services to allow e2e testing
	docker-compose -f compose/docker-compose.jaeger.yaml \
				   -f compose/docker-compose.postgres.yaml \
				   -f compose/docker-compose.authentication_handler_service.yaml \
				   -f compose/docker-compose.dev.yaml up --remove-orphans

stop_services_locally:
	@echo stopping the container services used for testing
	docker-compose -f compose/docker-compose.jaeger.yaml \
				   -f compose/docker-compose.postgres.yaml \
 				   -f compose/docker-compose.authentication_handler_service.yaml \
 				   -f compose/docker-compose.dev.yaml down --rmi all --volumes

test:
	go test -v -race -cover -coverprofile=prof.out ./...


test-suite: start_services_locally
	sh -c "echo waiting 25s to let dependencies start... && sleep 25"
	make test
	make stop_services_locally

coverage: build_and_start_services_locally
	sh -c "echo waiting 25s to let dependencies start... && sleep 25"
	make test
	go tool cover -html=prof.out
	make stop_services_locally
