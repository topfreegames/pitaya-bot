Test Writing
==========

## Command Options

The execution of pitaya-bot offers many command options, that enable/disable many functionalities and different types of workflows. The command options can be found in the [command section](command_options.html).

## Configuration

It is important to create the config.yaml file before running the tests, so that pitaya-bot knows which server to access and which report metrics to use. The configuration options can be found in the [configuration section](configuration.html).

## Spec Configuration

Before executing any spec, it is possible to use the following options:

* `numberOfInstances`: The number of instances(go routines) that will run the same spec in parallel

## Bots

There are bots that will be able to follow the operations given to them in each spec file. The available bots are:

### Sequential Bot

This bot will follow the orders contained inside a spec file sequentially and chronologically. The possible operation types for it are:

* `Request`: Requests pitaya server being tested
* `Notify`: Notifies pitaya server being tested
* `Function`: Internal operations for the bot, such as:
	* `Disconnect`: Disconnect from pitaya server
	* `Connect`: Connect to pitaya server
	* `Reconnect`: Reconnects to pitaya server
* `Listen`: Listen to push notifications from pitaya server

## Operation

Operation is the generalistic struct which contains the action that the specified bot will do. The fields are:

* `Type`: Type of operation which the bot will do. Each bot has different types
* `Timeout`: Time that the bot has to execute given operation
* `DontWait`: If true fires the operation and goes to next one, defaults to false; valid only for type equals push and request
* `Uri`: URI which the bot will use to make request, notification, listen, ...
* `Args`: Arguments that will be used in given operation
* `Expect`: Expected result from operation
* `Store`: Which field from the response it should retain

## Special Fields

These are fields that when used will fetch the information from given structure:

* `$response`: When used in `Expect` field as key, will get the object response, that can access his attributes via `.` or `[]`
* `$store`: The information contained inside a storage, can be used as a `Expect` value or `Args` value.

### Config example

Below is a simple example of a config file, for another one which is being used, check: [config](https://github.com/topfreegames/pitaya-bot/blob/master/testing/config/config.yaml)

```
game: "example"

storage:
  type: "memory"

server:
  host: "localhost",
  tls: true

prometheus:
  port: 9191
```

### Spec example

Below is a base example of a spec file, for a working example, check: [spec](https://github.com/topfreegames/pitaya-bot/blob/master/testing/specs/default.json)

```
{
  "numberOfInstances": 1,
  "sequentialOperations": [
  {
    "type": "request",
    "uri": "connector.gameHandler.create",
    "expect": {
      "$response.code": {
        "type": "string",
        "value": "200"
      } 
    },
    "store": {
      "playerAccessToken": {
        "type": "string",
        "value": "$response.token"
      }
    }
  }
  ]
}
```

### Testing example

For a complete working example, check the [testing example](https://github.com/topfreegames/pitaya-bot/tree/master/testing).
