Overview
========

Pitaya-Bot is an easy to use, fast and lightweight test server framework for [Pitaya](https://github.com/topfreegames/pitaya).
The goal of pitaya-bot is to provide a basic development framework for testing pitaya servers via integration tests or stress tests.

## Features

* **No code writing** - Pitaya-Bot only needs JSON specs and a configuration YAML, in order to work. It is simple to create and test directly into any environment, be it development or production.
* **Concurrency** - Configurable number of instances, which will run the tests.
* **Monitoring** - Pitaya-Bot is configurable to work with [Prometeus](https://prometheus.io/). It allows the user to see metrics of the server, which the tests are being run. This way, it is possible to run stress or integration tests.
* **Communication** - Communication between server and client enabled for TCP via JSON.
* **Handler Support** - Support handler messages, to emulate the behaviour of customers using Pitaya.
* **Summary** - At the end of the tests, returns if the requests made have the expected responses. Perfect for testing idempotence of the server.

## Who's Using it

Well, right now, only us at TFG Co, are using it, but it would be great to get a community around the project. Hope to hear from you guys soon!

## How To Contribute?

Just the usual: Fork, Hack, Pull Request. Rinse and Repeat. Also don't forget to include tests and docs (we are very fond of both).
