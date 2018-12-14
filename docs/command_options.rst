****************
Command Options
****************

Pitaya-Bot is a CLI application, that has many command options, which will be described below by topic. We judge the default values are good for most cases, but might need to be changed for some use cases. The default verbosity for the application logger is Debug.

Pitaya-Bot
=================

Base configuration needed to run pitaya-bot

.. list-table::
  :widths: 15 10 10 10 50
  :header-rows: 1
  :stub-columns: 1

  * - Command
    - Command Letter
    - Default value
    - Type
    - Description
  * - config
    - 
    - ./config/config.yaml
    - string
    - Config file path from pitaya-bot
  * - dir
    - d
    - ./specs/
    - string
    - Specs directory
  * - duration
    - 
    - 1m
    - time.Duration
    - Minimum total duration of tests
  * - report-metrics
    - 
    - false
    - bool
    - Enable/Disable metrics reporter
  * - pitaya-bot-type
    - t
    - local
    - string
    - Pitaya-Bot workflow type that will be executed. It can be: local, local-manager, remote-manager, deploy-manager
  * - kill
    - k
    - false
    - bool
    - Should delete all pitaya-bot instances from game(configuration parameter) inside kubernetes

Logger
=================

These are logging configurations

.. list-table::
  :widths: 15 10 10 10 50
  :header-rows: 1
  :stub-columns: 1

  * - Command
    - Command Letter
    - Default value
    - Type
    - Description
  * - verbose
    - v
    - 3
    - int
    - Logger verbosity level => v0: Error, v1=Warning, v2=Info, v3=Debug
  * - logJSON
    - j
    - false
    - bool
    - Enable/Disable logJSON output mode

