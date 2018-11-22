# pitaya-bot [![GoDoc][1]][2] [![Go Report Card][3]][4] [![MIT licensed][5]][6]

[1]: https://godoc.org/github.com/topfreegames/pitaya-bot?status.svg
[2]: https://godoc.org/github.com/topfreegames/pitaya-bot
[3]: https://goreportcard.com/badge/github.com/topfreegames/pitaya-bot
[4]: https://goreportcard.com/report/github.com/topfreegames/pitaya-bot
[5]: https://img.shields.io/badge/license-MIT-blue.svg
[6]: LICENSE

Pitaya-Bot is an easy to use, fast and lightweight test server framework for [Pitaya](https://github.com/topfreegames/pitaya).
The goal of pitaya-bot is to provide a basic development framework for testing pitaya servers via integration tests or stress tests.

## Getting Started

### Prerequisites

* [Go](https://golang.org/) >= 1.10
* [etcd](https://github.com/coreos/etcd) (optional: for running the testing example)
* [nats](https://github.com/nats-io/go-nats) (optional: for running the testing example)
* [docker](https://www.docker.com) (optional: for running the testing example)

### Installing
clone the repo
```
go get -u github.com/topfreegames//pitaya-bot
```
setup pitaya-bot dependencies
```
cd $GOPATH/src/github.com/topfreegames/pitaya-bot/
make setup
```

### Hacking pitaya-bot

Here's how to run the testing example:

Start etcd and nats (this command requires docker-compose and will run etcd and nats containers locally, you may run etcd and/or nats without docker if you prefer)
```
docker-compose -f ./testing/docker-compose.yml up -d etcd
```
run the server from testing example
```
make run-testing-server
```

Now, it's supposed to be a pitaya server running in one terminal.
In another terminal, you can use pitaya-bot testing example:
```
$ pitaya-bot -d ./testing/specs/ --config ./testing/config/config.yaml
testing/specs/default.json 1755
INFO[0000] Found 1 specs to be executed                  function=launch source=pitaya-bot
...
```

## Running the tests
```
make test
```
This command will run both unit and e2e tests.

## Contributing
#TODO

## Authors
* **TFG Co** - Initial work

## License
[MIT License](./LICENSE)

## Resources

- Other pitaya-related projects
  + [libpitaya](https://github.com/topfreegames/libpitaya)
  + [libpitaya-cluster](https://github.com/topfreegames/libpitaya-cluster)
  + [pitaya](https://github.com/topfreegames/pitaya)
  + [pitaya-admin](https://github.com/topfreegames/pitaya-admin)
  + [pitaya-cli](https://github.com/topfreegames/pitaya-cli)
  + [pitaya-protos](https://github.com/topfreegames/pitaya-protos)

- Documents
  + [API Reference](https://godoc.org/github.com/topfreegames/pitaya-bot)

