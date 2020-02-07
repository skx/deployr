// Package statement contains our statements.
//
// A statement is a parsed command from the recipe-file.
//
// A statement will be one of the fixed token-types we
// have defined and an array of "arguments".
//
// For example the "Run "blah"" would become a statement
// with token "Run" and argument "blah".
//
// We setup an array here, but the most arguments supported
// is two, for the CopyFile & CopyTemplate commands.
package statement

import (
	"github.com/skx/deployr/token"
)

// Statement holds a single statement to be executed.
type Statement struct {
	// Token is the main action "Set", "Run", etc.
	Token token.Token

	// When running a command `Run`, `IfChanged` should we use
	// sudo?
	Sudo bool

	// Arguments contains the arguments to the operation.
	Arguments []token.Token
}
