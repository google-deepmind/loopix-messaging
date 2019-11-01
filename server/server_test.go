// Copyright 2018 The Loopix-Messaging Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"loopix-messaging/config"
	"loopix-messaging/helpers"
	"loopix-messaging/node"
	"loopix-messaging/sphinx"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"crypto/elliptic"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"testing"
)

var mixServer *MixServer
var providerServer *ProviderServer

const (
	testDatabase = "testDatabase.db"
)

func createTestProvider() (*ProviderServer, error) {
	pub, priv, err := sphinx.GenerateKeyPair()
	if err != nil {
		return nil, err
	}
	node := node.NewMix(pub, priv)
	provider := ProviderServer{host: "localhost", port: "9999", Mix: node}
	provider.config = config.MixConfig{Id: provider.id, Host: provider.host, Port: provider.port, PubKey: provider.GetPublicKey()}
	provider.assignedClients = make(map[string]ClientRecord)
	return &provider, nil
}

func createTestMixnode() (*MixServer, error) {
	pub, priv, err := sphinx.GenerateKeyPair()
	if err != nil {
		return nil, err
	}
	node := node.NewMix(pub, priv)
	mix := MixServer{host: "localhost", port: "9995", Mix: node}
	mix.config = config.MixConfig{Id: mix.id, Host: mix.host, Port: mix.port, PubKey: mix.GetPublicKey()}
	addr, err := helpers.ResolveTCPAddress(mix.host, mix.port)
	if err != nil {
		return nil, err
	}

	mix.listener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &mix, nil
}

func createFakeClientListener(host, port string) (*net.TCPListener, error) {
	addr, err := helpers.ResolveTCPAddress(host, port)
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}
	return listener, nil
}

func clean() {
	os.RemoveAll("./inboxes")
}

func TestMain(m *testing.M) {
	var err error
	mixServer, err = createTestMixnode()
	if err != nil {
		fmt.Println(err)
		panic(m)
	}

	providerServer, err = createTestProvider()
	if err != nil {
		fmt.Println(err)
		panic(m)
	}

	code := m.Run()
	clean()
	os.Exit(code)
}

func TestProviderServer_AuthenticateUser_Pass(t *testing.T) {
	testToken := []byte("AuthenticationToken")
	record := ClientRecord{id: "Alice", host: "localhost", port: "1111", pubKey: nil, token: testToken}
	providerServer.assignedClients["Alice"] = record
	assert.True(t, providerServer.authenticateUser("Alice", []byte("AuthenticationToken")), " Authentication should be successful")
}

func TestProviderServer_AuthenticateUser_Fail(t *testing.T) {
	record := ClientRecord{id: "Alice", host: "localhost", port: "1111", pubKey: nil, token: []byte("AuthenticationToken")}
	providerServer.assignedClients["Alice"] = record
	assert.False(t, providerServer.authenticateUser("Alice", []byte("WrongAuthToken")), " Authentication should not be successful")
}

func createInbox(id string, t *testing.T) {
	path := filepath.Join("./inboxes", id)
	exists, err := helpers.DirExists(path)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		os.RemoveAll(path)
		os.MkdirAll(path, 0755)
	} else {
		os.MkdirAll(path, 0755)
	}
}

func createTestMessage(id string, t *testing.T) {

	file, err := os.Create(filepath.Join("./inboxes", id, "TestMessage.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	_, err = file.Write([]byte("This is a test message"))
	if err != nil {
		t.Fatal(err)
	}

}

func TestProviderServer_FetchMessages_FullInbox(t *testing.T) {
	clientListener, err := createFakeClientListener("localhost", "9999")
	defer clientListener.Close()

	providerServer.assignedClients["FakeClient"] = ClientRecord{"FakeClient",
		"localhost",
		"9999",
		[]byte("FakePublicKey"),
		[]byte("TestToken")}

	createInbox("FakeClient", t)
	createTestMessage("FakeClient", t)

	signal, err := providerServer.fetchMessages("FakeClient")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "SI", signal, " For inbox containing messages the signal should be SI")
}

func TestProviderServer_FetchMessages_EmptyInbox(t *testing.T) {
	createInbox("EmptyInbox", t)
	signal, err := providerServer.fetchMessages("EmptyInbox")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "EI", signal, " For an empty inbox id the function should return signal EI")
}

