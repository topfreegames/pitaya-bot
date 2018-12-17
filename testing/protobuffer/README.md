# pitaya-bot-testing
Player handler for testing purposes based on [pitaya](https://github.com/topfreegames/pitaya)

refs: https://github.com/topfreegames/pitaya

## Required
- golang
- docker-compose

## Run
First of all, in one terminal, make the application that we will test online:
```
docker-compose -f docker-compose.yml up -d etcd nats
go run main.go
```

You need to create the protobuffer files with:
```
make protos-compile
```

After making it online, in another terminal, it's time to test it via pitaya-bot:
```
go run ./../main.go run
```

If pitaya-bot has been built, instead of the command go run, it's possible to use:
```
pitaya-bot run
```
