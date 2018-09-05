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
