// Package evaluator is the core of our run-time.
//
// Given a parsed series of statements we execute each of them in turn.
package evaluator

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/sfreiberg/simplessh"
	"github.com/skx/deployr/statement"
	"github.com/skx/deployr/util"
	"golang.org/x/crypto/ssh/terminal"
)

// Evaluator holds our internal state.
type Evaluator struct {

	// Program is our parsed program, which is an array of statements.
	Program []statement.Statement

	// Identity holds the SSH key to authenticate with
	Identity string

	// Verbose is true if the execution should be verbose.
	Verbose bool

	// NOP records if we should pretend to work, or work for real.
	NOP bool

	// Variables is a map which holds the names/values of all defined
	// variables.  (Being declared/set/updated via the 'Set' primitive.)
	Variables map[string]string

	// ROVariables is a map which is similar to Variables, but only
	// contains values set on the command-line.  Variables are looked
	// for here first.
	ROVariables map[string]string

	// Connection holds the SSH-connection to the remote-host.
	Connection *simplessh.Client

	// Changed records whether the last copy operaton resulted in a change.
	Changed bool
}

// New creates our evaluator object, which will execute the supplied
// statements.
func New(program []statement.Statement) *Evaluator {
	p := &Evaluator{Program: program}

	// Setup the maps for storing variable names & values.
	p.Variables = make(map[string]string)
	p.ROVariables = make(map[string]string)

	return p
}

// SetIdentity specifies the SSH identity file to authenticate with
func (e *Evaluator) SetIdentity(file string) {
	if file != "" {
		e.Identity = file
	} else {
		e.Identity = os.Getenv("HOME") + "/.ssh/id_rsa"
	}
}

// SetNOP specifies whether we should run for real, or not at all.
func (e *Evaluator) SetNOP(verb bool) {
	e.NOP = verb
}

// SetVerbose specifies whether we should run verbosely or not.
func (e *Evaluator) SetVerbose(verb bool) {
	e.Verbose = verb
}

// ConnectTo opens the SSH connection to the specified target-host.
//
// If a connection is already open then it is maintained, and not replaced.
// This allows the command-line to override the destination which might be
// baked into a configuration-recipe.
//
func (e *Evaluator) ConnectTo(target string) error {
	var err error

	if e.Connection != nil {
		fmt.Printf("Ignoring request to change target mid-run!\n")
		return nil
	}

	//
	// Default username + port
	//
	user := "root"
	port := "22"
	host := ""

	//
	// Setup the user if we have it
	//
	if strings.Contains(target, "@") {
		fields := strings.Split(target, "@")
		user = fields[0]
		host = fields[1]
	} else {
		host = target
	}

	//
	// Setup the port if we have it
	//
	if strings.Contains(host, ":") {
		fields := strings.Split(host, ":")
		host = fields[0]
		port = fields[1]
	}

	//
	// Store our connection-details in the variable-list
	//
	e.Variables["host"] = host
	e.Variables["port"] = port
	e.Variables["user"] = user

	//
	// Setup our destination with the host/port
	//
	destination := fmt.Sprintf("%s:%s", host, port)

	//
	// Finally connect.
	//
	if util.HasSSHAgent() {
		e.Connection, err = simplessh.ConnectWithAgent(destination, user)
	} else {
		e.Connection, err = simplessh.ConnectWithKeyFile(destination, user, e.Identity)
	}
	if err != nil {
		return err
	}

	return nil
}

