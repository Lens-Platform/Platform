FROM golang:1.15-alpine as builder

ARG REVISION

RUN mkdir -p /business_account_service/

WORKDIR /business_account_service

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 go build -ldflags "-s -w \
    -X github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/version.REVISION=${REVISION}" \
    -a -o bin/business_account_service cmd/business_account_service/*

RUN CGO_ENABLED=0 go build -ldflags "-s -w \
    -X github.com/BlackspaceInc/BlackspacePlatform/src/services/business_account_service/pkg/version.REVISION=${REVISION}" \
    -a -o bin/business_account_service_cli cmd/business_account_service_cli/*

FROM alpine:3.12

ARG BUILD_DATE
ARG VERSION
ARG REVISION

LABEL maintainer="github.com/BlackspaceInc/BlackspacePlatform"

RUN addgroup -S app \
    && adduser -S -g app app \
    && apk --no-cache add \
    ca-certificates curl netcat-openbsd

WORKDIR /home/app

COPY --from=builder /business_account_service/bin/business_account_service .
COPY --from=builder /business_account_service/bin/business_account_service_cli /usr/local/bin/business_account_service_cli
COPY ./ui ./ui
RUN chown -R app:app ./

USER app

CMD ["./business_account_service"]
