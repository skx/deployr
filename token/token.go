// Package token contains the token-types which our lexer produces,
// and which our parser understands.
package token

// Type is a string
type Type string

// Token struct represent the lexer token
type Token struct {
	Type    Type
	Literal string
}

// pre-defined TokenTypes
const (
	EOF     = "EOF"
	IDENT   = "IDENT"
	ILLEGAL = "ILLEGAL"
	STRING  = "STRING"

	// Our keywords.
	COPYFILE     = "CopyFile"
	COPYTEMPLATE = "CopyTemplate"
	DEPLOYTO     = "DeployTo"
	IFCHANGED    = "IfChanged"
	RUN          = "Run"
	SET          = "Set"
	SUDO         = "Sudo"
)

// keywords holds our reversed keywords
var keywords = map[string]Type{
	"CopyFile":     COPYFILE,
	"CopyTemplate": COPYTEMPLATE,
	"DeployTo":     DEPLOYTO,
	"IfChanged":    IFCHANGED,
	"Run":          RUN,
	"Set":          SET,
	"Sudo":         SUDO,
}

// LookupIdentifier used to determinate whether identifier is keyword nor not
func LookupIdentifier(identifier string) Type {
	if tok, ok := keywords[identifier]; ok {
		return tok
	}
	return IDENT
}
