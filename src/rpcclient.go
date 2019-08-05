package main

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/kolo/xmlrpc"
)

// RPC client object to call the XML-RPC server
type rpcClient struct {
	url        string
	user       string
	password   string
	session    string
	connection *xmlrpc.Client
}

// RPCClient object constructor
func RPCClient(url string, user string, password string, insecure bool) *rpcClient {
	client := new(rpcClient)
	client.url = url
	client.user = user
	client.password = password
	client.session, _ = client.getSession()
	client.connection, _ = xmlrpc.NewClient(client.url, &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure}})

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

// Check if login is required
func (client *rpcClient) isAuthRequired(err error) bool {
	// Detect auth failure error in Uyuni :-)
	// Better detection, perhaps?
	return strings.Contains(err.Error(), "Could not find translator for class java.lang.String to interface com.redhat.rhn.domain.user.User") ||
		strings.Contains(err.Error(), "com.redhat.rhn.common.hibernate.LookupException: Could not find session with id")
}

// Request a function call on the remote
func (client *rpcClient) requestFuction(name string, args ...interface{}) (v interface{}) {
	var result interface{}
	err := client.connection.Call(name, args, &result)

	if err != nil && client.isAuthRequired(err) {
		client.auth()
		// Repeat it again with replaced first element, which is always session token
		nArgs := make([]interface{}, len(args))
		nArgs[0] = client.session
		Console.checkError(client.connection.Call(name, nArgs, &result))
	} else {
		Console.checkError(err)
	}

	return result
}

var rpc rpcClient

func init() {
	rpc = *RPCClient("https://suma-refhead-srv.mgr.suse.de/rpc/api", "admin", "admin", true) // XXX: todo get creds from the STDIN
}
