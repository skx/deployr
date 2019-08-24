//
// Test-cases for our parser.
//
// The parser is designed to consume tokens from our lexer so we have to
// fake-feed them in.  We do this via the `FakeLexer` helper.
//

package parser

import (
	"strings"
	"testing"

	"github.com/skx/deployr/token"
)

//
// FakeLexer is a fake Lexer.  D'oh.
//
type FakeLexer struct {
	Offset int
	Tokens []token.Token
}

//
// NewFakeLexer creates a fake lexer which will output the given tokens, in turn.
func NewFakeLexer(t []token.Token) *FakeLexer {
	l := &FakeLexer{Tokens: t, Offset: 0}
	return l
}

//
// Retrieve and return the next token from our list.
//
func (f *FakeLexer) NextToken() token.Token {
	t := f.Tokens[f.Offset]
	f.Offset++
	return t
}

// TestEOF just runs a basic sanity-check
func TestEOF(t *testing.T) {

	//
	// Create a fake-lexer holding a "NOP" program
	//
	toks := []token.Token{
		{Type: "EOF", Literal: "EOF"},
	}
	fl := NewFakeLexer(toks)

	//
	// Now parse into statements.
	//
	p := New(fl)

	program, err := p.Parse()
	if err != nil {
		t.Fatalf("Found unexpected error parsing: %s\n", err.Error())
	}

	if len(program) != 0 {
		t.Fatalf("Unexpected length\n")
	}
}

//
// testSingleArgument is a utility function which is designed to test
// that the given function handles both valid and invalid arguments
// correctly.
//
// For example "Run" expects a string, so we can test it like so:
//
// testSingleArgument( "Run", "STRING", "IDENT" )
//
// "STRING" is valid, "IDENT" is bogus.
//
func testSingleArgument(t *testing.T, tokenName token.Type, validType token.Type, bogusType token.Type) {

	//
	// The fake-program which should be valid
	//
	valid := []token.Token{
		{Type: tokenName, Literal: string(tokenName)},
		{Type: validType, Literal: "My argument here"},
		{Type: "EOF", Literal: "EOF"},
	}

	//
	// The fake-program which should be invalid
	//
	invalid := []token.Token{
		{Type: tokenName, Literal: string(tokenName)},
		{Type: bogusType, Literal: "My argument here"},
		{Type: "EOF", Literal: "EOF"},
	}

	// Parse the valid program
	flv := NewFakeLexer(valid)
	pv := New(flv)
	program, err := pv.Parse()

	// We expect to receive no error.
	if err != nil {
		t.Fatalf("Received error parsing: %s\n", err.Error())
	}

	//
	// We expect our statement to be "DeployTo" with the argument
	// pointing to example.com
	if program[0].Token.Type != tokenName {
		t.Fatalf("Unexpected statement-type : %s\n", program[0].Token.Type)
	}
	if len(program[0].Arguments) != 1 {
		t.Fatalf("Unexpected argument length - got %d\n", len(program[0].Arguments))
	}
	if program[0].Arguments[0].Literal != "My argument here" {
		t.Fatalf("Unexpected argument: %s\n", program[0].Arguments[0].Literal)
	}

	// Parse the invalid program
	fli := NewFakeLexer(invalid)
	pi := New(fli)
	_, err = pi.Parse()

	// We expect to receive an error.
	if err == nil {
		t.Fatalf("Expected to receive an error got none\n")
	}
	if !strings.Contains(err.Error(), "expected "+string(validType)) {
		t.Fatalf("We received an error, but not the correct one: %s\n", err.Error())
	}

}

// TestRun tests "Run" handling.
func TestRun(t *testing.T) {
	testSingleArgument(t, "Run", "STRING", "IDENT")
}

// TestDeployTo tests "DeployTo" handling.
func TestDeployTo(t *testing.T) {
	testSingleArgument(t, "DeployTo", "IDENT", "STRING")
}

// TestIfChanged tests "IfChanged" handling.
func TestIfChanged(t *testing.T) {
	testSingleArgument(t, "IfChanged", "STRING", "IDENT")
}

