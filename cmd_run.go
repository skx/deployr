//
// Execute the recipe from the given file(s).
//

package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/google/subcommands"
	"github.com/skx/deployr/evaluator"
	"github.com/skx/deployr/lexer"
	"github.com/skx/deployr/parser"
	"github.com/skx/deployr/util"
)

//
// runCmd holds the state for this sub-command.
//
type runCmd struct {
	target  string
	verbose bool
	nop     bool
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
	f.BoolVar(&r.nop, "nop", false, "No operation - just pretend to run.")
	f.BoolVar(&r.verbose, "verbose", false, "Run verbosely.")
	f.StringVar(&r.target, "target", "", "The target host to execute the recipe against.")
}

//
// Run the given recipe
//
func (r *runCmd) Run(file string) {

	//
	// Read the contents of the file.
	//
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("Error reading file %s - %s\n", file, err.Error())
		return
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
		return
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
			return

		}
	}

	//
	// Set our flags verbosity-level
	//
	e.SetVerbose(r.verbose)
	if r.nop {
		e.SetVerbose(true)
		e.SetNOP(true)
	}

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

//
// Entry-point.
//
func (r *runCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	//
	// For each file we were given.
	//
	for _, file := range f.Args() {
		r.Run(file)
	}

	//
	// Fallback.
	//
	if len(f.Args()) < 1 {
		if util.FileExists("deploy.recipe") {
			r.Run("deploy.recipe")
		}
	}

	return subcommands.ExitSuccess
}