func TestProviderServer_FetchMessages_NoInbox(t *testing.T) {
	signal, err := providerServer.fetchMessages("NonExistingInbox")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "NI", signal, " For a non-existing inbox id the function should return signal NI")
}

func TestProviderServer_StoreMessage(t *testing.T) {

	inboxId := "ClientInbox"
	fileId := "12345"
	inboxDir := "./inboxes/" + inboxId
	filePath := inboxDir + "/" + fileId + ".txt"

	err := os.MkdirAll(inboxDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	message := []byte("Hello world message")
	providerServer.storeMessage(message, inboxId, fileId)

	_, err = os.Stat(filePath)
	if err != nil {
		t.Fatal(err)
	}
	assert.Nil(t, err, "The file with the message should be created")

	dat, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, message, dat, "Messages should be the same")

}

func TestProviderServer_HandlePullRequest_Pass(t *testing.T) {
	testPullRequest := config.PullRequest{ClientId: "PassTestId", Token: []byte("TestToken")}
	providerServer.assignedClients["PassTestId"] = ClientRecord{id: "TestId", host: "localhost", port: "1111", pubKey: nil, token: []byte("TestToken")}
	bTestPullRequest, err := proto.Marshal(&testPullRequest)
	if err != nil {
		t.Error(err)
	}
	err = providerServer.handlePullRequest(bTestPullRequest)
	if err != nil {
		t.Error(err)
	}
}

func TestProviderServer_HandlePullRequest_Fail(t *testing.T) {
	testPullRequest := config.PullRequest{ClientId: "FailTestId", Token: []byte("TestToken")}
	providerServer.assignedClients = map[string]ClientRecord{}
	bTestPullRequest, err := proto.Marshal(&testPullRequest)
	if err != nil {
		t.Error(err)
	}
	err = providerServer.handlePullRequest(bTestPullRequest)
	assert.EqualError(t, errors.New("authentication went wrong"), err.Error(), "HandlePullRequest should return an error if authentication failed")
}

func TestProviderServer_RegisterNewClient(t *testing.T) {
	newClient := config.ClientConfig{Id: "NewClient", Host: "localhost", Port: "9998", PubKey: nil}
	bNewClient, err := proto.Marshal(&newClient)
	if err != nil {
		t.Fatal(err)
	}
	token, addr, err := providerServer.registerNewClient(bNewClient)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "localhost:9998", addr, "Returned address should be the same as registered client address")
	assert.Equal(t, helpers.SHA256([]byte("TMP_Token"+"NewClient")), token, "Returned token should be equal to the hash of clients id")

	path := fmt.Sprintf("./inboxes/%s", "NewClient")
	exists, err := helpers.DirExists(path)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, exists, "When a new client is registered an inbox should be created")
}

func TestProviderServer_HandleAssignRequest(t *testing.T) {
	clientListener, err := createFakeClientListener("localhost", "9999")
	defer clientListener.Close()

	newClient := config.ClientConfig{Id: "ClientXYZ", Host: "localhost", Port: "9999", PubKey: nil}
	bNewClient, err := proto.Marshal(&newClient)
	if err != nil {
		t.Fatal(err)
	}
	err = providerServer.handleAssignRequest(bNewClient)
	if err != nil {
		t.Fatal(err)
	}
}

func createTestPacket(t *testing.T) *sphinx.SphinxPacket {
	path := config.E2EPath{IngressProvider: providerServer.config, Mixes: []config.MixConfig{mixServer.config}, EgressProvider: providerServer.config}
	sphinxPacket, err := sphinx.PackForwardMessage(elliptic.P224(), path, []float64{0.1, 0.2, 0.3}, "Hello world")
	if err != nil {
		t.Fatal(err)
		return nil
	}
	return &sphinxPacket
}

func TestProviderServer_ReceivedPacket(t *testing.T) {
	sphinxPacket := createTestPacket(t)
	bSphinxPacket, err := proto.Marshal(sphinxPacket)
	if err != nil {
		t.Fatal(err)
	}
	err = providerServer.receivedPacket(bSphinxPacket)
	if err != nil {
		t.Fatal(err)
	}
}

func TestProviderServer_HandleConnection(t *testing.T) {
	serverConn, _ := net.Pipe()
	errs := make(chan error, 1)
	// serverConn.Write([]byte("test"))
	go func() {
		providerServer.handleConnection(serverConn, errs)
		err := <-errs
		if err != nil {
			t.Fatal(err)
		}
		serverConn.Close()
	}()
	serverConn.Close()

}
