//
// Evaluator is the core of our run-time.
//
// Given a parsed series of statements we execute each of them in turn.
//

package evaluator

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/sfreiberg/simplessh"
	"github.com/skx/deployr/statement"
)

// Evaluator holds our internal state.
type Evaluator struct {

	// Program is our parsed program, which is an array of statements.
	Program []statement.Statement

	// Verbose is true if the execution should be verbose.
	Verbose bool

	// Variables is a map which holds the names/values of all defined
	// variables.  (Being declared/set/updated via the 'Set' primitive.)
	Variables map[string]string

	// Connection holds the SSH-connection to the remote-host.
	Connection *simplessh.Client

	// Changed records whether the last copy operaton resulted in a change.
	Changed bool
}

// New creates our evaluator object, which will execute the supplied
// statements.
func New(program []statement.Statement) *Evaluator {
	p := &Evaluator{Program: program}

	// Setup the map for storing variable names/values.
	p.Variables = make(map[string]string)

	return p
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

	if e.Verbose {
		fmt.Printf("Connecting to %s\n", target)
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
	e.Connection, err = simplessh.ConnectWithKeyFile(destination, user, os.Getenv("HOME")+"/.ssh/id_rsa")
	if err != nil {
		return err
	}

	return nil
}

// Run evaluates our program, continuing until all statements have been
// executed - unless an error was encountered.
func (e *Evaluator) Run() error {

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
				return fmt.Errorf("Tried to run a command, but not connected to a target!")
			}

			//
			// Get the arguments and run the copy.
			//
			src := statement.Arguments[0].Literal
			dst := statement.Arguments[1].Literal
			e.copyFile(src, dst, true)
			break

		case "CopyFile":

			//
			// Ensure we're connected.
			//
			if e.Connection == nil {
				return fmt.Errorf("Tried to run a command, but not connected to a target!")
			}

			//
			// Get the arguments and run the copy.
			//
			src := statement.Arguments[0].Literal
			dst := statement.Arguments[1].Literal
			e.copyFile(src, dst, false)
			break

		case "DeployTo":

			//
			// Get the arguments, and connect.
			//
			err := e.ConnectTo(statement.Arguments[0].Literal)
			if err != nil {
				return err
			}

			break

		case "IfChanged":

			//
			// If the previous copy didn't change then we can
			// just skip this command.
			//
			if e.Changed == false {
				break
			}

			//
			// Ensure we're connected.
			//
			if e.Connection == nil {
				return fmt.Errorf("Tried to run a command, but not connected to a target!")
			}

			//
			// Get the command to execute.
			//
			cmd := statement.Arguments[0].Literal
			result, err := e.Connection.Exec(cmd)
			if err != nil {
				return (fmt.Errorf("Failed to run command '%s': %s\n", cmd, err.Error()))
			}

			//
			// Show the output
			//
			fmt.Printf("%s", result)
			break

		case "Run":

			//
			// Ensure we're connected.
			//
			if e.Connection == nil {
				return fmt.Errorf("Tried to run a command, but not connected to a target!")
			}

			cmd := statement.Arguments[0].Literal
			result, err := e.Connection.Exec(cmd)
			if err != nil {
				return (fmt.Errorf("Failed to run command '%s': %s\n", cmd, err.Error()))
			}

			//
			// Show the output
			//
			fmt.Printf("%s", result)

			break

		case "Set":

			//
			// Get the arguments and set the variable.
			//
			key := statement.Arguments[0].Literal
			val := statement.Arguments[1].Literal
			e.Variables[key] = val

			break

		default:
			return fmt.Errorf("Unhandled statement - %v\n", statement.Token)
			break
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

//
// hashFile returns the SHA1-hash of the contents of the specified file.
//
func (e *Evaluator) hashFile(filePath string) (string, error) {
	var returnSHA1String string

	file, err := os.Open(filePath)
	if err != nil {
		return returnSHA1String, err
	}

	defer file.Close()

	hash := sha1.New()

	if _, err := io.Copy(hash, file); err != nil {
		return returnSHA1String, err
	}

	hashInBytes := hash.Sum(nil)[:20]
	returnSHA1String = hex.EncodeToString(hashInBytes)

	return returnSHA1String, nil
}

// copyFile is designed to copy the local file to the remote system.
//
// It is a little complex because it does two extra things:
//
// * It only copies files if the local/remote differ.
//
// * It optionally expands template-variables.
//
func (e *Evaluator) copyFile(local string, remote string, expand bool) {

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
	hashLocal, err := e.hashFile(local)
	if err != nil {
		fmt.Printf("Failed to hash local file %s\n", err.Error())

		//
		// If we're trying to copy a file that doesn't exist that
		// is a fatal error.
		//
		os.Exit(11)
		return
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
		hashRemote, err := e.hashFile(tmpfile.Name())
		if err != nil {
			fmt.Printf("Failed to hash remote file %s\n", err.Error())

			// If expanding variables we replaced our
			// input-file with the temporary result of
			// expansion.
			if expand {
				os.Remove(local)
			}
			return
		}

		if hashRemote != hashLocal {
			if e.Verbose {
				fmt.Printf("\tFile on remote host needs replacing.\n")
			}

			e.Changed = true
		} else {
			if e.Verbose {
				fmt.Printf("\tFile on remote host doesn't need to be changed.\n")
			}
			e.Changed = false
		}
	} else {

		//
		// If we failed to find the file we
		// assume thati t doesn't exist
		//
		if strings.Contains(err.Error(), "not exist") {
			e.Changed = true
		} else {
			e.Changed = false
		}
	}

	//
	// Upload the file, if it didn't change
	//
	if e.Changed == true {
		err = e.Connection.Upload(local, remote)
		if err != nil {
			fmt.Printf("Failed to upload '%s' to '%s': %s\n", local, remote, err.Error())

			// If expanding variables we replaced our
			// input-file with the temporary result of
			// expansion.
			if expand {
				os.Remove(local)
			}

			return
		}
	}
	// If expanding variables we replaced our
	// input-file with the temporary result of
	// expansion.
	if expand {
		os.Remove(local)
	}
}