// Run evaluates our program, continuing until all statements have been
// executed - unless an error was encountered.
func (e *Evaluator) Run() error {

	//
	// Do any of our program-statements require the use of Sudo?
	//
	sudo := false
	for _, statement := range e.Program {
		if statement.Sudo {
			sudo = true
		}
	}

	//
	// OK we need a sudo-password.  So prompt for it.
	//
	sudoPassword := ""
	if sudo {
		fmt.Printf("Please enter your password for sudo: ")

		text, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return err
		}
		fmt.Printf("\n")
		sudoPassword = string(text)
	}

	//
	// For each statement ..
	//
	for _, statement := range e.Program {

		//
		// The action to be taken will depend upon the type
		// of the token.
		//
		switch statement.Token.Type {

		case "CopyTemplate":

			//
			// Ensure we're connected.
			//
			if e.Connection == nil {
				return fmt.Errorf("tried to run a command, but not connected to a target")
			}

			//
			// Get the arguments and run the copy.
			//
			src := e.expandString(statement.Arguments[0].Literal)
			dst := e.expandString(statement.Arguments[1].Literal)
			if e.Verbose {
				fmt.Printf("CopyTemplate(\"%s\", \"%s\")\n", src, dst)
			}

			if e.NOP {
				break
			}
			e.Changed = e.copyFiles(src, dst, true)

		case "CopyFile":

			//
			// Ensure we're connected.
			//
			if e.Connection == nil {
				return fmt.Errorf("tried to run a command, but not connected to a target")
			}

			//
			// Get the arguments and run the copy.
			//
			src := e.expandString(statement.Arguments[0].Literal)
			dst := e.expandString(statement.Arguments[1].Literal)

			if e.Verbose {
				fmt.Printf("CopyFile(\"%s\", \"%s\")\n", src, dst)
			}

			if e.NOP {
				break
			}

			e.Changed = e.copyFiles(src, dst, false)

		case "DeployTo":

			//
			// Get the arguments, and connect.
			//
			arg := e.expandString(statement.Arguments[0].Literal)

			if e.Verbose {
				fmt.Printf("DeployTo(\"%s\")\n", arg)
			}

			err := e.ConnectTo(arg)
			if err != nil {
				return err
			}

		case "IfChanged":

			//
			// If the previous copy didn't change then we can
			// just skip this command.
			//
			if !e.Changed {
				break
			}

			//
			// Ensure we're connected.
			//
			if e.Connection == nil {
				return fmt.Errorf("tried to run a command, but not connected to a target")
			}

			//
			// Get the command to execute.
			//
			cmd := e.expandString(statement.Arguments[0].Literal)

			if e.Verbose {
				if statement.Sudo {
					fmt.Printf("Sudo ")
				}
				fmt.Printf("IfChanged(\"%s\")\n", cmd)
			}

			if e.NOP {
				break
			}

			//
			// Holder for results of execution.
			//
			var result []byte
			var err error

			//
			// Run via sudo or normally.
			//
			if statement.Sudo {
				result, err = e.Connection.ExecSudo(cmd, sudoPassword)
			} else {
				result, err = e.Connection.Exec(cmd)
			}
			if err != nil {
				return (fmt.Errorf("failed to run command '%s': %s\n%s", cmd, err.Error(), result))
			}

			//
			// Show the output
			//
			fmt.Printf("%s", result)

		case "Run":

			//
			// Ensure we're connected.
			//
			if e.Connection == nil {
				return fmt.Errorf("tried to run a command, but not connected to a target")
			}

			cmd := e.expandString(statement.Arguments[0].Literal)

			if e.Verbose {
				if statement.Sudo {
					fmt.Printf("Sudo ")
				}

				fmt.Printf("Run(\"%s\")\n", cmd)
			}

			if e.NOP {
				break
			}

			//
			// Holder for results of execution.
			//
			var result []byte
			var err error

			//
			// Run via sudo or normally.
			//
			if statement.Sudo {
				result, err = e.Connection.ExecSudo(cmd, sudoPassword)
			} else {
				result, err = e.Connection.Exec(cmd)
			}
			if err != nil {
				return (fmt.Errorf("failed to run command '%s': %s\n%s", cmd, err.Error(), result))
			}

			//
			// Show the output
			//
			fmt.Printf("%s", result)

		case "Set":

			//
			// Get the arguments and set the variable.
			//
			key := statement.Arguments[0].Literal
			val := e.expandString(statement.Arguments[1].Literal)

			if e.Verbose {
				fmt.Printf("Set(\"%s\", \"%s\")\n", key, val)
			}
			e.Variables[key] = val

		case "Sudo":

			//
			// This is an error?
			//
		default:
			return fmt.Errorf("unhandled statement - %v", statement.Token)
		}
	}

	//
	// Disconnect from the remote host, if we connected.
	//
	if e.Connection != nil {
		if e.Verbose {
			fmt.Printf("Disconnecting from remote-host\n")
		}
		e.Connection.Close()
	}

	//
	// All done.
	//
	return nil
}

