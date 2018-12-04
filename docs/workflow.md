Workflow
=============

In this section we will describe in details the workflow process, since the setup until the end summary. The example is going to assume the usage of a sequential bot with TCP communication and JSON information format.

## Testing a scenario

The overview of what happens when the pitaya-bot is started:

* Initialization of app, configuration fetch, specs directory lookup and creation of metric reporters
* Instantiation of many go routines, which are defined in spec files
* Validation of selected bot and written specs
* Execution of specs
* Notification of the result to all of metrics reporter
* Summarization of all results

### Initialization

The first thing pitaya-bot does is instantiate an App struct based on the config file, receiving the metric reporters that will be used (Promethues, ...) and name of the pitaya game which will be tested.

The configuration is also passed to the bots that will follow the specs, so that they know which storage will be used, endpoint to access, etc. 

Another important point is the directory where the specs are located, because it will use the number of spec files as the number of go routines that will execute each one of them in a parallel way.

### Instances

Based on the spec file, the field `numberOfInstances` will dictate how many go routines will be created to run each of the written scenarios.

### Validation

For each spec, it will validate if it was able to:

* Create the given type of bot
* Initialize the bot
* Run the given spec without problems
* Finalize the bot

### Execution of given spec

In the moment that the bot is initialized, it will fetch all the information contained in the spec and create operations that will be executed. The operations can vary, it can make all the possible operations that a pitaya client can do and also store informations from the received responses. It is important to mention that each bot has different operations that can be used, so consult them before writing your own testing scenarios.

### Metric Reporter

After each request to a pitaya server, the pitaya-bot will inform the metric reporter of the response time, which is important to see the overall QoS.

### Summary

After all the specs have been run, it will gather all the results obtained and return in the screen, informing if it was a total success or if some errors occurred.
