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

package util

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

// BytesToString converts
func BytesToString(b []byte) string {
	return string(b[:])
}

// StringInSlice checks if a string exists in a slice
func StringInSlice(checkval string, list []string) bool {
	for _, b := range list {
		if b == checkval {
			return true
		}
	}
	return false
}

// ReadFile reads a file
func ReadFile(path string) (filecontent string, error error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Debug(err)
		return "", fmt.Errorf("could not read README file: " + path)
	}
	text := string(content)
	return text, nil
}

//GetApikey retrieves an API key from env var or the local Docker config file
func GetApikey(servername string) (apikey string, error error) {
	log.Debug("util.GetApikey called")

	genericEnvval := os.Getenv("DOCKER_APIKEY")
	if genericEnvval != "" {
		return genericEnvval, nil
	}

	envkey := "APIKEY__" + strings.ToUpper(strings.Replace(servername, ".", "_", -1))
	querykey := "plugins.docker-pushrm.apikey_" + servername

	envval := os.Getenv(envkey)

	if envval != "" {
		apikey = envval
		return apikey, nil
	} else {
		cfgval := viper.GetString(querykey)
		if cfgval != "" {
			apikey = cfgval
			return apikey, nil
		} else {
			return "", fmt.Errorf("could not find api key for server " + servername + ". Either specify env var DOCKER_APIKEY or env var " + envkey + " or " + querykey + " in the local Docker config file. ")
		}

	}

}

//GetDockerCreds retrieves credentials from the Docker creds store
func GetDockerCreds(authident string, authidentIsFuzzy bool) (dockerUser string, dockerPasswd string, error error) {

	var candidates []string
	if authidentIsFuzzy == true {
		candidates = []string{authident, ("https://" + authident), ("https://" + authident + "/"), ("http://" + authident), ("http://" + authident + "/")}
	} else {
		candidates = []string{authident}
	}
	for _, candidate := range candidates {
		dockerUser, dockerPasswd = "", ""
		dockerUser, dockerPasswd, err := QueryDockerCreds(candidate)
		if err != nil {
			log.Debug("tried candidate " + candidate + ", got error: " + err.Error())
		}

		if dockerUser == "" && dockerPasswd == "" {
			log.Debug("tried candidate " + candidate + ": could not find credentials")
		} else {
			log.Debug("tried candidate " + candidate + ": found credentials for user " + dockerUser)
			return dockerUser, dockerPasswd, nil
		}

	}

	return "", "", fmt.Errorf("no Docker credentials found for this server/provider. Run 'docker login' first. ")

}

//QueryDockerCreds fetches credentials for an authid
func QueryDockerCreds(authident string) (dockerUser string, dockerPasswd string, error error) {

	log.Debug("util.GetDockerCreds called")

	if viper.GetString("auths."+authident+".auth") != "" {

		credsb64 := viper.GetString("auths." + authident + ".auth")
		credsclearb, err := base64.StdEncoding.DecodeString(credsb64)
		if err != nil {
			log.Debug(err)
			return "", "", fmt.Errorf("Error parsing auth info from the Docker config file. Check your local Docker config. ")
		}
		credsclear := string(credsclearb)
		credssplit := strings.Split(credsclear, ":")
		dockerUser = credssplit[0]
		dockerPasswd = credsclear[len(dockerUser)+1:]

	} else {
		if viper.GetString("credsStore") != "" {
			executable := "docker-credential-" + viper.GetString("credsStore")
			if viper.GetString("credsStore") == "wincred" {
				executable = executable + ".exe"
			}
			shx := exec.Command(executable, "get")
			stdin, err := shx.StdinPipe()
			if err != nil {
				log.Debug(err)
				return "", "", fmt.Errorf("Error executing the Docker credentials helper. Check your local Docker config and/or installation. ")
			}

			done := make(chan bool)

			go func() {
				defer func() { done <- true }()
				defer stdin.Close()
				io.WriteString(stdin, authident)

			}()

			<-done

			out, err := shx.CombinedOutput()
			if err != nil {
				log.Debug(err)
				return "", "", fmt.Errorf("no Docker credentials found for this server/provider. Run 'docker login' first. ")
			}

			var dat map[string]interface{}
			if err := json.Unmarshal(out, &dat); err != nil {
				log.Debug(err)
				return "", "", fmt.Errorf("Error parsing credentials from Docker creds provider. Run 'docker login' first. ")
			}
			dockerUser = dat["Username"].(string)
			dockerPasswd = dat["Secret"].(string)

		} else {
			return "", "", fmt.Errorf("no Docker credentials found for this server/provider. Run 'docker login' first. ")
		}
	}

	return dockerUser, dockerPasswd, nil
}

//FindReadmeFile trys to find a readme file in the cwd
func FindReadmeFile() (foundfile string, error error) {
	preferedfilenames := []string{"./README-containers.md", "./README.md"} //prefer these filenames in this order

	for _, preferedfilename := range preferedfilenames {
		matches, err := filepath.Glob(preferedfilename)
		if err != nil {
			log.Debug(err)
			return "", fmt.Errorf("error while searching for default readme file")
		}
		if len(matches) >= 1 {
			foundfile = preferedfilename
			break
		}
	}

	if foundfile == "" {
		matches, err := filepath.Glob("./[R|r][E|e][A|a][D|d][M|m][E|e]*")
		if err != nil {
			log.Debug(err)
			return "", fmt.Errorf("error while searching for alternate readme file")
		}
		if len(matches) < 1 {
			return "", fmt.Errorf("README file not found in the current working directory. Create a file \"README-containers.md\" or \"README.md\" or \"cd\" into a directory that contains a README file. ")
		} else {
			foundfile = matches[0]
		}
	}

	return foundfile, nil
}
