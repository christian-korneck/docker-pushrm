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

package quay

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/christian-korneck/docker-pushrm/util"
	log "github.com/sirupsen/logrus"
)

//Quay struct
type Quay struct {
}

//Pushrm is the main provider function
func (f Quay) Pushrm(servername string, namespacename string, reponame string, tagname string, dockerUser string, dockerPasswd string, readme string, shortdesc string) error {

	if shortdesc != "" {
		log.Warn("Short description not supported for provider \"quay\". Ignoring.")
	}

	log.Debug("Quay.Pushrm called")

	apikey, err := util.GetApikey(servername)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	//log.Debug("apikey: " + apikey)
	log.Debug("apikey: " + "********")

	err = PatchDescription(apikey, readme, servername, namespacename, reponame)
	if err != nil {
		log.Debug(err)
		return fmt.Errorf("error pushing readme to repo server. See error message below. Run with \"--debug\" for more details. \n\n" + err.Error())
	}

	return nil
}

//GetAuthident returns authident for local Docker credentials store
func (f Quay) GetAuthident() (authident string) {
	log.Debug("Quay.GetAuthident called")
	authident = "__NONE__"
	return
}

//PatchDescription - api call to update the repo description
func PatchDescription(quaytoken string, readme string, servername string, namespacename string, reponame string) (error error) {

	apiurl := "https://" + servername + "/api/v1/repository/" + namespacename + "/" + reponame
	method := "PUT"

	jsonbody, _ := json.Marshal(map[string]string{"description": readme})
	payload := strings.NewReader(string(jsonbody))

	client := &http.Client{}
	req, err := http.NewRequest(method, apiurl, payload)
	if err != nil {
		log.Debug(err)
		return fmt.Errorf("error pushing README, error creating http request")
	}

	req.Header.Add("Authorization", "Bearer "+quaytoken)
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
	// we need the json response only in case of failure to get the error message
	if res.StatusCode != 200 {
		if err := json.Unmarshal(body, &dat); err != nil {
			log.Debug(err)
		}
	}

	log.Debug("push README, status code: ", res.StatusCode)

	if res.StatusCode != 200 {
		msg := "error pushing README, bad status code for response: " + res.Status
		if dat["error_message"] != nil {

			msg = msg + ". Server responded: \"" + dat["error_message"].(string) + "\""
		}
		if res.StatusCode == 403 {
			msg = msg + ". Try \"docker logout\" and \"docker login\""

		}
		return fmt.Errorf(msg)

	} else {
		log.Debug("status code OK, readme successfully pushed to repo server")
		return nil
	}

}