// copyFiles is designed to copy a file/template from the local
// system to the remote host.
//
// It might be called with a glob, or with a single file.
//
func (e *Evaluator) copyFiles(pattern string, destination string, expand bool) bool {

	//
	// If our input pattern ends with a "/" we just add "*"
	//
	if strings.HasSuffix(pattern, "/") {
		pattern += "*"
	}

	//
	// Expand the pattern we received.
	//
	files, err := filepath.Glob(pattern)
	if err != nil {
		return false
	}

	//
	// Did we fail to find file(s)?
	//
	if len(files) < 1 {
		fmt.Printf("Failed to find file(s) matching %s\n", pattern)
		return false
	}

	//
	// Did we receive more than one file?
	//
	if len(files) == 1 && files[0] == pattern {

		//
		// OK just copying a single file.
		//
		return (e.copyFile(pattern, destination, expand))
	}

	//
	// OK we now have to copy each entry.  Since we're copying
	// from a pattern our destination will need to be updated ensure
	// we have a trailing "/" for that.
	//
	if !strings.HasSuffix(destination, "/") {
		destination += "/"
	}

	//
	// We record a change if we updated ANY of the files.
	//
	changed := false

	//
	// Now process each file.
	//
	for _, file := range files {

		fi, err := os.Stat(file)
		if err != nil {
			fmt.Printf("Failed to stat(%s) %s\n", file, err.Error())
			continue
		}
		switch mode := fi.Mode(); {
		case mode.IsDir():
			if e.Verbose {
				fmt.Printf("Skipping directory %s\n", file)
			}
		case mode.IsRegular():
			name := path.Base(file)
			c := e.copyFile(file, destination+name, expand)
			if c {
				changed = c
			}
		}
	}

	//
	return changed
}

