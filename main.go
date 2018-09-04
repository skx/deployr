// deployr
//
// A simple script which will process of recipe for commands.
//
// Commands are generally:
//
//  * Upload a file to the remote host
//
//  * Execute a command
//
// There are a couple of supporting functions to set variables which
// are expanded in the way you might expect ("${blah}" is the contents
// of the variable "blah").
//
// Steve
// --
//

package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/sfreiberg/simplessh"
)

//
// The SSH-client conneciton
//
var client *simplessh.Client

//
// A map of variables which have been initialized/set with "Set".
//
var variables map[string]string

//
// Are we running verbosely?
//
var g_verbose bool

//
// Did the last upload change the remote file?
//
var changed bool

//
// Show a message if running verbosely
//
func log_message(format string, args ...interface{}) {
	if g_verbose == true {
		str := fmt.Sprintf(format, args...)
		fmt.Fprintf(os.Stderr, "%v", str)
	}
}

//
// hash_file returns the SHA1-hash of the contents of the specified file.
//
func hash_file(filePath string) (string, error) {
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
func copyFile(local string, remote string, expand bool) {

	//
	// If we're expanding templates then do that first of all.
	//
	// * Load the file.
	//
	// * Expand the template-variables.
	//
	// * Write to a temporary file.
	//
	// * Swap out the local-file name with the temporary-file.
	//
	if expand {

		//
		// Read the input file.
		//
		data, err := ioutil.ReadFile(local)

		if err != nil {
			log.Fatal("Failed to read local file to expand template-variables %s\n", err.Error())
			return
		}

		//
		// Define a helper-function that users can call to get
		// the variables they've set.
		//
		funcMap := template.FuncMap{
			"get": func(s string) string {
				return (variables[s])
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
		tmpl.Execute(buf, variables)

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
	hash_local, err := hash_file(local)
	if err != nil {
		log.Fatal("Failed to hash local file %s\n", err.Error())
		return
	}

	//
	// Now fetch the file from the remote host, if we
	// can.
	//
	tmpfile, _ := ioutil.TempFile("", "example")
	defer os.Remove(tmpfile.Name()) // clean up

	err = client.Download(remote, tmpfile.Name())
	if err == nil {

		//
		// We had no error - so we now have the
		// remote file copied here.
		//
		hash_remote, err := hash_file(tmpfile.Name())
		if err != nil {
			log.Fatal("Failed to hash remote file %s\n", err.Error())

			// If expanding variables we replaced our
			// input-file with the temporary result of
			// expansion.
			if expand {
				os.Remove(local)
			}
			return
		}

		if hash_remote != hash_local {
			log_message("\tFile on remote host needs replacing.\n")

			changed = true
		} else {
			log_message("\tFile on remote host doesn't need to be changed.\n")
			changed = false
		}
	} else {

		//
		// If we failed to find the file we
		// assume thati t doesn't exist
		//
		if strings.Contains(err.Error(), "not exist") {
			changed = true
		} else {
			changed = false
		}
	}

	//
	// Upload the file, if it didn't change
	//
	if changed == true {
		err = client.Upload(local, remote)
		if err != nil {
			log.Fatal(err)

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

// connecToHost connects to the given host.
//
// The complete-form of the host-string might be:
//   user@host:port
//
// But also "user@host", or "host:port" are acceptible.
func connectToHost(str string) {
	var err error

	if client != nil {
		fmt.Printf("Ignoring request to change target mid-run!\n")
		return
	}
	log_message("Deploying to %s\n", str)

	//
	// Default username + port
	//
	user := "root"
	port := "22"
	host := ""

	//
	// Setup the user if we have it
	//
	if strings.Contains(str, "@") {
		fields := strings.Split(str, "@")
		user = fields[0]
		host = fields[1]
	} else {
		host = str
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
	variables["host"] = host
	variables["port"] = port
	variables["user"] = user

	//
	// Setup our destination with the host/port
	//
	destination := fmt.Sprintf("%s:%s", host, port)

	//
	// Finally connect.
	//
	client, err = simplessh.ConnectWithKeyFile(destination, user, os.Getenv("HOME")+"/.ssh/id_rsa")
	if err != nil {
		panic(err)
	}

}

// processFile processes the given recipe.
//
// Comments could be improved, obviously.
func processFile(filename string) {
	log_message("Processing recipe-file %s\n", filename)

	//
	// Open the file.
	//
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer file.Close()

	//
	// Process the file line-by-line
	//
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		//
		// The line we read.
		//
		line := scanner.Text()

		//
		// Expand any variables which have previously been
		// declared.
		//
		re := regexp.MustCompile("\\$\\{([^\\}]+)\\}")
		line = re.ReplaceAllStringFunc(line, func(in string) string {

			in = strings.TrimPrefix(in, "${")
			in = strings.TrimSuffix(in, "}")

			if len(variables[in]) > 0 {
				return (variables[in])
			} else {
				return "${" + in + "}"
			}
		})

		//
		// Now split the line into tokens, post-expansion.
		//
		tokens := strings.Fields(line)

		//
		// Process each line
		//
		if line == "" {
			//
			// Skip any empty lines.
			//
		} else if strings.HasPrefix(line, "#") {

			//
			// Comments are prefixed with "#"
			//
			continue
		} else if strings.HasPrefix(line, "DeployTo") {

			//
			// DeployTo sets the target to connect to
			//
			connectToHost(tokens[1])
		} else if strings.HasPrefix(line, "Run") {

			//
			// Ensure we're connected
			//
			if client == nil {
				fmt.Printf("Not connected - either add 'DeployTo' to your recipe, or pass a system via\n")
				fmt.Printf("--target=[user@]host.example.com[:port]\n")
				continue
			}

			//
			// Run runs a command.  Unconditionally
			//
			cmd := strings.TrimPrefix(line, "Run ")

			log_message("Running command '%s'\n", cmd)
			result, err := client.Exec(cmd)
			if err != nil {
				log.Fatal(err)
				break
			}

			fmt.Printf("%s", result)
		} else if strings.HasPrefix(line, "IfChanged") {

			//
			// Ensure we're connected
			//
			if client == nil {
				fmt.Printf("Not connected - either add 'DeployTo' to your recipe, or pass a system via\n")
				fmt.Printf("--target=[user@]host.example.com[:port]\n")
				continue
			}

			//
			// IfChanged runs a command only if the
			// previous CopyFile resulted in a change.
			//
			cmd := strings.TrimPrefix(line, "IfChanged ")

			if changed == true {
				log_message("Running command '%s'\n", cmd)
				result, err := client.Exec(cmd)
				if err != nil {
					log.Fatal(err)
					break
				}
				fmt.Printf("%s", result)
			} else {
				log_message("Skipping command - previous copy operation didn't result in a change - %s\n", cmd)
			}
		} else if strings.HasPrefix(line, "CopyTemplate") {

			//
			// Ensure we're connected
			//
			if client == nil {
				fmt.Printf("Not connected - either add 'DeployTo' to your recipe, or pass a system via\n")
				fmt.Printf("--target=[user@]host.example.com[:port]\n")
				continue
			}

			//
			// Local filename
			//
			local := tokens[1]

			//
			// Remote target.
			//
			remote := tokens[2]

			log_message("Copying local file '%s' to remote file '%s'\n", local, remote)

			copyFile(local, remote, true)

		} else if strings.HasPrefix(line, "CopyFile") {

			//
			// Ensure we're connected
			//
			if client == nil {
				fmt.Printf("Not connected - either add 'DeployTo' to your recipe, or pass a system via\n")
				fmt.Printf("--target=[user@]host.example.com[:port]\n")
				continue
			}

			//
			// Local filename
			//
			local := tokens[1]

			//
			// Remote target.
			//
			remote := tokens[2]

			log_message("Copying local file '%s' to remote file '%s'\n", local, remote)

			copyFile(local, remote, false)
		} else if strings.HasPrefix(line, "Set") {

			//
			// Set sets a variable
			//
			key := tokens[1]
			val := tokens[2]

			log_message("Set variable '%s' to '%s'\n", key, val)

			variables[key] = val

		} else {
			fmt.Printf("Unknown input in file: %s\n", line)
			os.Exit(99)
		}
	}

	//
	// Did we have any error(s) in our scanner?
	//
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
		return
	}

	//
	// Close the SSH connection, if it was opened.
	//
	if client != nil {
		client.Close()
	}
}

// main is our entry-point
func main() {

	//
	// Setup space to hold any variables which are defined.
	//
	variables = make(map[string]string)

	//
	// Process our command-line arguments, if any.
	//
	target := flag.String("target", "", "The target host to deploy to")
	verbose := flag.Bool("verbose", false, "Should we run with extra detail")
	flag.Parse()

	//
	// Save the global verbosity-flag.
	//
	g_verbose = *verbose

	//
	// If we received a target then connect now.
	//
	if *target != "" {
		connectToHost(*target)
	}

	//
	// Process each file on the command-line
	//
	for _, arg := range flag.Args() {
		processFile(arg)
	}

}
