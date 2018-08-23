FROM golang:1.10-alpine AS build-env

MAINTAINER TFG Co <backend@tfgco.com>

RUN apk update && apk add git

RUN mkdir -p /go/src/github.com/topfreegames/pitaya-bot
ADD . /go/src/github.com/topfreegames/pitaya-bot

WORKDIR /go/src/github.com/topfreegames/pitaya-bot
RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure
RUN go build -o main .

FROM alpine:3.8

RUN apk update && apk add ca-certificates

WORKDIR /app
COPY --from=build-env /go/src/github.com/topfreegames/pitaya-bot/main /app

CMD ["./main", "run", "--config", "/etc/pitaya-bot/config.yaml", "--duration", "5m", "-d", "/etc/pitaya-bot/specs"]
