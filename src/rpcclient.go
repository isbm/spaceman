package main

import (
	"log"

	"github.com/mattn/go-xmlrpc"
)

// RPC client object to call the XML-RPC server
type rpcClient struct {
	url      string
	user     string
	password string
}

// RPCClient object constructor
func RPCClient(url string, user string, password string) *rpcClient {
	client := new(rpcClient)
	client.url = url
	client.user = user
	client.password = password
	return client
}

/*
Request a function call on the remote.
*/
func (client rpcClient) requestFuction(name string, args ...interface{}) (v interface{}) {
	ret, err := xmlrpc.Call(client.url, name, args...)
	if err != nil {
		log.Fatal(err)
	}
	return ret
}

var rpc rpcClient

func init() {
	rpc = *RPCClient("http://localhost:8000", "user", "password") // XXX: todo get creds from the STDIN
}
