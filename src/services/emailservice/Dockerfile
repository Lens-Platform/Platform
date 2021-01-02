FROM golang:1.13-alpine as builder

ARG REVISION

RUN mkdir -p /email-service/

WORKDIR /email-service

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 go build -ldflags "-s -w \
    -X github.com/BlackspaceInc/email-service/pkg/version.REVISION=${REVISION}" \
    -a -o bin/podinfo cmd/podinfo/*

FROM alpine:3.12

ARG BUILD_DATE
ARG VERSION
ARG REVISION

LABEL maintainer="Blackspaceinc" \
  org.opencontainers.image.created=$BUILD_DATE \
  org.opencontainers.image.url="https://github.com/Blackspaceinc/email-service" \
  org.opencontainers.image.source="https://github.com/Blackspaceinc/email-service" \
  org.opencontainers.image.version=$VERSION \
  org.opencontainers.image.revision=$REVISION \
  org.opencontainers.image.vendor="Blackspaceinc" \
  org.opencontainers.image.title="email service (blackspace)" \
  org.opencontainers.image.description="blackspace platform | email service" \
  org.opencontainers.image.licenses="MIT"

RUN addgroup -S app \
    && adduser -S -g app app \
    && apk --no-cache add \
    ca-certificates curl netcat-openbsd

WORKDIR /home/app

COPY --from=builder /email-service/bin/podinfo .
COPY ./ui ./ui
RUN chown -R app:app ./

USER app

CMD ["./podinfo"]
