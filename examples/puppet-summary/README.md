# Puppet-Summary Overview

[puppet-summary](https://github.com/skx/puppet-summary) presents a simple dashboard showing the activity of a puppet-master.

In short if you use puppet to control a bunch of hosts the dashboard allows you to view failures/successes in a simple fashion.


## Puppet-Summary Deployment

There are two parts to deploying overseer:

* Fetching the binary.
* Setting up the service.

There is only a single service which launches the deamon, ahead of that is a
public-facing `nginx` proxy-server.  (We should configure that with `deployr`
too, but currently we don't.)
