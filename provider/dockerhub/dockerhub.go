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

package dockerhub

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

//Dockerhub struct
type Dockerhub struct {
}

//Pushrm is the main provider function
func (f Dockerhub) Pushrm(servername string, namespacename string, reponame string, tagname string, dockerUser string, dockerPasswd string, readme string, shortdesc string) error {

	log.Debug("Dockerhub.Pushrm called")
	jwt, err := GetJwt(dockerUser, dockerPasswd)
	if err != nil {
		log.Debug(err)
		return fmt.Errorf("error trying to get a JWT token from Dockerhub for the stored Docker login. Try \"docker logout\" and \"docker login\". Also, if you have 2FA auth enabled in Dockerhub you'll need to disable it for this tool to work. (This is an unfortunate Dockerhub limitation, see docs for more infos). ")
	}
	err = PatchDescription(jwt, readme, namespacename, reponame, shortdesc)
	if err != nil {
		log.Debug(err)
		return fmt.Errorf("error pushing readme to repo server. See error message below. Run with \"--debug\" for more details. \n\n" + err.Error())
	}

	return nil
}

//GetAuthident returns authident for local Docker credentials store
func (f Dockerhub) GetAuthident() (authident string) {
	log.Debug("Dockerhub.GetAuthident called")
	authident = "https://index.docker.io/v1/"
	return
}

//GetJwt Auth against Dockerhub with user/passwd and request a jwt token
func GetJwt(dockerUser string, dockerPasswd string) (jwt string, error error) {

	url := "https://hub.docker.com/v2/users/login/"
	method := "POST"

	payload := strings.NewReader("{\n    \"username\": \"" + dockerUser + "\",\n    \"password\": \"" + dockerPasswd + "\"\n}")

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		log.Debug(err)
		return "", fmt.Errorf("error retrieving Dockerhub jwt token, error creating http request")
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		log.Debug(err)
		return "", fmt.Errorf("error retrieving Dockerhub jwt token, error making http request")
	}

	log.Debug("retrieve Dockerhub jwt token, status code: ", res.StatusCode)

	if res.StatusCode != 200 {
		return "", fmt.Errorf("error retrieving Dockerhub jwt token, bad status code for response: " + res.Status)

	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Debug(err)
		return "", fmt.Errorf("error retrieving Dockerhub jwt token, error reading response body")
	}

	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		log.Debug(err)
		return "", fmt.Errorf("error retrieving Dockerhub jwt token, error parsing json")
	}

	if dat["token"] == nil || dat["token"].(string) == "" {
		return "", fmt.Errorf("error retrieving Dockerhub jwt token, no jtw token received")
	}

	return dat["token"].(string), nil
}

//PatchDescription - api call to update the repo description
func PatchDescription(jwt string, readme string, namespacename string, reponame string, shortdesc string) (error error) {

	// trailing slash is crucial
	apiurl := "https://hub.docker.com/v2/repositories/" + namespacename + "/" + reponame + "/"
	method := "PATCH"

	bodydata := make(map[string]string)
	bodydata["full_description"] = readme
	if shortdesc != "" {
		bodydata["description"] = shortdesc
	}
	jsonbody, _ := json.Marshal(bodydata)

	payload := strings.NewReader(string(jsonbody))
	client := &http.Client{}
	req, err := http.NewRequest(method, apiurl, payload)
	if err != nil {
		log.Debug(err)
		return fmt.Errorf("error pushing README, error creating http request")
	}

	req.Header.Add("Authorization", "JWT "+jwt)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		log.Debug(err)
		return fmt.Errorf("error pushing README, error creating http request")
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Debug(err)
		return fmt.Errorf("error pushing README, error reading response body")
	}

	log.Debug("push readme, response body: ", string(body))

	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		log.Debug(err)
		return fmt.Errorf("error pushing README, error parsing returned json")
	}

	log.Debug("push README, status code: ", res.StatusCode)

	if res.StatusCode != 200 {
		msg := "error pushing README, bad status code for response: " + res.Status
		if dat["detail"] != nil {
			msg = msg + ". Server responded: \"" + dat["detail"].(string) + "\""
		}
		if res.StatusCode == 403 {
			msg = msg + ". Try \"docker logout\" and \"docker login\". You cannot use a personal access token to log in and must use username and password. If you have 2FA auth enabled in Dockerhub you'll need to disable it for this tool to work. (This is an unfortunate Dockerhub limitation, see docs for more infos.)"

		}
		return fmt.Errorf(msg)

	}

	if dat["full_description"] != readme {
		return fmt.Errorf("error pushing README, pushed readme to repo server but validation failed")
	}

	if shortdesc != "" && dat["description"] != shortdesc {
		return fmt.Errorf("error setting Short Description, pushed to repo server but validation failed")
	}

	log.Debug("content validation successfull, readme successfully pushed to repo server")
	return nil

}
