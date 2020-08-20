// Package parser is the package which parses our input.
//
// Given a lexer, wrapping a given input-file, we parse tokens from
// it into a series of statements which we then return for processing.
package parser

import (
	"fmt"

	"github.com/skx/deployr/statement"
	"github.com/skx/deployr/token"
)

// tokenizer is the interface for our tokenzier.
//
// We expect to consume tokens from our Lexer, but we use a layer
// of indirection by constructing our parser with an interface
// instead.
//
// This allows us to create a `FakeLexer` which satisfies the
// interface for testing-purposes.
//
type tokenizer interface {

	// Our tokenizer interface requires anything we
	// use to implement the NextToken() method - which
	// should return the next token in the stream.
	NextToken() token.Token
}

// Parser holds our internal state.
type Parser struct {
	// Our tokenizer.
	Tokenizer tokenizer
}

// New returns a new Parser object, consuming tokens from the specified
// tokenizer-interface.
func New(tk tokenizer) *Parser {
	l := &Parser{Tokenizer: tk}
	return l
}

// Parse the given program, catching errors.
func (p *Parser) Parse() ([]statement.Statement, error) {
	var result []statement.Statement

	//
	// Does the next command use Sudo?
	//
	sudo := false

	//
	// We have a lexer, so we process each token in-turn until we
	// hit the end-of-file.
	//
	run := true
	for run {

		//
		// Get the next token.
		//
		tok := p.Tokenizer.NextToken()

		//
		// Process each token-type appropriately.
		//
		// Basically we're performing validation here that there
		// are arguments of the appropriate type.
		//
		switch tok.Type {

		case "ILLEGAL":
			//
			// If we encounter an illegal-token that means
			// the lexer itself found something invalid.
			//
			// That might be a bogus number (if we supported numbers),
			// or an unterminated string.
			//
			return result, fmt.Errorf("error received from the lexer - %s", tok.Literal)
		case "IDENT":
			//
			// If we find a bare-ident which is not an argument
			// then we're either out of sync with reality or
			// the user has tried to run a bogus-program:
			//
			//   /usr/bin/id
			//   Run "/usr/bin/id"
			//
			// Either way this is an error.
			//
			return result, fmt.Errorf("found unexpected identifier '%s'", tok.Literal)
		case "STRING":
			//
			// If we find a bare-string which is not an argument
			// then we're either out of sync with reality or
			// the user has tried to run a bogus-program:
			//
			//   "Test"
			//   Run "/usr/bin/id"
			//
			// Either way this is an error.
			//
			return result, fmt.Errorf("found unexpected string '%s'", tok.Literal)
		case "CopyTemplate":
			//
			// We should have two arguments to CopyTemplate:
			//
			//  1. IDENT
			//  2. IDENT
			//
			// (Here IDENT means "path".)
			//
			expected := []token.Token{
				{Type: "IDENT"},
				{Type: "IDENT"},
			}

			//
			// Get the arguments, validating types.
			//
			args, err := p.GetArguments(expected)

			//
			// Error?
			//
			if err != nil {
				return result, err
			}

			//
			// Otherwise we can store this statement.
			//
			s := statement.Statement{Token: tok}
			s.Arguments = args
			result = append(result, s)

		case "CopyFile":

			//
			// We should have two arguments to CopyFile:
			//
			//  1. IDENT
			//  2. IDENT
			//
			// (Here IDENT means "path".)
			//
			expected := []token.Token{
				{Type: "IDENT"},
				{Type: "IDENT"},
			}

			//
			// Get the arguments, validating types.
			//
			args, err := p.GetArguments(expected)

			//
			// Error?
			//
			if err != nil {
				return result, err
			}

			//
			// Otherwise we can store this statement.
			//
			s := statement.Statement{Token: tok}
			s.Arguments = args
			result = append(result, s)

		case "DeployTo":
			//
			// We should have one arguments to DeployTo:
			//
			//  1. IDENT
			//
			expected := []token.Token{
				{Type: "IDENT"},
			}

			//
			// Get the arguments, validating types.
			//
			args, err := p.GetArguments(expected)

			//
			// Error?
			//
			if err != nil {
				return result, err
			}

			//
			// Otherwise we can store this statement.
			//
			s := statement.Statement{Token: tok}
			s.Arguments = args
			result = append(result, s)

		case "IfChanged":

			//
			// We should have one arguments to IfChanged:
			//
			//  1. String
			//
			expected := []token.Token{
				{Type: "STRING"},
			}

			//
			// Get the arguments, validating types.
			//
			args, err := p.GetArguments(expected)

			//
			// Error?
			//
			if err != nil {
				return result, err
			}

			//
			// Otherwise we can store this statement.
			//
			s := statement.Statement{Token: tok}
			s.Arguments = args

			//
			// Preserve the SUDO state
			//
			s.Sudo = sudo
			sudo = false

			result = append(result, s)

		case "Run":

			//
			// We should have one arguments to Run:
			//
			//  1. String
			//
			expected := []token.Token{
				{Type: "STRING"},
			}

			//
			// Get the arguments, validating types.
			//
			args, err := p.GetArguments(expected)

			//
			// Error?
			//
			if err != nil {
				return result, err
			}

			//
			// Otherwise we can store this statement.
			//
			s := statement.Statement{Token: tok}
			s.Arguments = args

			//
			// Preserve the SUDO state
			//
			s.Sudo = sudo
			sudo = false

			result = append(result, s)

		case "Set":

			//
			// We should have two arguments to set:
			//
			//  1. Ident.
			//  2. String
			//
			expected := []token.Token{
				{Type: "IDENT"},
				{Type: "STRING"},
			}

			//
			// Get the arguments, validating types.
			//
			args, err := p.GetArguments(expected)

			//
			// Error?
			//
			if err != nil {
				return result, err
			}

			//
			// Otherwise we can store this statement.
			//
			s := statement.Statement{Token: tok}
			s.Arguments = args
			result = append(result, s)

		case "Sudo":
			sudo = true

		case "EOF":

			//
			// This causes our parsing-loop to terminate.
			//
			run = false
		default:

			//
			// If we hit this point there is a token-type we
			// did not handle.
			//
			return nil, fmt.Errorf("unhandled statement - %v", tok)

		}
	}
	return result, nil
}

// GetArguments fetches arguments from the lexer, ensuring they're
// the expected types.
func (p *Parser) GetArguments(expected []token.Token) ([]token.Token, error) {
	var ret []token.Token

	for i, arg := range expected {

		next := p.Tokenizer.NextToken()
		if next.Type != arg.Type {
			return nil, fmt.Errorf("expected %v as argument %d - Got %v", arg.Type, i+1, next.Type)
		}

		ret = append(ret, next)

	}
	return ret, nil
}
