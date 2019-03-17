package lexer

import (
	"testing"

	"github.com/skx/deployr/token"
)

// TestSomeStrings tests that the input of a pair of strings is tokenized
// appropriately.
func TestSomeStrings(t *testing.T) {
	input := `"Steve" "Kemp"`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.STRING, "Steve"},
		{token.STRING, "Kemp"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestEscape ensures that strings have escape-characters processed.
func TestStringEscape(t *testing.T) {
	input := `"Steve\n\r\\" "Kemp\n\t\n" "Inline \"quotes\"."`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.STRING, "Steve\n\r\\"},
		{token.STRING, "Kemp\n\t\n"},
		{token.STRING, "Inline \"quotes\"."},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestComments ensures that single-line comments work.
func TestComments(t *testing.T) {
	input := `# This is a comment
"Steve"
# This is another comment`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.STRING, "Steve"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestShebang skips the shebang
func TestShebang(t *testing.T) {
	input := `#!/usr/bin/env deployr
"Steve"
# This is another comment`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.STRING, "Steve"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestRun tests a simple run-statement.
func TestRun(t *testing.T) {
	input := `#!/usr/bin/env deployr
Run "Steve"
# This is another comment`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.RUN, "Run"},
		{token.STRING, "Steve"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestUnterminated string ensures that an unclosed-string is an error
func TestUnterminatedString(t *testing.T) {
	input := `#!/usr/bin/env deployr
Run "Steve`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.RUN, "Run"},
		{token.ILLEGAL, "unterminated string"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestContinue checks we continue newlines.
func TestContinue(t *testing.T) {
	input := `#!/usr/bin/env deployr
Run "This is a test \
which continues"
`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.RUN, "Run"},
		{token.STRING, "This is a test which continues"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestDump just calls dump on some tokens.
func TestDump(t *testing.T) {
	input := `#!/usr/bin/env deployr
# This is another comment`

	l := New(input)
	l.Dump()

	//
	// Since we've consumed all the input we expect we'll
	// read past the input here.
	//
	if l.peekChar() != rune(0) {
		t.Fatalf("We still have input, after dumping our stream")
	}
}
