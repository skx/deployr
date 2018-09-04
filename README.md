[![Travis CI](https://img.shields.io/travis/skx/deployr/master.svg?style=flat-square)](https://travis-ci.org/skx/deployr)
[![Go Report Card](https://goreportcard.com/badge/github.com/skx/deployr)](https://goreportcard.com/report/github.com/skx/deployr)
[![license](https://img.shields.io/github/license/skx/deployr.svg)](https://github.com/skx/deployr/blob/master/LICENSE)
[![Release](https://img.shields.io/github/release/skx/deployr.svg)](https://github.com/skx/deployr/releases/latest)


# deployr

`deployr` is a simple utility which is designed to allow you to easily
automate simple application-deployment via SSH.

The core idea behind `deployr` is that installing software upon remote hosts
frequently consists of a small number of steps:

* Uploading a file, or small number of files.
* Running a command, or two, to enable and start the software.

This is particularly true for golang-based applications which frequently consist
of an single binary, a single configuration file, and a systemd unit to
ensure the service is running.

If you want to keep your deployment recipes automatable, and reproducible,
then scripting them with a tool like this is ideal.


## Installation

Providing you have a working go-installation you should be able to install this software by running:

    go get -u  github.com/skx/deployr
    go install github.com/skx/deployr

> **NOTE**: If you've previously downloaded the code this will update your installation to the most recent available version.

If you don't have a golang environment setup you should be able to download a binary for GNU/Linux from [our release page](https://github.com/skx/deployr/releases).



## Overview

`deployr` is invoked by specifying the name of a file to process:


    $ deployr [options] recipe1 recipe2 .. recipeN

Each recipe file is read line-by-line, and the commands inside it interpreted.

The following snippet is a complete and valid example-file:

     #
     # Deploy Steve's puppet-summary application.
     #
     # The puppet-summary application is developed on github, and only
     # fixed releases are installed via release-page of the project:
     #
     #    https://github.com/skx/puppet-summary/releases
     #
     # The software is deployed to the single host `master.steve.org.uk`
     # and I connect to that host via SSH as the root-user.
     #
     # At the time of writing the most recent release is version 1.2
     #


     #
     # Specify the host to which we're deploying.
     #
     # Public-key authentication is both assumed and required.
     #
     DeployTo root@master.steve.org.uk

     #
     # Set the release version as a variable named "RELEASE".
     #
     Set RELEASE 1.2

     #
     # Now that RELEASE is defined it can be used as ${RELEASE} in the rest of
     # this file.
     #

     #
     # To deploy the software I need to fetch it.  Do that with `wget`:
     #
     Run wget --quiet -O /srv/puppet-summary/puppet-summary-linux-amd64-${RELEASE} \
        https://github.com/skx/puppet-summary/releases/download/release-${RELEASE}/puppet-summary-linux-amd64

     #
     # Create a symlink from this versioned download to an unqualified name.
     #
     Run ln -sf /srv/puppet-summary/puppet-summary-linux-amd64-${RELEASE} \
        /srv/puppet-summary/puppet-summary


     #
     # Ensure the downloaded file is executable.
     #
     Run chmod 755 /srv/puppet-summary/puppet*

     #
     # Finally we need to make sure there is a systemd unit-file in-place which
     # will start the service
     #
     CopyFile puppet-summary.service /lib/systemd/system/puppet-summary.service
     IfChanged systemctl daemon-reload
     IfChanged systemctl enable puppet-summary.service


     #
     # And now we should be able to stop/start the service
     #
     Run systemctl stop  puppet-summary.service
     Run systemctl start puppet-summary.service

With this example saved to `Recipe` I can now install, or update, the
application on that host via:

    $ deployr ./puppet.recipe

For more verbose output the `-verbose` flag can be specified:

    $ deployr -verbose ./Recipe



## Command-Line Flags

There are only two command-line flags supported in this initial release:

* `-verbose`
  * Operate verbosely
* `-target`
  * Specify the details of the system to deploy _to_.

You might prefer to specify the connection-details of your target inside
the recipe, as we did in the example we showed earlier.  If you don't
you could write a simple file like this:

    # Sample.Recipe
    Run touch /tmp/blah

Then execute it like this:

    $ deployr -target username@host.example.com ./Sample.Recipe


## Supported Features

The previous example showed a good overview of the facilities available.

The list of supported-primitives is:

* `CopyFile local/path remote/path`
  * Copy the specified local file to the specified path on the remote system.
  * If the local & remote files differ then this will be recorded.
* `CopyTemplate local/path remote/path`
  * Copy the specified local file to the specified path on the remote system, expanding variables prior to running the copy.
  * If the local & remote files differ then this will be recorded.
* `DeployTo [user@]hostname[:port]`
  * Specify the details of the host to connect to.
  * If you don't specify a target within your recipe itself you can instead pass them on the command-line via `-target [user@]hostname.example.com[:port]`.
* `IfChanged`
  * The `CopyFile`, & `CopyTemplate` primitive record whether they made a change to the remote system.
  * The `IfChanged` primitive will only execute the given command if the previous copy-operation resulted in the remote system being changed.
* `Run Command`
  * Run the given command (unconditionally) upon the remote-host.
* `Set name value`
  * Set the variable "name" to have the value "value".
  * Once set a variable can be used in the recipe, or as part of template-expansion.



## Template Expansion

In addition to copying files literally from the local system to the remote
host it is also possible perform some limited template-expansion.

In the previous example we saw that a variable was defined, and then used,
like so:

    Set RELEASE 1.2
    Run wget ... ${RELEASE}

Any variable which is set like this, (via `Set`), along with basic details
of the host being deployed to can be inserted into template-files which
are copied from the local-host to the remote system.

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


## COmm
Steve
--
