# Overseer Overview

[Overseer](https://github.com/skx/overseer) is a network-testing tool, which allows you to run tests of services on remote hosts.  You write your tests as plain-text, and then execute them.

To allow scaling the execution of tests isn't done directly, instead (parsed) tests are added to a (redis) queue, from where they can be pulled.

## Overseer Deployment

There are two parts to deploying overseer:

* Fetching the binary.
* Setting up the services.

There are two services:

* overseer-enqueue.service
  * This adds tests to the queue every two minutes.
* overseer-worker.service
  * This launches a worker which will fetch tests from the queue and execute them.

The recipe in this directory sets up the tests.