// TestCopy tests our two copy operations.
//
// We call first of all with two IDENTS, which is valid.  Then try two
// bogus versions calling with:
//   STRING IDENT
//   IDENT  STRING
// This should ensure that the argument testing is exercised.
func TestCopy(t *testing.T) {

	//
	// We'll repeat our tests with both "CopyFile" and
	// "CopyTemplate"
	//
	terms := []token.Type{"CopyFile", "CopyTemplate"}

	for _, term := range terms {

		//
		// A program which is valid.
		//
		valid := []token.Token{
			{Type: term, Literal: string(term)},
			{Type: "IDENT", Literal: "/path/to/src"},
			{Type: "IDENT", Literal: "/path/to/dst"},
			{Type: "EOF", Literal: "EOF"},
		}

		//
		// Parse the program, we expect no errors
		//
		flv := NewFakeLexer(valid)
		pv := New(flv)
		program, err := pv.Parse()

		if err != nil {
			t.Fatalf("Received unexpected error parsing: %s\n", err.Error())
		}
		if len(program) != 1 {
			t.Fatalf("Our program should have one statement - found %d\n", len(program))
		}
		if len(program[0].Arguments) != 2 {
			t.Fatalf("Our statement should have two arguments - found %d\n", len(program[0].Arguments))
		}

		//
		// Now test an invalid program.
		//
		bogus1 := []token.Token{
			{Type: term, Literal: string(term)},
			{Type: "STRING", Literal: "/path/to/src"},
			{Type: "IDENT", Literal: "/path/to/dst"},
			{Type: "EOF", Literal: "EOF"},
		}

		//
		// Parse the program, we expect no errors
		//
		flb1 := NewFakeLexer(bogus1)
		pb1 := New(flb1)
		_, err = pb1.Parse()

		if err == nil {
			t.Fatalf("Expected to receive an error, got none")
		}

		if !strings.Contains(err.Error(), "as argument 1") {
			t.Fatalf("Our error was misleading: %s", err.Error())
		}

		//
		// Test another invalid program.
		//
		bogus2 := []token.Token{
			{Type: term, Literal: string(term)},
			{Type: "IDENT", Literal: "/path/to/src"},
			{Type: "STRING", Literal: "/path/to/dst"},
			{Type: "EOF", Literal: "EOF"},
		}

		//
		// Parse the program, we expect no errors
		//
		flb2 := NewFakeLexer(bogus2)
		pb2 := New(flb2)
		_, err = pb2.Parse()

		if err == nil {
			t.Fatalf("Expected to receive an error, got none")
		}

		if !strings.Contains(err.Error(), "as argument 2") {
			t.Fatalf("Our error was misleading %s", err.Error())
		}

	}
}

// TestBareString tests our error-handling.
func TestBareString(t *testing.T) {

	//
	// The stream of tokens we'll parse.
	//
	toks := []token.Token{
		{Type: "STRING", Literal: "/bin/ls"},
		{Type: "EOF", Literal: "EOF"},
	}

	//
	// Now parse into statements.
	//
	fl := NewFakeLexer(toks)
	p := New(fl)
	program, err := p.Parse()

	//
	// We expect to receive an error & empty-program
	//
	if err == nil {
		t.Fatalf("We expected an error, but saw none!")
	}
	if len(program) != 0 {
		t.Fatalf("Unexpected length, wanted 0 got %d\n", len(program))
	}
}

// TestBareIdentifier tests our error-handling.
func TestBareIdentifier(t *testing.T) {

	//
	// The stream of tokens we'll parse.
	//
	toks := []token.Token{
		{Type: "IDENT", Literal: "/bin/ls"},
		{Type: "EOF", Literal: "EOF"},
	}

	//
	// Now parse into statements.
	//
	fl := NewFakeLexer(toks)
	p := New(fl)
	program, err := p.Parse()

	//
	// We expect to receive an error & empty-program
	//
	if err == nil {
		t.Fatalf("We expected an error, but saw none!")
	}
	if len(program) != 0 {
		t.Fatalf("Unexpected length, wanted 0 got %d\n", len(program))
	}
}

// TestDefault tests our unhandled-token setting.
func TestDefault(t *testing.T) {

	//
	// The stream of tokens we'll parse.
	//
	toks := []token.Token{
		{Type: "MOI", Literal: "KISSA"},
		{Type: "EOF", Literal: "EOF"},
	}

	//
	// Now parse into statements.
	//
	fl := NewFakeLexer(toks)
	p := New(fl)
	_, err := p.Parse()

	//
	// We expect one statement, with zero errors.
	//
	if err == nil {
		t.Fatalf("We expected an error, but saw none!")
	}
	if !strings.Contains(err.Error(), "unhandled") {
		t.Fatalf("We received an error, but got the wrong one: %s\n", err.Error())
	}
}

// TestIllegal tests our error-handling.
func TestIllegal(t *testing.T) {

	//
	// The stream of tokens we'll parse.
	//
	toks := []token.Token{
		{Type: "ILLEGAL", Literal: "I like cake"},
		{Type: "EOF", Literal: "EOF"},
	}

	//
	// Now parse into statements.
	//
	fl := NewFakeLexer(toks)
	p := New(fl)
	program, err := p.Parse()

	//
	// We expect one statement, with zero errors.
	//
	if err == nil {
		t.Fatalf("We expected an error, but saw none!")
	}
	if len(program) != 0 {
		t.Fatalf("Unexpected length, wanted 0 got %d\n", len(program))
	}
	if !strings.Contains(err.Error(), "I like cake") {
		t.Fatalf("Our error didn't contain the message we set: %s\n", err.Error())
	}
}

