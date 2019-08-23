Features
========

Pitaya-Bot has been developed in conjunction with [Pitaya](https://github.com/topfreegames/pitaya), to allow the usage of every feature contained in Pitaya, inside this testing framework. It has been created to fulfill every possible testing scenario and make it easy to be used, without the need to write code.

Some of its core features are described below.

## No code writing

The tests which will be run don't need the knowledge of Golang. The writting of JSON specs and configuration are more than enough.

## Handler Support

It is only possible to test handlers, due to the fact that this framework is focused on the scenarios which the user takes part.

The tests can be created to test idempotency or stress the server and see how it behaves. 

## Bots

Bots are "fake" users, which will be doing requests to Pitaya servers. All of them must implement the [Bot interface](https://github.com/topfreegames/pitaya-bot/blob/master/bot/bot.go). 

Pitaya-Bot comes with a few implemented bots, and more can be implemented as needed. The current existing bots are:

### Sequential

This bot follows exactly the orders written inside the JSON spec and chronologically, one bot after another in each instance.

## Concurrency

In the test setup, it is possible to inform the number of instances that will be doing it. So that it is possible not only to make integration tests, but also stress tests.

## Monitoring

Pitaya-Bot is configurable to measure the server health via [Prometeus](https://prometheus.io/). It is perfect for the testing, because the tester will be able to see how the server behaves with any number of requests and any handler that he wants to test.

## Storage

Storage is the space that the Bot will retain the information received from Pitaya servers, so that it can be used in future use cases. All of them must implement the [Storage interface](https://github.com/topfreegames/pitaya-bot/blob/master/storage/storage.go).
The desired storage must be set via configuration and will be created via factory method `NewStorage`. Remember to add new storages into this factory.

Pitaya-Bot comes with a few implemented storages, and more can be implemented as needed. The current existing storages are:

### Memory

This storage retains all information inside the testing machine memory. The stored information is not persistent and will be flushed with the end of the test. 

## Custom initialization and wrap-up

Specs can specify custom initialization and wrap-up routines to do operations such as fetching an initial state from some storage and saving the final state to a storage.

To define an initialization function in the script you should create a *preRun* field, with *function* specifying which function should be run. It also accepts *args* as an object with arguments to be passed to the function.

To define a wrap-up function in the script you should create a *postRun* field, with *function* specifying which function should be run. It also accepts *args* as an object with arguments to be passed to the function.

The JSON testing sample has an example with these fields.

### Redis

These initialization and wrap-routines run lua scripts in redis and come with default scripts to fetch a state from a set and save it to another. The preRun script is expected to return the initial state for the bot and the postRun script receives the final state and is expected to do something with it.

The default initialization script tries to fetch an element from the set *${name}:available* and write it to *{name}:used*.

The default wrap-up script writes the state to *${name}:available*.

The initialization script accepts two arguments:

- **name (required)**: the key argument that is passed to the lua script
- **failEmpty (optional)**: a boolean indicating if the method should fail if the script returns nil

The wrap-up script accepts one argument:

- **name (required)**: the key argument that is passed to the lua script

## Serializers

Pitaya-Bot supports both JSON and Protobuf serializers out of the box for the messages sent to and from the client, the default serializer is JSON.

## Spec generation

It is possible to create specs from pitaya-cli history by using the `parseHistory` command.