// copyFile is designed to copy the local file to the remote system.
//
// It is a little complex because it does two extra things:
//
// * It only copies files if the local/remote differ.
//
// * It optionally expands template-variables.
//
func (e *Evaluator) copyFile(local string, remote string, expand bool) bool {

	//
	// Did we result in a change?
	//
	changed := false

	if e.Verbose {
		if expand {
			fmt.Printf("CopyTemplate(\"%s\",\"%s\")\n", local, remote)
		} else {
			fmt.Printf("CopyFile(\"%s\",\"%s\")\n", local, remote)
		}

	}
	//
	// If we're expanding templates then do that first of all.
	//
	// * Load the source file.
	//
	// * Perform the template-expansion of variables.
	//
	// * Write that expanded result to a temporary file.
	//
	// * Swap out the local-file name with the temporary-file.
	//
	if expand {

		//
		// Read the input file.
		//
		data, err := ioutil.ReadFile(local)

		//
		// If we can't read the input-file that's a fatal error.
		//
		if err != nil {
			fmt.Printf("Failed to read local file to expand template-variables %s\n", err.Error())
			os.Exit(11)
		}

		//
		// Define a helper-function that users can call to get
		// the variables they've set.
		//
		funcMap := template.FuncMap{
			"get": func(s string) string {
				if len(e.ROVariables[s]) > 0 {
					return (e.ROVariables[s])
				}
				return (e.Variables[s])
			},
			"now": time.Now,
		}

		//
		// Load the file as a template.
		//
		tmpl := template.Must(template.New("tmpl").Funcs(funcMap).Parse(string(data)))

		//
		// Now expand the template into a temporary-buffer.
		//
		buf := &bytes.Buffer{}
		tmpl.Execute(buf, e.Variables)

		//
		// Finally write that to a temporary file, and ensure
		// that is the source of the copy.
		//
		tmpfile, _ := ioutil.TempFile("", "tmpl")
		local = tmpfile.Name()
		ioutil.WriteFile(local, buf.Bytes(), 0600)
	}

	//
	// Copying a file to the remote host is
	// very simple - BUT we want to know if the
	// remote file changed, so we can make a
	// conditional result sometimes.
	//
	// So we need to hash the local file, and
	// the remote (if it exists) and compare
	// the two.
	//
	//
	// NOTE: We do this after we've expanded any variables.
	//
	var hashLocal string
	var err error
	hashLocal, err = util.HashFile(local)
	if err != nil {
		fmt.Printf("Failed to hash local file %s\n", err.Error())

		//
		// If we're trying to copy a file that doesn't exist that
		// is a fatal error.
		//
		os.Exit(11)
	}

	//
	// Now fetch the file from the remote host, if we can.
	//
	tmpfile, _ := ioutil.TempFile("", "example")
	defer os.Remove(tmpfile.Name()) // clean up

	err = e.Connection.Download(remote, tmpfile.Name())
	if err == nil {

		//
		// We had no error - so we now have the
		// remote file copied here.
		//
		var hashRemote string
		hashRemote, err = util.HashFile(tmpfile.Name())
		if err != nil {
			fmt.Printf("Failed to hash remote file %s\n", err.Error())

			// If expanding variables we replaced our
			// input-file with the temporary result of
			// expansion.
			if expand {
				os.Remove(local)
			}
			return changed
		}

		if hashRemote != hashLocal {
			if e.Verbose {
				fmt.Printf("\tFile on remote host needs replacing.\n")
			}

			changed = true
		} else {
			if e.Verbose {
				fmt.Printf("\tFile on remote host doesn't need to be changed.\n")
			}
		}
	} else {

		//
		// If we failed to find the file we
		// assume thati t doesn't exist
		//
		if strings.Contains(err.Error(), "not exist") {
			changed = true
		}
	}

	//
	// Upload the file, if it changed
	//
	if changed {
		err = e.Connection.Upload(local, remote)
		if err != nil {
			fmt.Printf("Failed to upload '%s' to '%s': %s\n", local, remote, err.Error())

			// If expanding variables we replaced our
			// input-file with the temporary result of
			// expansion.
			if expand {
				os.Remove(local)
			}

			return changed
		}
	}
	// If expanding variables we replaced our
	// input-file with the temporary result of
	// expansion.
	if expand {
		os.Remove(local)
	}

	return changed
}

// expandString expands tokens of the form "${blah}" into the
// value of the variable "blah".
func (e *Evaluator) expandString(in string) string {

	//
	// Expand any variables which have previously been
	// declared.
	//
	re := regexp.MustCompile(`\$\{([^\}]+)\}`)
	in = re.ReplaceAllStringFunc(in, func(in string) string {

		in = strings.TrimPrefix(in, "${")
		in = strings.TrimSuffix(in, "}")

		// Look for read-only variables first
		if len(e.ROVariables[in]) > 0 {
			return (e.ROVariables[in])
		}

		// Now look for normal-variable
		if len(e.Variables[in]) > 0 {
			return (e.Variables[in])
		}

		// Finally we found neither, just leave the
		// expansion alone.
		return "${" + in + "}"
	})

	return in
}

// SetVariable sets the content of a read-only variable
func (e *Evaluator) SetVariable(key string, val string) {
	e.ROVariables[key] = val
}
