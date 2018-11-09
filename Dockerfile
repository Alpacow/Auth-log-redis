FROM golang

ADD . /go/src/logredis

WORKDIR /go/src

RUN go get github.com/mediocregopher/radix.v2/redis

RUN go install logredis

ENTRYPOINT /go/bin/logredis

EXPOSE 5000