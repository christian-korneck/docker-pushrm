/*
Copyright © 2020 Christian Korneck <christian@korneck.de>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package main

import (
	"os"

	"github.com/christian-korneck/docker-pushrm/cmd"
	"github.com/christian-korneck/docker-pushrm/util"
)

func main() {

	// if called as a standalone tool (not as a docker cli plugin), proxy directly to the `pushrm` subcommand
	if (os.Getenv("DOCKER_CLI_PLUGIN_ORIGINAL_CLI_COMMAND") == "") && (util.StringInSlice("docker-cli-plugin-metadata", os.Args) == false) {

		newArgs := make([]string, (len(os.Args) + 1))

		newArgs[0] = os.Args[0]
		newArgs[1] = "pushrm"
		for i := 2; i <= len(os.Args); i++ {
			newArgs[i] = os.Args[i-1]
		}
		os.Args = newArgs

	}

	// Cobra, übernehmen Sie
	cmd.Execute()
}
