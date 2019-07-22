package main

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// RPC client object to call the XML-RPC server
type rpcClient struct {
	url      string
	user     string
	password string
	session  string
}

// RPCClient object constructor
func RPCClient(url string, user string, password string, insecure bool) *rpcClient {
	HttpClient = &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure}}, Timeout: 10 * time.Second}
	client := new(rpcClient)
	client.url = url
	client.user = user
	client.password = password
	client.session, _ = client.getSession()
	return client
}

// Store session into the file
func (client *rpcClient) storeSession() error {
	return ioutil.WriteFile(configuration.session, []byte(client.session+"\n"), 0600)
}

// Get stored session
func (client *rpcClient) getSession() (string, error) {
	var err error
	var session string
	if !fileExists(configuration.session) {
		session = ""
		err = errors.New("Session file does not exists")
	} else {
		data, err := ioutil.ReadFile(configuration.session)
		Console.checkError(err)
		session = strings.TrimSpace(string(data))
	}

	return session, err
}

func (client *rpcClient) auth() {
	client.session = client.requestFuction("auth.login", client.user, client.password).(string)
	Console.checkError(client.storeSession())
}

/*
Request a function call on the remote.
*/
func (client *rpcClient) requestFuction(name string, args ...interface{}) (v interface{}) {
	var ret interface{}
	var err error

	ret, err = Call(client.url, name, args...)
	if err != nil && err.Error() == "invalid response: missing params" {
		client.auth()
		ret, err = Call(client.url, name, args...) // Call it again, refreshed session
	}
	Console.checkError(err)
	return ret
}

var rpc rpcClient

func init() {
	rpc = *RPCClient("http://localhost:8000", "user", "password") // XXX: todo get creds from the STDIN
}
