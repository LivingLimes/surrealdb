// Copyright © 2016 Abcum Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/abcum/surreal/log"
)

var exportCmd = &cobra.Command{
	Use:     "export [flags] <file>",
	Short:   "Export data from an existing database",
	Example: "  surreal export --auth root:root backup.db",
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		var fle *os.File
		var req *http.Request
		var res *http.Response

		// Ensure that the command has a filepath
		// as the output file argument. If no filepath
		// has been provided then return an error.

		if len(args) != 1 {
			log.Fatalln("No filepath provided.")
			return
		}

		// Attempt to open or create the specified file
		// in write-only mode, and if there is a problem
		// creating the file, then return an error.

		if fle, err = os.OpenFile(args[0], os.O_CREATE|os.O_WRONLY, 0644); err != nil {
			log.Fatalln("Export failed - please check the filepath and try again.")
			return
		}

		// Ensure that we properly close the file handle
		// when we have finished with the file so that
		// the file descriptor is released.

		defer fle.Close()

		// Check to see if the http request type has
		// been specified as eith 'http' or 'https'
		// as these are the only supported schemes.

		if opts.DB.Type != "http" && opts.DB.Type != "https" {
			log.Fatalln("Connection failed - please specify 'http' or 'https' for the scheme.")
			return
		}

		// Configure the export connection endpoint url
		// and specify the authentication header using
		// basic auth for root login.

		url := fmt.Sprintf("%s://%s@%s:%s/export", opts.DB.Type, opts.Auth.Auth, opts.DB.Host, opts.DB.Port)

		// Create a new http request object that we
		// can use to connect to the export endpoint
		// using a GET http request type.

		if req, err = http.NewRequest("GET", url, nil); err != nil {
			log.Fatalln("Connection failed - check the connection details and try again.")
			return
		}

		// Specify that the request is an octet stream
		// so that we can stream the file contents to
		// the server without reading the whole file.

		req.Header.Set("Content-Type", "application/octet-stream")

		// Attempt to dial the export endpoint and
		// if there is an error then stop execution
		// and return the connection error.

		if res, err = http.DefaultClient.Do(req); err != nil {
			log.Fatalln("Connection failed - check the connection details and try again.")
			return
		}

		// Ensure that we close the body, otherwise
		// if the Body is not closed, the Client can
		// not use the underlying transport again.

		defer res.Body.Close()

		// Ensure that we didn't receive a 401 status
		// code back from the server, otherwise there
		// was a problem with our authentication.

		if res.StatusCode == 401 {
			log.Fatalln("Authentication failed - check the connection details and try again.")
			return
		}

		// Ensure that we received a http 200 status
		// code back from the server, otherwise there
		// was a problem with our request.

		if res.StatusCode != 200 {
			bdy, _ := ioutil.ReadAll(res.Body)
			log.Fatalf("%s", bdy)
			return
		}

		// Copy the http response body to stdOut so
		// that we can pipe the response to other
		// commands or processes.

		if _, err = io.Copy(fle, res.Body); err != nil {
			log.Fatalln("Export failed - there was an error saving the database content.")
			return
		}

		return

	},
}

func init() {

	exportCmd.PersistentFlags().StringVarP(&opts.Auth.Auth, "auth", "a", "root:root", "Master authentication details to use when connecting.")
	exportCmd.PersistentFlags().StringVar(&opts.DB.Type, "scheme", "https", "HTTP connection scheme to use to connect to the database.")
	exportCmd.PersistentFlags().StringVar(&opts.DB.Host, "host", "surreal.io", "Database server host to connect to.")
	exportCmd.PersistentFlags().StringVar(&opts.DB.Port, "port", "80", "Database server port to connect to.")

}
