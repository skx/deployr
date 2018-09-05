//
// Given a lexer, wrapping a given input-file, we parse tokens from
// it into a series of statements.
//

package parser

import (
	"fmt"

	"github.com/skx/deployr/lexer"
	"github.com/skx/deployr/statement"
)

// Parser holds our internal state.
type Parser struct {
	// Lexer is our lexer.
	Lexer *lexer.Lexer
}

// New returns a new Parser object, consuming tokens from the specified
// lexer.
func New(lexer *lexer.Lexer) *Parser {
	l := &Parser{Lexer: lexer}
	return l
}

// Parse the given program, catching errors.
func (p *Parser) Parse() ([]statement.Statement, error) {
	var result []statement.Statement

	//
	// We have a lexer, so we process each token in-turn until we
	// hit the end-of-file.
	//
	run := true
	for run {

		//
		// Get the next token.
		//
		tok := p.Lexer.NextToken()

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
			return result, fmt.Errorf("Error retrieceved from the lexer - %s\n", tok.Literal)
		case "CopyTemplate":
			//
			// We should have two arguments to CopyTemplate:
			//
			//  1. IDENT
			//  2. IDENT
			//
			// (Here IDENT means "path".)
			//
			str1 := p.Lexer.NextToken()
			if str1.Type != "IDENT" {
				return nil, fmt.Errorf("Expected IDENT as first argument to 'CopyTemplate' - Got %v", str1)
			}

			str2 := p.Lexer.NextToken()
			if str2.Type != "IDENT" {
				return nil, fmt.Errorf("Expected IDENT as second argument to 'CopyTemplate' - Got %v", str2)
			}

			//
			// Now we can store this statement
			//
			s := statement.Statement{Token: tok}
			s.Arguments = append(s.Arguments, str1)
			s.Arguments = append(s.Arguments, str2)
			result = append(result, s)
			break

		case "CopyFile":

			//
			// We should have two arguments to CopyFile:
			//
			//  1. IDENT
			//  2. IDENT
			//
			// (Here IDENT means "path".)
			//
			str1 := p.Lexer.NextToken()
			if str1.Type != "IDENT" {
				return nil, fmt.Errorf("Expected IDENT as first argument to 'CopyFile' - Got %v", str1)
			}

			str2 := p.Lexer.NextToken()
			if str2.Type != "IDENT" {
				return nil, fmt.Errorf("Expected IDENT as second argument to 'CopyFile' - Got %v", str2)
			}

			//
			// Now we can store this statement
			//
			s := statement.Statement{Token: tok}
			s.Arguments = append(s.Arguments, str1)
			s.Arguments = append(s.Arguments, str2)
			result = append(result, s)
			break

		case "DeployTo":
			//
			// We should have one arguments to DeployTo:
			//
			//  1. IDENT
			//
			str := p.Lexer.NextToken()
			if str.Type != "IDENT" {
				return nil, fmt.Errorf("Expected IDENT as first argument to 'DeployTo' - Got %v", str)
			}

			//
			// Now we can store this statement
			//
			s := statement.Statement{Token: tok}
			s.Arguments = append(s.Arguments, str)
			result = append(result, s)
			break

		case "IfChanged":

			//
			// We should have one arguments to IfChanged:
			//
			//  1. String
			//
			str := p.Lexer.NextToken()
			if str.Type != "STRING" {
				return nil, fmt.Errorf("Expected STRING as first argument to 'IfChanged' - Got %v", str)
			}

			//
			// Now we can store this statement
			//
			s := statement.Statement{Token: tok}
			s.Arguments = append(s.Arguments, str)
			result = append(result, s)
			break

		case "Run":

			//
			// We should have one arguments to Run:
			//
			//  1. String
			//
			str := p.Lexer.NextToken()
			if str.Type != "STRING" {
				return nil, fmt.Errorf("Expected STRING as first argument to 'Run' - Got %v", str)
			}

			//
			// Now we can store this statement
			//
			s := statement.Statement{Token: tok}
			s.Arguments = append(s.Arguments, str)
			result = append(result, s)
			break

		case "Set":

			//
			// We should have two arguments to set:
			//
			//  1. Ident.
			//  2. String
			//
			id := p.Lexer.NextToken()
			if id.Type != "IDENT" {
				return nil, fmt.Errorf("Expected IDENT as first argument to Set - Got %v", id)
			}
			str := p.Lexer.NextToken()
			if str.Type != "STRING" {
				return nil, fmt.Errorf("Expected STRING as second argument to Set - Got %v", str)
			}

			//
			// Now we can store this statement
			//
			s := statement.Statement{Token: tok}
			s.Arguments = append(s.Arguments, id)
			s.Arguments = append(s.Arguments, str)
			result = append(result, s)

		case "EOF":

			//
			// This causes our parsing-loop to terminate.
			//
			run = false
			break
		default:

			//
			// If we hit this point there is a token-type we
			// did not handle.
			//
			return nil, fmt.Errorf("Unhandled statement - %v\n", tok)
			break
		}
	}
	return result, nil
}
