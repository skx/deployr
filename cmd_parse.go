//
// Show the result of invoking our parser on the given input-file(s).
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
	"github.com/skx/deployr/parser"
)

//
// parseCmd is the structure for this sub-command.
//
type parseCmd struct {
}

//
// Glue
//
func (*parseCmd) Name() string     { return "parse" }
func (*parseCmd) Synopsis() string { return "Show our parser output." }
func (*parseCmd) Usage() string {
	return `parser :
  Show the output of running our parser on the given file(s).
`
}

//
// Flag setup
//
func (p *parseCmd) SetFlags(f *flag.FlagSet) {
}

//
// Entry-point.
//
func (p *parseCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	//
	// For each file we were given.
	//
	for _, file := range f.Args() {

		//
		// Read the contents of the file.
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
		// Create a parser, using the lexer.
		//
		p := parser.New(l)

		//
		// Parse the program, looking for errors.
		//
		statements, err := p.Parse()
		if err != nil {
			fmt.Printf("Error parsing program: %s\n", err.Error())
			return subcommands.ExitFailure
		}

		//
		// No errors?  Great.
		//
		// We can dump the parsed statements.
		//
		for _, statement := range statements {
			fmt.Printf("%v\n", statement)
		}
	}

	return subcommands.ExitSuccess
}
