FROM alpine:3.8

RUN apk update && apk add ca-certificates

WORKDIR /app
ADD ./out/pitaya-bot-linux /app/main

CMD ["./main", "run", "--config", "/etc/pitaya-bot/config.yaml", "--duration", "5m", "-d", "/etc/pitaya-bot/specs"]
