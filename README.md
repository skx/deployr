[![Travis CI](https://img.shields.io/travis/skx/deployr/master.svg?style=flat-square)](https://travis-ci.org/skx/deployr)
[![Go Report Card](https://goreportcard.com/badge/github.com/skx/deployr)](https://goreportcard.com/report/github.com/skx/deployr)
[![license](https://img.shields.io/github/license/skx/deployr.svg)](https://github.com/skx/deployr/blob/master/LICENSE)
[![Release](https://img.shields.io/github/release/skx/deployr.svg)](https://github.com/skx/deployr/releases/latest)


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

One obvious facility that most similar systems, such as ansible, offer is the ability to perform looping operations, and comparisions.  We don't offer that and I'm not sure we ever will - even if we did add the ability to add cronjobs, etc.



## Installation

Providing you have a working go-installation you should be able to install this tool by running:

    go get -u  github.com/skx/deployr
    go install github.com/skx/deployr

> **NOTE**: If you've previously downloaded the code this will update your installation to the most recent available version.

If you don't have a golang environment setup you should be able to download a binary for GNU/Linux from [our release page](https://github.com/skx/deployr/releases).



## Overview

`deployr` has various sub-commands, the most useful is the `run` command which
allows you to execute a recipe-file:

    $ deployr run [options] recipe1 recipe2 .. recipeN

Each specified recipe is parsed and the primitives inside them are then executed line by line.  The following primitives/commands are available:

* `CopyFile local/path remote/path`
  * Copy the specified local file to the specified path on the remote system.
  * If the local & remote files were identical, such that no change was made, then this fact will be noted.
* `CopyTemplate local/path remote/path`
  * Copy the specified local file to the specified path on the remote system, expanding variables prior to running the copy.
  * If the local & remote files were identical, such that no change was made, then this fact will be noted.
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

**NOTE**: Previously the "Run" and "IfChanged" primitives took bare arguments, now they __must__ be quoted.  For example:

     Run "/usr/bin/id"
     IfChanged "/usr/bin/uptime"

There are several examples included beneath [examples/](examples/), the shortest one [examples/simple/](examples/simple/) is a particularly good recipe to examine to get a feel for the system:

    $ cd ./examples/simple/
    $ deployr run -target [user@]host.example.com[:port] ./deployr.recipe

For more verbose output the `-verbose` flag may be added:

    $ cd ./examples/simple/
    $ deployr run -target [user@]host.example.com[:port] -verbose ./deployr.recipe



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

If you do this any attempt to `Set` the variable inside the recipe itself will be silently ignored.  (i.e. A variable which is set on the command-line will become essentially read-only.)   This is useful if you have a recipe where the only real difference is the set of configuration files, and the destination host.  For example you could write all your copies like so:

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

Any variable defined with `Set` will be available to you, as well as the
[predefined variables](#predefined-variables) noted above.



## Missing Primitives?

If there are primitives you think would be useful to add then please do
[file a bug](http://github.com/skx/deployr/issues).


Steve
--
