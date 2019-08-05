package utils

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
	client.session = client.GetSession()
	client.connection, _ = xmlrpc.NewClient(client.url, &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure}})

	return client
}

// Store session into the file
func (client *rpcClient) storeSession() error {
	return ioutil.WriteFile(Configuration.GetSessionConfFilePath(), []byte(client.session+"\n"), 0600)
}

// Get stored session
func (client *rpcClient) GetSession() string {
	var session string
	if !fileExists(Configuration.GetSessionConfFilePath()) {
		session = ""
		Console.CheckError(errors.New("Session file does not exists"))
	} else {
		data, err := ioutil.ReadFile(Configuration.GetSessionConfFilePath())
		Console.CheckError(err)
		session = strings.TrimSpace(string(data))
	}

	return session
}

func (client *rpcClient) auth() {
	client.session = client.RequestFuction("auth.login", client.user, client.password).(string)
	Console.CheckError(client.storeSession())
}

// Check if login is required
func (client *rpcClient) isAuthRequired(err error) bool {
	// Detect auth failure error in Uyuni :-)
	// Better detection, perhaps?
	return strings.Contains(err.Error(), "Could not find translator for class java.lang.String to interface com.redhat.rhn.domain.user.User") ||
		strings.Contains(err.Error(), "com.redhat.rhn.common.hibernate.LookupException: Could not find session with id")
}

// Request a function call on the remote
func (client *rpcClient) RequestFuction(name string, args ...interface{}) (v interface{}) {
	var result interface{}
	err := client.connection.Call(name, args, &result)

	if err != nil && client.isAuthRequired(err) {
		client.auth()
		// Repeat it again with replaced first element, which is always session token
		nArgs := make([]interface{}, len(args))
		nArgs[0] = client.session
		Console.CheckError(client.connection.Call(name, nArgs, &result))
	} else {
		Console.CheckError(err)
	}

	return result
}

var RPC rpcClient

func init() {
	RPC = *RPCClient("https://suma-refhead-srv.mgr.suse.de/rpc/api", "admin", "admin", true) // XXX: todo get creds from the STDIN
}
