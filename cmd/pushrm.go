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
	"os"
	"regexp"
	"strings"

	"github.com/christian-korneck/docker-pushrm/provider/dockerhub"
	"github.com/christian-korneck/docker-pushrm/provider/harbor2"
	"github.com/christian-korneck/docker-pushrm/provider/provider"
	"github.com/christian-korneck/docker-pushrm/provider/quay"
	"github.com/christian-korneck/docker-pushrm/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var providername string

// pushrmCmd represents the pushrm command
var pushrmCmd = &cobra.Command{
	Use:     " NAME[:TAG]",
	Aliases: []string{"pushrm"},
	Args:    cobra.ExactArgs(1),
	Short:   "push README file from current working directory to container registry (Dockerhub, quay, harbor2)",
	Long: `help for docker pushrm
	
	docker pushrm NAME[:TAG] [flags]
	
	pushes the README.md file from the current working
	directory to the container registry (Dockerhub, quay, harbor2)
	where it appears as repo description.



	Usage Examples:
	===============


	Dockerhub (hub.docker.com cloud)
	--------------------------------
	docker pushrm myaccount/hello-world
	docker pushrm docker.io/my-account/hello-world


	Quay (quay.io cloud or self-hosted)
	-----------------------------------
	docker pushrm quay.io/my-organization/hello-world
	docker pushrm --provider quay my-quay-server.com/my-user/hello-world


	Harbor (self-hosted)
	--------------------
	docker pushrm --provider harbor2 my-harbor-server.com/my-project/hello-world



	How to login
	=============


	For most registry providers docker-pushrm uses
	the registry login from Docker's credentials store.
	For some providers an API key needs to be specified
	via env var or Docker config file.


	Dockerhub
	---------
	run 'docker login'


	quay
	----
	- get an api key (bearer token) from the quay webinterface

	  - option 1: env var APIKEY__<SERVERNAME>_<DOMAIN>=<apikey> 
		(example for quay.io: 'export APIKEY__QUAY_IO=myapikey')
		
	  - option 2: in the Docker config file (default: "$HOME/.docker/config.json") 
		add the json key 'plugins.docker-pushrm.apikey_<servername>' with the 
		apikey as string value (example for quay.io: key name
		'plugins.docker-pushrm.apikey_quay.io')
	
	Env var takes precedence.
	

	harbor
	------
	run 'docker login <servername>' (example: 'docker login demo.goharbor.io')



	Creating a README file
	=======================

	Valid names: 'README.md' or 'README-containers.md'.

	'README-containers.md' takes precedence. This allows to set up a
	different README file for the container registry then for the git repo.
	
	The README file needs to be in the current working directory from which
	docker pushrm is being called.


	Optional [:TAG] argument
	========================

	The [:TAG] argument is optional and has no effect for the currently
	supported providers (which only support a README per repo, not per
	tag). It is in place in case that additional providers get added in
	the future that support READMEs on the tag level.


`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Debug("subcommand \"pushrm\" called")
		log.Debug("Using config file: ", viper.ConfigFileUsed())

		if viper.ConfigFileUsed() == "" {
			log.Error("Docker config file not found. Run \"docker login\" first to create it. ")
			os.Exit(1)
		}

		//fmt.Println(os.Getenv("DOCKER_CLI_PLUGIN_ORIGINAL_CLI_COMMAND"))

		// our only positional argument: <servername>/<namespacename>/<reponame>:<tag> (servername + tag are optional)
		targetinfo := args[0]
		// fail if namespacename is missing
		if len(strings.Split(targetinfo, "/")) < 2 {
			log.Error("Invalid [IMAGE] argument - missing namespace. Example: docker.io/mynamespace/myrepo:latest")
			os.Exit(1)
		}
		// fill up default servername, if missing
		if len(strings.Split(targetinfo, "/")) < 3 {
			targetinfo = "docker.io/" + targetinfo
		}
		// fill up default tagname, if missing
		if strings.Contains(targetinfo, ":") != true {
			targetinfo = targetinfo + ":latest"
		}
		log.Debug("Using target: ", targetinfo)

		foundfile, erro := util.FindReadmeFile()
		if erro != nil {
			log.Error(erro)
			os.Exit(1)
		}

		log.Debug("using README file: " + foundfile)

		readme, erro := util.ReadFile(foundfile)
		if erro != nil {
			log.Error(erro)
			os.Exit(1)
		}

		if (len(strings.Split(targetinfo, "/")) != 3) || (len(strings.Split(strings.Split(targetinfo, "/")[2], ":")) != 2) {
			log.Error("Invalid [IMAGE] argument - too many separators. Example: docker.io/mynamespace/myrepo:latest")
			os.Exit(1)
		}

		servername := strings.ToLower(strings.Split(targetinfo, "/")[0])
		namespacename := strings.Split(targetinfo, "/")[1]
		reponame := strings.Split(strings.Split(targetinfo, "/")[2], ":")[0]
		tagname := strings.Split(strings.Split(targetinfo, "/")[2], ":")[1]
		if servername == "docker.io" {
			providername = "dockerhub"
		}
		if servername == "quay.io" {
			providername = "quay"
		}
		log.Debug("server: ", servername)
		log.Debug("namespace: ", namespacename)
		log.Debug("repo: ", reponame)
		log.Debug("tag: ", tagname)
		log.Debug("repo provider: ", providername)

		for _, e := range []string{namespacename, reponame, tagname, servername} {
			// yes, dots are allowed in all these fields
			if regexp.MustCompile(`^[0-9a-zA-Z\-_.]+$`).MatchString(e) == false {
				log.Error("Invalid [IMAGE argument] - bad characters or empty value. Example: docker.io/mynamespace/myrepo:latest")
				os.Exit(1)
			}
		}

		if providername == "dockerhub" && servername != "docker.io" {
			log.Error("servername ", servername, " is not valid for provider ", providername, " (try \"docker.io\")")
			os.Exit(1)
		}

		var prov provider.Provider

		switch providername {
		case "dockerhub":
			prov = dockerhub.Dockerhub{}
		case "quay":
			prov = quay.Quay{}
		case "harbor2":
			prov = harbor2.Harbor2{}
		default:
			log.Error("unsupported repo provider: ", providername+". See \"--help\" for supported providers. ")
			os.Exit(1)
		}

		authident := prov.GetAuthident()
		var authidentIsFuzzy bool
		authidentIsFuzzy = false

		if authident == "__SERVERNAME__" {
			authident = servername
			authidentIsFuzzy = true
		}

		var dockerUser string
		var dockerPasswd string
		var err error
		// a provider can request to handle auth itself with authident __NONE__
		if authident != "__NONE__" {
			dockerUser, dockerPasswd, err = util.GetDockerCreds(authident, authidentIsFuzzy)
			if err != nil {
				log.Error(err)
				os.Exit(1)
			}
		} else {
			dockerUser = ""
			dockerPasswd = ""
		}

		//log.Debug("Using Docker creds: ", dockerUser, " ", dockerPasswd)
		log.Debug("Using Docker creds: ", dockerUser, " ", "********")

		err = prov.Pushrm(servername, namespacename, reponame, tagname, dockerUser, dockerPasswd, readme)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		// ---------
	},
}

func init() {
	rootCmd.AddCommand(pushrmCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pushrmCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pushrmCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	pushrmCmd.Flags().StringVarP(&providername, "provider", "p", "dockerhub", "repo type: dockerhub, harbor2, quay")

}
