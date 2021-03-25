# pitaya-bot [![GoDoc][1]][2] [![Docs][3]][4] [![Go Report Card][5]][6] [![MIT licensed][7]][8] [![Build Status][9]][10] [![Coverage Status][11]][12]

[1]: https://godoc.org/github.com/topfreegames/pitaya-bot?status.svg
[2]: https://godoc.org/github.com/topfreegames/pitaya-bot
[3]: https://readthedocs.org/projects/pitaya-bot/badge/?version=latest
[4]: https://pitaya-bot.readthedocs.io/en/latest/?badge=latest
[5]: https://goreportcard.com/badge/github.com/topfreegames/pitaya-bot
[6]: https://goreportcard.com/report/github.com/topfreegames/pitaya-bot
[7]: https://img.shields.io/badge/license-MIT-blue.svg
[8]: LICENSE
[9]: https://travis-ci.com/topfreegames/pitaya-bot.svg?branch=master
[10]: https://travis-ci.com/topfreegames/pitaya-bot
[11]: https://coveralls.io/repos/github/topfreegames/pitaya-bot/badge.svg?branch=master
[12]: https://coveralls.io/github/topfreegames/pitaya-bot?branch=master

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
go get -u github.com/topfreegames/pitaya-bot
```
setup pitaya-bot dependencies
```
cd $GOPATH/src/github.com/topfreegames/pitaya-bot/
make setup
```

### Running pitaya-bot

Here's how to run the testing example with JSON serializer:

Start the dependencies (this command requires docker-compose, but you may run the dependencies locally if need be) and the pitaya server:
```
$ make run-testing-json-server
```

Now a pitaya server should be running in one terminal. In another one, you can run pitaya-bot with the test specs:
```
$ make run-testing-json-bots
```

For the examples with protobuf, instead run:
```
$ make run-testing-proto-server
$ make run-testing-proto-server
```

## Running the tests
```
make test
```

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

