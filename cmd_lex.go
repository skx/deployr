//
// Invoke our lexer on the specified input-file(s) and dump the output.
//

package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/google/subcommands"
	"github.com/skx/deployr/lexer"
)

//
// lexCmd is the structure for this sub-command.
//
type lexCmd struct {
}

//
// Glue
//
func (*lexCmd) Name() string     { return "lex" }
func (*lexCmd) Synopsis() string { return "Show our lexer output." }
func (*lexCmd) Usage() string {
	return `lex :
  Show the output of running our lexer on the given file(s).
`
}

//
// Flag setup
//
func (p *lexCmd) SetFlags(f *flag.FlagSet) {
}

//
// Entry-point.
//
func (p *lexCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	//
	// For each file we've been passed.
	//
	for _, file := range f.Args() {

		//
		// Read the file contents.
		//
		dat, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Printf("Error reading file %s - %s\n", file, err.Error())
			os.Exit(1)
		}

		//
		// Create a lexer object with those contents.
		//
		l := lexer.New(string(dat))

		//
		// Dump the tokens.
		//
		l.Dump()
	}

	return subcommands.ExitSuccess
}
