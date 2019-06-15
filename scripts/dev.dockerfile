FROM golang:latest

RUN go get github.com/codegangsta/gin

CMD cd /app \
  && gin -b go-up -i run