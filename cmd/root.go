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
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string
var isDebug bool

var dockerGlobalConfig string
var dockerGlobalContext string
var dockerGlobalHost string
var dockerGlobalTlscacert string
var dockerGlobalLoglevel string
var dockerGlobalTlscert string
var dockerGlobalTlskey string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "docker-pushrm",
	Short: "push README file from current working directory to container registry (Dockerhub, quay, harbor2)",
	Long: `push README file from current working directory to container registry (Dockerhub, quay, harbor2)
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		//at this point we don't have logrus yet, using print instead
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.docker/config.json)")
	rootCmd.PersistentFlags().BoolVarP(&isDebug, "debug", "D", false, "Enable debug mode")

	// these are the docker cli global flags
	// (we define them here so that our plugin doesn't break if they're set, but don't do anything with em)
	rootCmd.PersistentFlags().StringVarP(&dockerGlobalContext, "context", "c", "", "(not supported)")
	rootCmd.PersistentFlags().StringVarP(&dockerGlobalHost, "host", "H", "", "(not supported)")
	rootCmd.PersistentFlags().StringVarP(&dockerGlobalLoglevel, "log-level", "l", "", "(not supported)")
	rootCmd.PersistentFlags().Bool("tls", true, "(not supported)")
	rootCmd.PersistentFlags().Bool("tlsverify", true, "(not supported)")
	rootCmd.PersistentFlags().StringVar(&dockerGlobalTlscacert, "tlscacert", "", "(not supported)")
	rootCmd.PersistentFlags().StringVar(&dockerGlobalTlscert, "tlscert", "", "(not supported)")
	rootCmd.PersistentFlags().StringVar(&dockerGlobalTlskey, "tlskey", "", "(not supported)")

	// hide unsupported flags so that they don't show up with `docker-pushrm pushrm --help`
	rootCmd.PersistentFlags().MarkHidden("context")
	rootCmd.PersistentFlags().MarkHidden("host")
	rootCmd.PersistentFlags().MarkHidden("log-level")
	rootCmd.PersistentFlags().MarkHidden("tls")
	rootCmd.PersistentFlags().MarkHidden("tlsverify")
	rootCmd.PersistentFlags().MarkHidden("tlscacert")
	rootCmd.PersistentFlags().MarkHidden("tlscert")
	rootCmd.PersistentFlags().MarkHidden("tlskey")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if isDebug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})

	log.Debug("root cmd init config")

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Debug(err)
			log.Error("can't find home dir / Docker config file")
			os.Exit(1)
		}
		log.Debug("home dir: ", home)

		// Search config in home directory with name ".docker-pushrm" (without extension).
		viper.AddConfigPath(filepath.Join(home, "/.docker"))

		viper.SetConfigName("config") //filename without .json extension
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
	}

}
