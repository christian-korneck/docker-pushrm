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

package cmd

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

// dockerCliPluginMetadataCmd represents the dockerCliPluginMetadata command
var dockerCliPluginMetadataCmd = &cobra.Command{
	Use:   "docker-cli-plugin-metadata",
	Short: "provides plugin metadata to the docker cli",
	Long: `provides plugin metadata to the docker cli
	
	`,
	Run: func(cmd *cobra.Command, args []string) {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "    ")
		// specs: https://docs.docker.com/engine/extend/cli_plugins/#the-docker-cli-plugin-metadata-subcommand
		d := map[string]string{"SchemaVersion": "0.1.0", "Vendor": "Christian Korneck", "Version": "1.6.0", "ShortDescription": "Push Readme to container registry"}
		enc.Encode(d)
	},
}

func init() {
	rootCmd.AddCommand(dockerCliPluginMetadataCmd)
	dockerCliPluginMetadataCmd.Hidden = true

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dockerCliPluginMetadataCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dockerCliPluginMetadataCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
