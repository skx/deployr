package token

// TokenType is a string
type TokenType string

// Token struct represent the lexer token
type Token struct {
	Type    TokenType
	Literal string
}

// pre-defined TokenTypes
const (
	EOF    = "EOF"
	IDENT  = "IDENT"
	STRING = "STRING"

    // Our keywords.
	COPY_FILE     = "CopyFile"
	COPY_TEMPLATE = "CopyTemplate"
	DEPLOY_TO     = "DeployTo"
	IF_CHANGED    = "IfChanged"
	RUN           = "Run"
	SET           = "Set"
)

// keywords holds our reversed keywords
var keywords = map[string]TokenType{
	"CopyFile":     COPY_FILE,
	"CopyTemplate": COPY_TEMPLATE,
	"DeployTo":     DEPLOY_TO,
	"IfChanged":    IF_CHANGED,
	"Run":          RUN,
	"Set":          SET,
}

// LookupIdentifier used to determinate whether identifier is keyword nor not
func LookupIdentifier(identifier string) TokenType {
	if tok, ok := keywords[identifier]; ok {
		return tok
	}
	return IDENT
}
