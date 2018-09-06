//
// Execute the recipe from the given file(s).
//

package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/google/subcommands"
	"github.com/skx/deployr/evaluator"
	"github.com/skx/deployr/lexer"
	"github.com/skx/deployr/parser"
)

//
// runCmd holds the state for this sub-command.
//
type runCmd struct {
	target  string
	verbose bool
}

//
// Glue
//
func (*runCmd) Name() string     { return "run" }
func (*runCmd) Synopsis() string { return "Run the specified recipe(s)." }
func (*runCmd) Usage() string {
	return `run :
  Load and execute the recipe in the specified file(s).
`
}

//
// Flag setup
//
func (r *runCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&r.verbose, "verbose", false, "Run verbosely.")
	f.StringVar(&r.target, "target", "", "The target host to execute the recipe against.")
}

//
// Entry-point.
//
func (r *runCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

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
		// Create the evaluator - which will run the statements.
		//
		e := evaluator.New(statements)

		//
		// Set the target, if we've been given one.
		//
		if r.target != "" {
			err := e.ConnectTo(r.target)
			if err != nil {
				fmt.Printf("Failed to connect to target: %s\n", err.Error())
				return subcommands.ExitFailure

			}
		}

		//
		// Set the verbosity-level
		//
		e.SetVerbose(r.verbose)

		//
		// Now run the program.  Hurrah!
		//
		err = e.Run()

		//
		// Errors?  Boo!
		//
		if err != nil {
			fmt.Printf("Error running program\n%s\n", err.Error())
		}
	}

	return subcommands.ExitSuccess
}
