*************
Configuration
*************

Pitaya-Bot uses Viper to control its configuration. Below we describe the configuration variables split by topic. We judge the default values are good for most cases, but might need to be changed for some use cases. The default directory for the config file is: ./config/config.yaml.

General
=================

These are general configurations

.. list-table::
  :widths: 15 10 10 50
  :header-rows: 1
  :stub-columns: 1

  * - Configuration
    - Default value
    - Type
    - Description
  * - game
    - 
    - string
    - Name of the application being tested, to appear in Prometheus

Prometheus
=================

These configuration values configure the Prometheus monitoring service to check how the server being tested is behaving. To monitor the application, the option `report-metrics` must be true when starting the Pitaya-Bot.

.. list-table::
  :widths: 15 10 10 50
  :header-rows: 1
  :stub-columns: 1

  * - Configuration
    - Default value
    - Type
    - Description
  * - prometheus.port
    - 9191
    - int
    - Port which the Prometheus instance will run

Server
===========

The configurations needed to access the Pitaya server being tested

.. list-table::
  :widths: 15 10 10 50
  :header-rows: 1
  :stub-columns: 1

  * - Configuration
    - Default value
    - Type
    - Description
  * - server.host
    - localhost
    - string
    - Pitaya server host
  * - server.tls
    - false
    - bool
    - Boolean to enable/disable TLS to connect with Pitaya server
  * - server.serializer
    - json
    - string
    - must be json or protobuffer
  * - server.protobuffer.docs
    - ""
    - string
    - Route for server documentation. Target server must implement handlers for protobuf descriptors and auto documentation.
  * - server.protobuffer.pushinfo.routes
    - []
    - []string
    - Information about the protos used by push messages from the server, this part contains the routes of the messages
  * - server.protobuffer.pushinfo.protos
    - []
    - []string
    - Information about the protos used by push messages from the server, this part contains the names of the protos

If your application use protobuffers, specifying docs is required. You can also add a list of routes and protobuffer types if your application sends push information to the bot. See :file:`testing/protobuffer/config/config.yaml` for example.

Storage
==========

.. list-table::
  :widths: 15 10 10 50
  :header-rows: 1
  :stub-columns: 1

  * - Configuration
    - Default value
    - Type
    - Description
  * - storage.type
    - memory
    - string
    - Type of storage which the bot will use

Kubernetes
==========

.. list-table::
  :widths: 15 10 10 50
  :header-rows: 1
  :stub-columns: 1

  * - Configuration
    - Default value
    - Type
    - Description
  * - kubernetes.config
    - $HOME/.kube/config
    - string
    - Path where kubernetes configuration file is located
  * - kubernetes.context
    - 
    - string
    - Kubernetes configuration file context
  * - kubernetes.cpu
    - 250m
    - string
    - CPU which will be allocated for each Kubernetes Pod
  * - kubernetes.image
    - tfgco/pitaya-bot:latest
    - string
    - Pitaya-Bot docker image that kubernetes will use to deploy pods
  * - kubernetes.imagepull
    - Always
    - string
    - Kubernetes docker image pull policy
  * - kubernetes.masterurl
    - 
    - string
    - Master URL for Kubernetes
  * - kubernetes.memory
    - 256Mi
    - string
    - RAM Memory which will be allocated for each Kubernetes Pod
  * - kubernetes.namespace
    - default
    - string
    - Kubernetes namespace that will be used to deploy the jobs
  * - kubernetes.job.retry
    - 0
    - int
    - Backoff limit from the jobs that will run each spec file

Manager
==========

.. list-table::
  :widths: 15 10 10 50
  :header-rows: 1
  :stub-columns: 1

  * - Configuration
    - Default value
    - Type
    - Description
  * - manager.maxrequeues
    - 5
    - int
    - Maximum number of requeues that will be done, if some error occurs while processing a job
  * - manager.wait
    - 1s
    - time.Period
    - Waiting time between each job process

Bot
==========

.. list-table::
  :widths: 15 10 10 50
  :header-rows: 1
  :stub-columns: 1

  * - Configuration
    - Default value
    - Type
    - Description
  * - bot.operation.maxSleep
    - 500ms
    - time.Duration
    - Maximum sleep duration between bot operations, the launcher selects a random value in the range [0, maxSleep]
  * - bot.operation.stopOnError
    - false
    - bool
    - Defines if the bot should stop running on error, by default it restarts the spec
  * - bot.spec.parallelism
    - 1
    - int
    - Defines the number of instances to run for each spec when running on kubernetes

Custom initialization and wrap-up
==========

.. list-table::
  :widths: 15 10 10 50
  :header-rows: 1
  :stub-columns: 1

  * - Configuration
    - Default value
    - Type
    - Description
  * - custom.redis.pre.url
    - redis://localhost:9010
    - string
    - Redis url to connect if using a custom redis initialization
  * - custom.redis.pre.connectionTimeout
    - 10
    - int
    - Timeout in seconds to connect to redis
  * - custom.redis.pre.script
    - ""
    - string
    - Path to the lua script to run if using a custom redis initialization
  * - custom.redis.post.url
    - redis://localhost:9010
    - string
    - Redis url to connect if using a custom redis wrap-up
  * - custom.redis.post.connectionTimeout
    - 10
    - int
    - Timeout in seconds to connect to redis
  * - custom.redis.post.script
    - ""
    - string
    - Path to the lua script to run if using a custom redis wrap-up
