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

package harbor2

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

//Harbor2 struct
type Harbor2 struct {
}

//Pushrm is the main provider function
func (f Harbor2) Pushrm(servername string, namespacename string, reponame string, tagname string, dockerUser string, dockerPasswd string, readme string) error {
	log.Debug("Harbor2.Pushrm called")

	err := PatchDescription(dockerUser, dockerPasswd, readme, servername, namespacename, reponame)
	if err != nil {
		log.Debug(err)
		return fmt.Errorf("error pushing readme to repo server. See error message below. Run with \"--debug\" for more details. \n\n" + err.Error())
	}

	return nil
}

//GetAuthident returns authident for local Docker credentials store
func (f Harbor2) GetAuthident() (authident string) {
	log.Debug("Harbor2.GetAuthident called")
	authident = "__SERVERNAME__"
	return
}

//PatchDescription - api call to update the repo description
func PatchDescription(dockerUser string, dockerPasswd string, readme string, servername string, namespacename string, reponame string) (error error) {

	apiurl := "https://" + servername + "/api/v2.0/projects/" + namespacename + "/repositories/" + reponame
	method := "PUT"

	jsonbody, _ := json.Marshal(map[string]string{"description": readme})
	payload := strings.NewReader(string(jsonbody))

	client := &http.Client{}
	req, err := http.NewRequest(method, apiurl, payload)
	if err != nil {
		log.Debug(err)
		return fmt.Errorf("error pushing README, error creating http request")
	}

	creds := base64.StdEncoding.EncodeToString([]byte(dockerUser + ":" + dockerPasswd))
	req.Header.Add("Authorization", "Basic "+creds)
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

	log.Debug("push readme, response body: " + string(body))

	var dat map[string]interface{}
	// json only returned in case of failure for error message, otherwise empty
	if res.StatusCode != 200 {
		if err := json.Unmarshal(body, &dat); err != nil {
			log.Debug(err)
		}
	}

	log.Debug("push README, status code: ", res.StatusCode)

	if res.StatusCode != 200 {
		msg := "error pushing README, bad status code for response: " + res.Status
		if dat["errors"] != nil {

			firsterror := dat["errors"].([]interface{})[0].(map[string]interface{})
			msg = msg + ". Server responded: \"" + firsterror["code"].(string) + " - " + firsterror["message"].(string) + "\""
		}
		if res.StatusCode == 403 {
			msg = msg + ". Try \"docker logout\" and \"docker login\". "

		}
		return fmt.Errorf(msg)

	} else {
		log.Debug("status code OK, readme successfully pushed to repo server")
		return nil
	}

}
