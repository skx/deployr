// Package lexer contains a simple lexer for reading an input-string
// and converting it into a series of tokens.
package lexer

import (
	"errors"
	"fmt"

	"github.com/skx/deployr/token"
)

// Lexer is used as the lexer for our deployr "language".
type Lexer struct {
	position     int    //current character position
	readPosition int    //next character position
	ch           rune   //current character
	characters   []rune //rune slice of input string
}

// New a Lexer instance from string input.
func New(input string) *Lexer {
	l := &Lexer{characters: []rune(input)}
	l.readChar()
	return l
}

// Dump outputs the complete stream of tokens from the lexer,
// consuming all input as it does so.
func (l *Lexer) Dump() {
	for {
		tok := l.NextToken()
		fmt.Printf("%v\n", tok)
		if tok.Type == "EOF" {
			break
		}
	}
}

// read one forward character
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.characters) {
		l.ch = rune(0)
	} else {
		l.ch = l.characters[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

// NextToken to read next token, skipping the white space.
func (l *Lexer) NextToken() token.Token {

	var tok token.Token
	l.skipWhitespace()

	// skip shebang
	if l.ch == rune('#') && l.peekChar() == rune('!') && l.position == 0 {
		l.skipComment()
		return (l.NextToken())
	}

	// skip single-line comments
	if l.ch == rune('#') {
		l.skipComment()
		return (l.NextToken())
	}

	switch l.ch {
	case rune('"'):
		str, err := l.readString()

		if err == nil {
			tok.Type = token.STRING
			tok.Literal = str
		} else {
			tok.Type = token.ILLEGAL
			tok.Literal = err.Error()
		}
	case rune(0):
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		tok.Literal = l.readIdentifier()
		tok.Type = token.LookupIdentifier(tok.Literal)
		return tok
	}
	l.readChar()
	return tok
}

// read Identifier
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isIdentifier(l.ch) {
		l.readChar()
	}
	return string(l.characters[position:l.position])
}

// skip white space
func (l *Lexer) skipWhitespace() {
	for isWhitespace(l.ch) {
		l.readChar()
	}
}

// skip comment (until the end of the line).
func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != rune(0) {
		l.readChar()
	}
	l.skipWhitespace()
}

// read string
func (l *Lexer) readString() (string, error) {
	out := ""

	for {
		l.readChar()
		if l.ch == '"' {
			break
		}
		if l.ch == rune(0) {
			return "", errors.New("unterminated string")
		}

		//
		// Handle \n, \r, \t, \", etc.
		//
		if l.ch == '\\' {

			// Line ending with "\" + newline
			if l.peekChar() == '\n' {
				// consume the newline.
				l.readChar()
				continue
			}

			l.readChar()

			if l.ch == rune('n') {
				l.ch = '\n'
			}
			if l.ch == rune('r') {
				l.ch = '\r'
			}
			if l.ch == rune('t') {
				l.ch = '\t'
			}
			if l.ch == rune('"') {
				l.ch = '"'
			}
			if l.ch == rune('\\') {
				l.ch = '\\'
			}
		}
		out = out + string(l.ch)

	}

	return out, nil
}

// peek character
func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.characters) {
		return rune(0)
	}
	return l.characters[l.readPosition]
}

// determinate ch is identifier or not
func isIdentifier(ch rune) bool {
	return !isWhitespace(ch) && !isEmpty(ch)
}

// is white space
func isWhitespace(ch rune) bool {
	return ch == rune(' ') || ch == rune('\t') || ch == rune('\n') || ch == rune('\r')
}

// is empty
func isEmpty(ch rune) bool {
	return rune(0) == ch
}
