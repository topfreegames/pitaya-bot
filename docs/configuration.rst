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
    - Port which the Prometheus instance is running

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
  * - server.docs
    - ""
    - string
    - Route for server documentation. Target server must implement remotes for protobuf descriptors and auto documentation.

At the server is possible to insert two list of of parameters, server.pushinfo.routes and server.pushinfo.protos. This the the type of the protobufer for a especific route in a push response.

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
    - Path where kubernetes config file is located
  * - kubernetes.masterurl
    - 
    - string
    - Master URL for Kubernetes
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
