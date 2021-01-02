FROM golang

RUN mkdir -p /go/src/github.com/BlackspaceInc/email-service

ADD . /go/src/github.com/BlackspaceInc/email-service
WORKDIR /go/src/github.com/BlackspaceInc/email-service

RUN go get -t -v ./...
RUN go get github.com/canthefason/go-watcher
RUN go install github.com/canthefason/go-watcher/cmd/watcher

ENTRYPOINT watcher -run /go/src/github.com/BlackspaceInc/email-service/cmd/podinfo -watch github.com/BlackspaceInc/email-service
