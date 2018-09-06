# Parsing Overview

When a file is loaded it is split up into a series of tokens via the lexer:

* The token-types are basically "built-ins", "identifiers", or "strings".
  * These are defined in [token/token.go](token/token.go)

The lexer itself may be found in the file [lexer/lexer.go](lexer/lexer.go).

To execute the recipe we _could_ interpret each of the tokens the lexer
produced in-turn, because we have no loops, conditionals, or other
control-flow statements our program will always be executed from start to
finish.

However if we execute our tokens directly we will not be able to catch
syntax-errors until we reach them, and this would be unfortunate.

The following valid program:

    #!/usr/bin/env deployr
    # A comment
    Set greeting "Hello, world!"

Becomes this series of (valid) tokens:

* `{Set Set}`
* `{IDENT greeting}`
* `{STRING Hello, world!}`

Processing a stream of tokens like this, taking different action depending upon
the type of the token, and fetching successive arguments appropriately, would
work just fine for valid programs.

The following broken program shows what could go wrong though:

    Run "some command"
    Set greeting
    Run "another"

Our token-list would be:

* `{Run Run}`
* `{STRING "some command"}`
* `{Set Set}`
* `{IDENT greeting}`
* `{Run Run}`
* `{STRING "another"}`

Here we're missing a value to the Set-command, and we would only learn this
at the point we came to evaluate the `Set` command:

* Find `Run`.
  * Get the string-argument which is expected.
  * Execute the command.
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
stream of tokens entirely, then parse them for sanity and correctness into a
series of statements, which we'll execute in the future.

The parsing process will build up a structure which we'll actually execute,
and that will ensure that we don't execute programs that are invalid
because the parsing process will catch that for us before we come to
run them.

Our parser is located beneath [parser/parser.go](parser/parser.go) and
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
* No command takes more than two arguments.


## Actual Execution

Finally once we've parsed our lexed-tokens into a series of statements
with the parser we can execute them.

For completeness the evaluator is located in [evaluator/evaluator.go](evaluator/evaluator.go).
