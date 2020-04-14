[![Go Report Card](https://goreportcard.com/badge/github.com/skx/deployr)](https://goreportcard.com/report/github.com/skx/deployr)
[![license](https://img.shields.io/github/license/skx/deployr.svg)](https://github.com/skx/deployr/blob/master/LICENSE)
[![Release](https://img.shields.io/github/release/skx/deployr.svg)](https://github.com/skx/deployr/releases/latest)



Table of Contents
=================

* [deployr](#deployr)
* [Installation &amp; Dependencies](#installation--dependencies)
  * [Source Installation go &lt;=  1.11](#source-installation-go---111)
  * [Source installation go  &gt;= 1.12](#source-installation-go---112)
* [Overview](#overview)
  * [Authentication](#authentication)
  * [Examples](#examples)
  * [File Globs](#file-globs)
* [Variables](#variables)
  * [Predefined Variables](#predefined-variables)
* [Template Expansion](#template-expansion)
* [Missing Primitives?](#missing-primitives)
  * [Alternatives](#alternatives)
* [Github Setup](#github-setup)


# deployr

`deployr` is a simple utility which is designed to allow you to easily automate simple application-deployment via SSH.

The core idea behind `deployr` is that installing (simple) software upon remote hosts frequently consists of a small number of steps:

* Uploading a small number of files, for example:
  * A binary application.
  * A configuration-file.
  * A systemd unit-file
  * etc.
* Running a small number of commands, some conditionally, for example:
  * Enable the systemd unit-file.
  * Start the service.

This is particularly true for golang-based applications which frequently consist of an single binary, a single configuration file, and an init-file to ensure the service can be controlled.

If you want to keep your deployment recipes automatable, and reproducible, then scripting them with a tool like this is ideal.  (Though you might prefer something more popular & featureful such as `ansible`, `fabric`, `salt`, etc.)

"Competing" systems tend to offer more facilities, such as the ability to add Unix users, setup MySQL database, add cron-entries, etc.  Although it isn't impossible to do those things in `deployr` it is not as natural as other solutions.  (For example you can add a cron-entry by uploading a file to `/etc/cron.d/my-service`, or you can add a user via `Run adduser bob 2>/dev/null`.)

One obvious facility that most similar systems, such as ansible, offer is the ability to perform looping operations, and comparisons.  We don't offer that and I'm not sure we ever will - even if we did add the ability to add cronjobs, etc.

In short think of this as an alternative to using a bash-script, which invokes scp/rsync/ssh.  It is not going to compete with ansible, or similar.  (Though it is reasonably close in spirit to `fabric` albeit with a smaller set of primitives.)



## Installation & Dependencies

There are two ways to install this project from source, which depend on the version of the [go](https://golang.org/) version you're using.


### Source Installation go <=  1.11

If you're using `go` before 1.11 then the following command should fetch/update `deployr`, and install it upon your system:

     $ go get -u github.com/skx/deployr

### Source installation go  >= 1.12

If you're using a more recent version of `go` (which is _highly_ recommended), you need to clone to a directory which is not present upon your `GOPATH`:

    git clone https://github.com/skx/deployr
    cd deployr
    go install


If you don't have a golang environment setup you should be able to download a binary for GNU/Linux from [our release page](https://github.com/skx/deployr/releases).



## Overview

`deployr` has various sub-commands, the most useful is the `run` command which
allows you to execute a recipe-file:

    $ deployr run [options] recipe1 recipe2 .. recipeN

Each specified recipe is parsed and the primitives inside them are then executed line by line.  The following primitives/commands are available:

* `CopyFile local/path remote/path`
  * Copy the specified local file to the specified path on the remote system.
  * If the local & remote files were identical, such that no change was made, then this fact will be noted.
  * See later note on globs.
* `CopyTemplate local/path remote/path`
  * Copy the specified local file to the specified path on the remote system, expanding variables prior to running the copy.
  * If the local & remote files were identical, such that no change was made, then this fact will be noted.
  * See later note on globs.
* `DeployTo [user@]hostname[:port]`
  * Specify the details of the host to connect to, this is useful if a particular recipe should only be applied against a single host.
  * If you don't specify a target within your recipe itself you can instead pass it upon the command-line via the `-target` flag.
* `IfChanged "Command"`
  * The `CopyFile` and `CopyTemplate` primitives record whether they made a change to the remote system.
  * The `IfChanged` primitive will execute the specified command if the previous copy-operation resulted in the remote system being changed.
* `Run "Command"`
  * Run the given command (unconditionally) upon the remote-host.
* `Set name "value"`
  * Set the variable "name" to have the value "value".
  * Once set a variable can be used in the recipe, or as part of template-expansion.
* `Sudo` may be added as a prefix to `Run` and `IfChanged`.
  * If present this will ensure the specified command runs as `root`.
  * The sudo example found beneath [examples/sudo/](examples/sudo/) demonstrates usage.



### Authentication

Public-Key authentication is only supported mechanism for connecting to a remote host, or remote hosts.  There is zero support for authentication via passwords.

By default `~/.ssh/id_rsa` will be used as the key to connect with, but if you prefer you can specify a different private-key with the `-identity` flag to the run sub-command:

    $ deployr run -identity ~/.ssh/host

In addition to using a key specified via the command-line deployr also supports the use of `ssh-agent`.  Simply set the environmental-variable `SSH_AUTH_SOCK` to the path of your agent's socket.



### Examples

There are several examples included beneath [examples/](examples/), the shortest one [examples/simple/](examples/simple/) is a particularly good recipe to examine to get a feel for the system:

    $ cd ./examples/simple/
    $ deployr run -target [user@]host.example.com[:port] ./deployr.recipe

For more verbose output the `-verbose` flag may be added:

    $ cd ./examples/simple/
    $ deployr run -target [user@]host.example.com[:port] -verbose ./deployr.recipe

Some other flags are also available, consult "`deployr help run`" for details.


### File Globs

Both the `CopyFile` and `CopyTemplate` primitives allow the use of file-globs,
which allows you to write a line like this:

    CopyFile lib/systemd/system/* /lib/systemd/system/

Assuming you have the following input this will copy all the files, as you
would expect:

      ├── deploy.recipe
      └── lib
          └── systemd
              └── system
                  ├── overseer-enqueue.service
                  ├── overseer-enqueue.timer
                  ├── overseer-worker.service
                  └── purppura-bridge.service

**NOTE** That this wildcard support is _not_ the same as a recursive copy,
that is not supported.

The `IfChanged` primitive will regard a previous copy operation as having
resulted in a change if any single file changes during the run of a copy
operation that involves a glob.


## Variables

It is often useful to allow values to be stored in variables, for example if you're used to pulling a file from a remote host you might make the version of that release a variable.

Variables are defined with the `Set` primitive, which takes two arguments:

* The name of the variable.
* The value to set for that variable.
  * Values will be set as strings, in fact our mini-language only understands strings.

In the following example we declare the variable called "RELEASE" to have the value "1.2", and then use it in a command-execution:

    Set RELEASE "1.2"
    Run "wget -O /usr/local/bin/app-${RELEASE} \
           https://example.com/dist/app-${RELEASE}"

It is possible to override the value of a particular variable via a command-line argument, for example:

    $ deployr run --set "ENVIRONMENT=PRODUCTION" ...

If you do this any attempt to `Set` the variable inside the recipe itself will be silently ignored.  (i.e. A variable which is set on the command-line will become essentially read-only.) This is useful if you have a recipe where the only real difference is the set of configuration files, and the destination host. For example you could write all your copies like so:

    #
    # Lack of recursive copy is a pain here.
    # See:
    #   https://github.com/skx/deployr/issues/6
    #
    CopyFile files/${ENVIRONMENT}/etc/apache2.conf /etc/apache2/conf
    CopyFile files/${ENVIRONMENT}/etc/redis.conf   /etc/redis/redis.conf
    ..

Then have a tree of files:

      ├── files
          ├── development
          │   ├── apache2.conf
          │   └── redis.conf
          └── production
              ├── apache2.conf
              └── redis.conf

Another case where this come in handy is when dealing the secrets. Pass your secrets via command-line arguments instead of setting them in the recipe so you don't commit them by mistake, for example:

        $ deployr run --set "API_KEY=foobar" ...

Then use the `API_KEY`:

        Run "curl api.example.com/releases/latest -H 'Authorization: Bearer ${API_KEY}'"

In a CI environnement, use command-line arguments to retrieve environnement variables available in the CI.

        $ deployr run --set "RELEASE=$CI_COMMIT_TAG" ...

### Predefined Variables

The following variables are defined by default:

* `host`
  * The host being deployed to.
* `now`
  * An instance of the golang [time](https://golang.org/pkg/time/) object.
* `port`
  * The port used to connect to the remote host (22 by default).
* `user`
  * The username we login to the remote host as (root by default).



## Template Expansion

In addition to copying files literally from the local system to the remote
host it is also possible perform some limited template-expansion.

To copy a file literally you'd use the `CopyFile` primitive which copies the
file with no regards to the contents (handling binary content):

    CopyFile local.txt /tmp/remote.txt

To copy a file with template-expansion you should use the `CopyTemplate` primitive instead:

    CopyTemplate local.txt /tmp/remote.txt

The file being copied will then be processed with the `text/template` library
which means you can access values like so:

    #
    # This is a configuration file blah.conf
    # We can expand variables like so:
    #
    # Deployed version {{get "RELEASE"}} on Host:{{get "host"}}:{{get "port"}}
    # at {{now.UTC.Day}} {{now.UTC.Month}} {{now.UTC.Year}}
    #

In short you write `{{get "variable-name-here"}}` and the value of the variable
will be output inline.

Any variable defined with `Set` (or via a command-line argument) will be available to you, as well as the
[predefined variables](#predefined-variables) noted above.



## Missing Primitives?

If there are primitives you think would be useful to add then please do
[file a bug](http://github.com/skx/deployr/issues).


### Alternatives

There are many alternatives to this simple approach.  The most obvious two
would be:

* [ansible](https://www.ansible.com/)
  * Uses YAML to let you run commands on multiple remote hosts via SSH.
  * Very featureful, but also a bit hard to be readable due to the YAML use.
* [fabric](http://www.fabfile.org/)
  * Another Python-based project, which defines some simple primitive functions such as `run` and `put` to run commands, and upload files respectively.

As a very simple alternative I put together [marionette](https://github.com/skx/marionette/) which allows running commands, and setting file-content, but this works on the __local__ system only - no SSH involved.

For large-scale deployments you'll probably want to consider Puppet, Chef, or something more established and powerful.  Still this system has its place.



## Github Setup

This repository is configured to run tests upon every commit, and when
pull-requests are created/updated.  The testing is carried out via
[.github/run-tests.sh](.github/run-tests.sh) which is used by the
[github-action-tester](https://github.com/skx/github-action-tester) action.

Releases are automated in a similar fashion via [.github/build](.github/build),
and the [github-action-publish-binaries](https://github.com/skx/github-action-publish-binaries) action.

Steve
--
