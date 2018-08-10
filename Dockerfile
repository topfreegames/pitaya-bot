FROM golang:latest 

RUN mkdir -p $GOPATH/src/github.com/topfreegames/pitaya-bot
ADD . $GOPATH/src/github.com/topfreegames/pitaya-bot 
WORKDIR $GOPATH/src/github.com/topfreegames/pitaya-bot 

RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure
RUN go build -o main . 

CMD ["./main", "run", "--config", "/etc/pitaya-bot/config.yaml", "--duration", "5m", "-d", "/etc/pitaya-bot/specs"]
