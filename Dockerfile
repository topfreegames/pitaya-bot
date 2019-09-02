FROM golang:1.12-alpine AS build-env

MAINTAINER TFG Co <backend@tfgco.com>

RUN apk update && apk add git

RUN mkdir -p /pitaya-bot
ADD . /pitaya-bot

WORKDIR /pitaya-bot
RUN go build -o main .

FROM alpine:3.8

RUN apk update && apk add ca-certificates

WORKDIR /app
COPY --from=build-env /pitaya-bot/main /app

CMD ["./main", "run", "--config", "/etc/pitaya-bot/config.yaml", "--duration", "5m", "-d", "/etc/pitaya-bot/specs"]
