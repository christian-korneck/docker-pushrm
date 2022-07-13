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
	"errors"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"

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
var rfile string
var shortdesc string

// pushrmCmd represents the pushrm command
var pushrmCmd = &cobra.Command{
	Use:     " NAME[:TAG]",
	Aliases: []string{"pushrm"},
	Args:    cobra.MaximumNArgs(1),
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

	Alternatively credentials can be set as environment
	variables. Environment variables take precedence over
	the Docker credentials store.

	Environment variables can be specified with or without
	a server name. The variant without a server name takes
	precedence:

	 - DOCKER_USER and DOCKER_PASS
	 - DOCKER_USER__<SERVER>_<DOMAIN> and DOCKER_PASS__<SERVER>_<DOMAIN>
	   (example for server 'docker.io': DOCKER_USER__DOCKER_IO=my-user
	   and DOCKER_PASS__DOCKER_IO=my-password)

	The provider 'quay' needs an additional env var for the API key
	in form of APIKEY__<SERVERNAME>_<DOMAIN>=<apikey>.


	Dockerhub
	---------
	run 'docker login'

	(use password or Personal Access Token (PAT) with 'admin' scope)


	quay
	----
	- get an api key (bearer token) from the quay webinterface

	  - option 1: env var DOCKER_APIKEY=<apikey>
	    or env var APIKEY__<SERVERNAME>_<DOMAIN>=<apikey>
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

	It's also possible to specify a path to a README file with
	'--file <path>' ('-f <path>').


	Optional [:TAG] argument
	========================

	The [:TAG] argument is optional and has no effect for the currently
	supported providers (which only support a README per repo, not per
	tag). It is in place in case that additional providers get added in
	the future that support READMEs on the tag level.


	Supported environment variables
	===============================
	
	DOCKER_USER, DOCKER_PASS, DOCKER_APIKEY, APIKEY__<SERVER>_<DOMAIN>,
	PUSHRM_PROVIDER, PUSHRM_SHORT, PUSHRM_FILE, PUSHRM_DEBUG, PUSHRM_CONFIG,
	PUSHRM_TARGET

	Commandline parameters take precedence over environment variables.
	Login environment variables take precedence over the local credentials
	store.



`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := run(args); err != nil {
			return err
		}
		return nil
	},
}

func run(args []string) error {
	pushrmProvider := viper.GetString("provider")
	pushrmFile := viper.GetString("file")
	pushrmShortDesc := viper.GetString("short")

	log.Debug("subcommand \"pushrm\" called")

	//fmt.Println(os.Getenv("DOCKER_CLI_PLUGIN_ORIGINAL_CLI_COMMAND"))

	// lowest common ground: 100 runes (not bytes) on Dockerhub
	// this check is intentially global (not per provider) to
	// make cmd calls portable between providers without surprises
	if utf8.RuneCountInString(pushrmShortDesc) > 100 {
		log.Error("Short description is too long (max 100 characters)")
		os.Exit(1)
	}

	// our only positional argument: <servername>/<namespacename>/<reponame>:<tag> (servername + tag are optional)
	targetinfo := os.Getenv("PUSHRM_TARGET")
	if len(args) > 0 {
		if target := args[0]; target != "" {
			targetinfo = target
		}
	}

	if targetinfo == "" {
		return (errors.New("Missing [IMAGE] argument. Example: docker.io/mynamespace/myrepo:latest"))
		//log.Error("Missing [IMAGE] argument. Example: docker.io/mynamespace/myrepo:latest")
		//os.Exit(1)
	}

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

	var erro error

	if pushrmFile == "" {
		pushrmFile, erro = util.FindReadmeFile()
		if erro != nil {
			log.Error(erro)
			os.Exit(1)
		}
	}

	log.Debug("using README file: " + pushrmFile)

	readme, erro := util.ReadFile(pushrmFile)
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
		pushrmProvider = "dockerhub"
	}
	if servername == "quay.io" {
		pushrmProvider = "quay"
	}
	log.Debug("server: ", servername)
	log.Debug("namespace: ", namespacename)
	log.Debug("repo: ", reponame)
	log.Debug("tag: ", tagname)
	log.Debug("repo provider: ", pushrmProvider)

	for _, e := range []string{namespacename, reponame, tagname, servername} {
		// yes, dots are allowed in all these fields
		if regexp.MustCompile(`^[0-9a-zA-Z\-_.]+$`).MatchString(e) == false {
			log.Error("Invalid [IMAGE argument] - bad characters or empty value. Example: docker.io/mynamespace/myrepo:latest")
			os.Exit(1)
		}
	}

	if pushrmProvider == "dockerhub" && servername != "docker.io" {
		log.Error("servername ", servername, " is not valid for provider ", pushrmProvider, " (try \"docker.io\")")
		os.Exit(1)
	}

	var prov provider.Provider

	switch pushrmProvider {
	case "dockerhub":
		prov = dockerhub.Dockerhub{}
	case "quay":
		prov = quay.Quay{}
	case "harbor2":
		prov = harbor2.Harbor2{}
	default:
		log.Error("unsupported repo provider: ", pushrmProvider+". See \"--help\" for supported providers. ")
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

	// generic env var (no servername specified) takes precedence
	dockerUser = os.Getenv("DOCKER_USER")
	dockerPasswd = os.Getenv("DOCKER_PASS")
	if dockerUser != "" && dockerPasswd != "" {
		log.Debug("using credentials for user " + dockerUser + " from generic env var")
	}

	// env var with servername is next
	if dockerUser == "" || dockerPasswd == "" {
		suffix := strings.ToUpper(strings.Replace(servername, ".", "_", -1))
		dockerUser = os.Getenv("DOCKER_USER__" + suffix)
		dockerPasswd = os.Getenv("DOCKER_PASS__" + suffix)
		if dockerUser != "" && dockerPasswd != "" {
			log.Debug("using credentials for user " + dockerUser + " from env var for suffix " + suffix)
		}
	}

	// if credentials are not found in env vars, look in the Docker credentials store
	if (dockerUser == "" || dockerPasswd == "") && authident != "__NONE__" {
		log.Debug("no credentials found in env vars. Trying Docker credentials store")
		log.Debug("Using config file: ", viper.ConfigFileUsed())

		if viper.ConfigFileUsed() == "" {
			log.Error("Docker config file not found. Run \"docker login\" first to create it. ")
			os.Exit(1)
		}

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
	}

	//log.Debug("Using Docker creds: ", dockerUser, " ", dockerPasswd)
	log.Debug("Using Docker creds: ", dockerUser, " ", "********")

	err = prov.Pushrm(servername, namespacename, reponame, tagname, dockerUser, dockerPasswd, readme, pushrmShortDesc)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	return nil

	// ---------
}

func init() {

	const usageTemplate = `
Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}
  `

	const helpTemplate = `
{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}
{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}
  `

	rootCmd.AddCommand(pushrmCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pushrmCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pushrmCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	pushrmCmd.Flags().StringVarP(&providername, "provider", "p", "dockerhub", "repo type: dockerhub, harbor2, quay")
	pushrmCmd.Flags().StringVarP(&rfile, "file", "f", "", "README file (defaults: \"./README-containers.md\", \"./README.md\")")
	pushrmCmd.Flags().StringVarP(&shortdesc, "short", "s", "", "short description (optional)")
	pushrmCmd.Parent().SetUsageTemplate(usageTemplate)
	pushrmCmd.Parent().SetHelpTemplate(helpTemplate)

	viper.BindPFlag("provider", pushrmCmd.Flags().Lookup("provider"))
	viper.BindPFlag("file", pushrmCmd.Flags().Lookup("file"))
	viper.BindPFlag("short", pushrmCmd.Flags().Lookup("short"))
}