// TestSet tests our variable-setting primitive.
//
// We call first of all with IDENT + "String", which is valid.  Then try two
// bogus versions calling with:
//   IDENT IDENT
//   STRING STRING
// This should ensure that the argument testing is exercised.
func TestSet(t *testing.T) {

	//
	// A program which is valid.
	//
	valid := []token.Token{
		{Type: "Set", Literal: "Set"},
		{Type: "IDENT", Literal: "RELESE"},
		{Type: "STRING", Literal: "1.2"},
		{Type: "EOF", Literal: "EOF"},
	}

	//
	// Parse the program, we expect no errors
	//
	flv := NewFakeLexer(valid)
	pv := New(flv)
	program, err := pv.Parse()

	if err != nil {
		t.Fatalf("Received unexpected error parsing: %s\n", err.Error())
	}
	if len(program) != 1 {
		t.Fatalf("Our program should have one statement - found %d\n", len(program))
	}
	if len(program[0].Arguments) != 2 {
		t.Fatalf("Our statement should have two arguments - found %d\n", len(program[0].Arguments))
	}

	//
	// Now test an invalid program.
	//
	bogus1 := []token.Token{
		{Type: "Set", Literal: "Set"},
		{Type: "STRING", Literal: "RELEASE"},
		{Type: "STRING", Literal: "1.2"},
		{Type: "EOF", Literal: "EOF"},
	}

	//
	// Parse the program, we expect no errors
	//
	flb1 := NewFakeLexer(bogus1)
	pb1 := New(flb1)
	_, err = pb1.Parse()

	if err == nil {
		t.Fatalf("Expected to receive an error, got none")
	}

	if !strings.Contains(err.Error(), "as argument 1") {
		t.Fatalf("Our error was misleading:%s", err.Error())
	}

	//
	// Test another invalid program.
	//
	bogus2 := []token.Token{
		{Type: "Set", Literal: "Set"},
		{Type: "IDENT", Literal: "RELEASE"},
		{Type: "IDENT", Literal: "3.14"},
		{Type: "EOF", Literal: "EOF"},
	}

	//
	// Parse the program, we expect no errors
	//
	flb2 := NewFakeLexer(bogus2)
	pb2 := New(flb2)
	_, err = pb2.Parse()

	if err == nil {
		t.Fatalf("Expected to receive an error, got none")
	}

	if !strings.Contains(err.Error(), "as argument 2") {
		t.Fatalf("Our error was misleading: %s", err.Error())
	}
}

// TestSudoHandling tests that we have sudo-handing correct.
func TestSudoHandling(t *testing.T) {

	//
	// The stream of tokens we'll parse.
	//
	toks := []token.Token{
		{Type: "Sudo", Literal: "Sudo"},
		{Type: "EOF", Literal: "EOF"},
	}

	//
	// Now parse into statements.
	//
	fl := NewFakeLexer(toks)
	p := New(fl)
	program, err := p.Parse()

	//
	// We expect one statement, with zero errors.
	//
	if err != nil {
		t.Fatalf("Received an unexpected error!")
	}
	if len(program) != 0 {
		t.Fatalf("Unexpected length, wanted 0 got %d\n", len(program))
	}
}

// TestSudoFlag tests that we actually set the flag for a sudo-using
// command.
func TestSudoFlag(t *testing.T) {

	//
	// The stream of tokens we'll expecting a sudo-flag to be set.
	//
	set := []token.Token{
		{Type: "Sudo", Literal: "Sudo"},
		{Type: "Run", Literal: "Run"},
		{Type: "STRING", Literal: "/bin/ls"},
		{Type: "EOF", Literal: "EOF"},
	}

	//
	// The stream of tokens we'll parse expecting no sudo-flag to be
	// set.
	//
	unset := []token.Token{
		{Type: "Run", Literal: "Run"},
		{Type: "STRING", Literal: "/bin/ls"},
		{Type: "EOF", Literal: "EOF"},
	}

	//
	// Now parse into statements.
	//
	fl := NewFakeLexer(set)
	p := New(fl)
	program, err := p.Parse()

	//
	// We expect one statement, with zero errors.
	//
	if err != nil {
		t.Fatalf("Received an unexpected error!")
	}
	if len(program) != 1 {
		t.Fatalf("Unexpected length, wanted 1 got %d\n", len(program))
	}

	//
	// The statement will use sudo
	//

	if program[0].Sudo != true {
		t.Fatalf("We expected our Run command to use sudo %v", program[0])
	}

	//
	// Now parse into statements.
	//
	fl = NewFakeLexer(unset)
	p = New(fl)
	program, err = p.Parse()

	//
	// We expect one statement, with zero errors.
	//
	if err != nil {
		t.Fatalf("Received an unexpected error!")
	}
	if len(program) != 1 {
		t.Fatalf("Unexpected length, wanted 1 got %d\n", len(program))
	}

	//
	// The statement will not use sudo
	//
	if program[0].Sudo != false {
		t.Fatalf("We didn't expect our Run command to use sudo %v", program[0])
	}
}
