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

package provider

//Provider interface
type Provider interface {
	//GetAuthident - returns the key name under which the provider credentials are stored in Docker's credentials store. Special values: __SERVERNAME__ = use servername, __NONE__ = retrieving credentials will be handled by the provider
	GetAuthident() (authident string)
	//Pushrm function - main provider function, performs the api  call to update the repo description
	Pushrm(servername string, namespacename string, reponame string, tagname string, dockerUser string, dockerPasswd string, readme string, shortdesc string) error
}
