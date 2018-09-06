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

type Statement struct {
	Token     token.Token
	Arguments []token.Token
}
