FROM golang:1.11.9-alpine3.9

RUN apk add --no-cache git gcc g++
RUN go get github.com/codegangsta/gin

ENV GO111MODULE=on

CMD cd /app \
  && gin -b go-up -i run