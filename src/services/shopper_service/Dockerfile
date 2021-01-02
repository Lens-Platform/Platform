FROM golang:1.15-alpine as builder

ARG REVISION

RUN mkdir -p /shopper_service/

WORKDIR /shopper_service

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 go build -ldflags "-s -w \
    -X github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/version.REVISION=${REVISION}" \
    -a -o bin/shopper_service cmd/shopper_service/*

RUN CGO_ENABLED=0 go build -ldflags "-s -w \
    -X github.com/BlackspaceInc/BlackspacePlatform/src/services/shopper_service/pkg/version.REVISION=${REVISION}" \
    -a -o bin/podcli cmd/podcli/*

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

COPY --from=builder /shopper_service/bin/shopper_service .
COPY --from=builder /shopper_service/bin/podcli /usr/local/bin/podcli
COPY ./ui ./ui
RUN chown -R app:app ./

USER app

CMD ["./shopper_service"]
