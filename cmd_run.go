//
// Execute the recipe from the given file(s).
//

package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/google/subcommands"
	"github.com/skx/deployr/evaluator"
	"github.com/skx/deployr/lexer"
	"github.com/skx/deployr/parser"
	"github.com/skx/deployr/util"
)

// arrayFlags is the type of a flag that can be duplicated
type arrayFlags []string

// String returns a human-readable version of the flags-set
func (i *arrayFlags) String() string {
	return strings.Join(*i, ",")
}

// Set updates the value of this flag, by appending to the list.
func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

//
// runCmd holds the state for this sub-command.
//
type runCmd struct {
	// nop is true if we should pretend to run commands, not actually
	// run them for real.
	nop bool

	// identity holds the SSH identity file to use.
	identity string

	// target allows the target against which the recipe runs to be
	// set on the command-line.
	target string

	// vars stores any variables which are specified on the command-line.
	vars arrayFlags

	// verbose is true if we should be extra-verbose when running.
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
	f.BoolVar(&r.nop, "nop", false, "No operation - just pretend to run.")
	f.BoolVar(&r.verbose, "verbose", false, "Run verbosely.")
	f.StringVar(&r.identity, "identity", "", "The identity file to use for key-based authentication.")
	f.StringVar(&r.target, "target", "", "The target host to execute the recipe against.")
	f.Var(&r.vars, "set", "Set the value of a particular variable.  (May be repeated.)")
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
		err = e.ConnectTo(r.target)
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
	// Save the identity-flag - the default is ~/.ssh/id_rsa
	//
	e.SetIdentity(r.identity)

	//
	// Are there any variables set on the command-line?
	//
	re := regexp.MustCompile("^([^=]+)=(.*)$")
	for _, set := range r.vars {

		matches := re.FindStringSubmatch(set)
		if len(matches) == 3 {
			e.SetVariable(matches[1], matches[2])
		}
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
