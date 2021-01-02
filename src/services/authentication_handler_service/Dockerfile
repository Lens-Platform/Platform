FROM golang:1.15-alpine as builder

ARG REVISION

RUN mkdir -p /authentication_handler_service/

WORKDIR /authentication_handler_service

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 go build -ldflags "-s -w \
    -X github.com/BlackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/pkg/version.REVISION=${REVISION}" \
    -a -o bin/authentication_handler_service cmd/authentication_handler_service/*

RUN CGO_ENABLED=0 go build -ldflags "-s -w \
    -X github.com/BlackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/pkg/version.REVISION=${REVISION}" \
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

COPY --from=builder /authentication_handler_service/bin/authentication_handler_service .
COPY --from=builder /authentication_handler_service/bin/podcli /usr/local/bin/podcli
COPY ./ui ./ui
RUN chown -R app:app ./

USER app

CMD ["./authentication_handler_service"]
