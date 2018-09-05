# Overview

When a file is loaded it is split up into a series of tokens via the
lexer:

* The token-types are basically "built-ins", "identifiers", or "strings".
  * These are defined in [token/token.go](token/token.go)

The lexer implementation may be found at [lexer/lexer.go](lexer/lexer.go).

To execute the recipe we _could_ just interpret each of the tokens in-turn,
but if we do that we'll not be able to catch syntax-errors until we hit them.


## The Lexing Process

With the lexer the following input:

    Set greeting "Hello, world!"

Becomes this series of tokens which is 100% valid:

* {Set Set}
* {IDENT greeting}
* {STRING Hello, world!}

We could imagine we could process a stream of tokens like this, taking different
action depending upon the type of the token & fetching successive arguments
as appropriately.  However consider the following broken program:

    Run "some command"
    Set greeting
    Run "another"

Our token-list would be:

* {Run Run}
* {STRING "some command"}
* {Set Set}
* {IDENT greeting}
* {Run Run}
* {STRING "another"}

Here we're missing a value to the Set-command, and we would only learn this
at the point we came to evaluate the `Set` command:

* Find `Set`.
* Fetch the next token.
  * Which we would assume is an ident (i.e. variable-name).
* Fetch the next token.
  * Which we assume would be a string (i.e. the value we store in the variable).

Oops!  The second-fetch would find `Run` instead.  The program would have to
abort, mid-execution.

> It is __horrid__ to abort a recipe half-way through, because we might set the remote host into a broken state.


## Error Detection

To catch errors in our input, before we execute it, we must consume our
stream of tokens and parse them for sanity and correctness.

The parsing process will build up a structure which we actually execute,
and that will ensure that we don't execute programs that are invalid
because the parsing process will catch that for us before we come to
run them.

Our parse is located beneath [parser/parser.go](parser/parser.go) and
builds up an array of statements to execute.  Since we don't allow
control-flow, looping, or other complex facilities we only have to parse
statements minimally - validating the type of arguments to the various
primitives.

(i.e. We don't need to define an AST, we can continue to use the token-types
that the lexer gave us.  I see no value in wrapping them any further, given
that we only deal with strings, built-ins, and idents.)

A statement will consist of just an action (read token) and a set of
optional arguments:

* The `Run`-command takes a single-string argument.
   * As does `IfChanged`.
* The `Set`-command takes a pair of arguments.
   * An identifier and a string.
* No command takes more than two arguments..
