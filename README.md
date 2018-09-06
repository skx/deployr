[![Travis CI](https://img.shields.io/travis/skx/deployr/master.svg?style=flat-square)](https://travis-ci.org/skx/deployr)
[![Go Report Card](https://goreportcard.com/badge/github.com/skx/deployr)](https://goreportcard.com/report/github.com/skx/deployr)
[![license](https://img.shields.io/github/license/skx/deployr.svg)](https://github.com/skx/deployr/blob/master/LICENSE)
[![Release](https://img.shields.io/github/release/skx/deployr.svg)](https://github.com/skx/deployr/releases/latest)


# deployr

`deployr` is a simple utility which is designed to allow you to easily automate simple application-deployment via SSH.

The core idea behind `deployr` is that installing software upon remote hosts frequently consists of a small number of steps:

* Uploading a file, or small number of files.
* Running a command, or two, to enable and start the software.

This is particularly true for golang-based applications which frequently consist of an single binary, a single configuration file, and a systemd unit to ensure the service is running.

If you want to keep your deployment recipes automatable, and reproducible, then scripting them with a tool like this is ideal.  (Though you might prefer something more popular & featureful such as ansible, fabric, salt, etc.)


## Installation

Providing you have a working go-installation you should be able to install this software by running:

    go get -u  github.com/skx/deployr
    go install github.com/skx/deployr

> **NOTE**: If you've previously downloaded the code this will update your installation to the most recent available version.

If you don't have a golang environment setup you should be able to download a binary for GNU/Linux from [our release page](https://github.com/skx/deployr/releases).



## Overview

`deployr` has various sub-commands, the most useful is the `run` command which
allows you to execute a recipe-file:

    $ deployr run [options] recipe1 recipe2 .. recipeN

Each specified recipe is processed line-by-line, and the commands inside them are interpreted.

The following primitives/commands are available:

* `CopyFile local/path remote/path`
  * Copy the specified local file to the specified path on the remote system.
  * If the local & remote files were identical, such that no change was made, then this will be noted.
* `CopyTemplate local/path remote/path`
  * Copy the specified local file to the specified path on the remote system, expanding variables prior to running the copy.
  * If the local & remote files were identical, such that no change was made, then this will be noted.
* `DeployTo [user@]hostname[:port]`
  * Specify the details of the host to connect to, this is useful if a particular recipe should only be applied against a single host.
  * If you don't specify a target within your recipe itself you can instead pass it upon the command-line via the `-target` flag.
* `IfChanged "Command"`
  * The `CopyFile` and `CopyTemplate` commands record whether they made a change to the remote system.
  * The `IfChanged` primitive will execute the specified command if the previous copy-operation resulted in the remote system being changed.
* `Run "Command"`
  * Run the given command (unconditionally) upon the remote-host.
* `Set name "value"`
  * Set the variable "name" to have the value "value".
  * Once set a variable can be used in the recipe, or as part of template-expansion.

**NOTE**: Previously the "Run" and "IfChanged" primitives took bare arguments, now they __must__ be quoted.  For example:

     Run "/usr/bin/id"
     IfChanged "/usr/bin/uptime"

The included [example.recipe](example.recipe) demonstrates some of these commands, and can be launched like so:

    $ deployr run -target [user@]host.example.com[:port] ./example.recipe

For more verbose output the `-verbose` flag may be added:

    $ deployr run -target [user@]host.example.com[:port] -verbose ./example.recipe


## Template Expansion

In addition to copying files literally from the local system to the remote
host it is also possible perform some limited template-expansion.

We could declare a variable `RELEASE` to contain the version-number of a release we're pulling from a remote-host for example:

    Set RELEASE "1.2"
    Run "wget -O /usr/local/bin/app-${RELESAE} https://example.com/dist/app-${RELEASE}"

Any variable which is set like this, along with basic details of the host being
deployed to can be inserted into template-files which are copied from the local-host to the remote system.

To copy a file literally you'd write:

    CopyFile local.txt /tmp/remote.txt

To copy a file with template-expansion performed upon its contents you'd write:

    CopyTemplate local.txt /tmp/remote.txt

The file being copied will be processed with the `text/template` library
which means you can access values like so:

    #
    # This is a configuration file blah.conf
    # We can expand variables like so:
    #
    # Deployed version {{get "RELEASE"}} on Host:{{get "host"}}:{{get "port"}}
    # at {{now.UTC.Day}} {{now.UTC.Month}} {{now.UTC.Year}}
    #

In short you write `{{get "variable-name-here}}` and this will be replaced
when the file is uploaded.


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


## Missing Primitives?

If there are primitives you think would be useful to add then please do
[file a bug](http://github.com/skx/deployr/issues).


Steve
--
