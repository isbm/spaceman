package utils

import (
	"crypto/tls"
	"errors"
	"github.com/kolo/xmlrpc"
	"io/ioutil"
	"net/http"
	"strings"
)

// RPC client object to call the XML-RPC server
type rpcClient struct {
	url        string
	user       string
	password   string
	session    string
	connection *xmlrpc.Client
	inUse      bool
}

// RPCClient object constructor
func RPCClient() *rpcClient {
	client := new(rpcClient)
	client.session = client.GetSession()
	client.inUse = false

	return client
}

func (client *rpcClient) Connect(config map[string]interface{}) *rpcClient {
	serverConfig, exist := config["server"].(map[interface{}]interface{})
	if !exist {
		Console.CheckError(errors.New("Server configuration section is missing."))
	}

	url, exist := serverConfig["url"].(string)
	if !exist {
		Console.CheckError(errors.New("Server URL must be defined in 'server' section."))
	}
	client.url = url

	user, exist := serverConfig["user"].(string)
	if !exist {
		Console.CheckError(errors.New("User ID must be specified in 'server' section."))
	}
	client.user = user

	password, exist := serverConfig["password"].(string)
	if !exist {
		Console.CheckError(errors.New("Password should be set in 'server' section."))
	}
	client.password = password

	insecure, exist := serverConfig["insecure"].(bool)
	if !exist {
		insecure = false
	}

	client.connection, _ = xmlrpc.NewClient(client.url,
		&http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecure,
			},
		})

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
	client.inUse = true
	client.session = client.RequestFuction("auth.login", client.user, client.password).(string)
	Console.CheckError(client.storeSession())
}

// Request a function call on the remote
func (client *rpcClient) RequestFuction(name string, args ...interface{}) (v interface{}) {
	if client.connection == nil {
		Console.CheckError(errors.New("client is not connected yet"))
	}

	var result interface{}
	err := client.connection.Call(name, args, &result)

	if err != nil {
		if !client.inUse {
			client.auth()
			// Repeat it again with replaced first element, which is always session token
			nArgs := make([]interface{}, len(args))
			nArgs[0] = client.session
			Console.CheckError(client.connection.Call(name, nArgs, &result))
		} else {
			Console.CheckError(err)
		}
	}
	client.inUse = true

	return result
}

var RPC rpcClient

func init() {
	RPC = *RPCClient()
}
