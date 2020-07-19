FROM golang:alpine

COPY . /redditclone

WORKDIR /redditclone

RUN apk add make && make build

ENTRYPOINT ["/redditclone/bin/myapp"]