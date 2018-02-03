package server

import (
	"testing"
	"anonymous-messaging/sphinx"
	"anonymous-messaging/node"
	"anonymous-messaging/client"
	"fmt"
	"os"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"anonymous-messaging/clientCore"
	"crypto/elliptic"
	"anonymous-messaging/config"
	"anonymous-messaging/helpers"
	"net"
	"path/filepath"
)

var mixServer *MixServer
var providerServer *ProviderServer
var clientServer *client.Client

const (
	TEST_DATABASE = "testDatabase.db"
)


func createTestProvider() (*ProviderServer, error){
	pub, priv, err := sphinx.GenerateKeyPair()
	if err != nil{
		return nil, err
	}
	node := node.Mix{Id: "Provider", PubKey: pub, PrvKey: priv}
	provider := ProviderServer{Host: "localhost", Port: "9999", Mix: node}
	provider.assignedClients = make(map[string]ClientRecord)
	return &provider, nil
}

func createTestClient(provider config.MixConfig) (*client.Client, error){
	pub, priv, err := sphinx.GenerateKeyPair()
	if err != nil{
		return nil, err
	}
	core := clientCore.CryptoClient{Id: "TestClient", PubKey: pub, PrvKey: priv, Curve: elliptic.P224(), Provider: provider}
	client := &client.Client{Host: "localhost", Port: "9998", CryptoClient: core, PkiDir: TEST_DATABASE}
	addr, err := helpers.ResolveTCPAddress(client.Host, client.Port)
	if err != nil {
		return nil, err
	}

	client.Listener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func clean(){
	clientServer.Listener.Close()
	os.RemoveAll("./inboxes")
}

func TestMain(m *testing.M) {
	var err error
	providerServer, err = createTestProvider()
	clientServer, err = createTestClient(providerServer.Config)
	providerServer.assignedClients[clientServer.Id] = ClientRecord{clientServer.Id,
																	clientServer.Host,
																	clientServer.Port,
																	clientServer.PubKey,
																	[]byte("TestToken")}
	if err != nil{
		fmt.Println(err)
		panic(m)
	}

	code := m.Run()
	clean()
	os.Exit(code)
}



func TestProviderServer_AuthenticateUser_Pass(t *testing.T) {
	testToken := []byte("AuthenticationToken")
	record := ClientRecord{Id: "Alice", Host: "localhost", Port: "1111", PubKey: nil, Token: testToken}
	providerServer.assignedClients["Alice"] = record
	assert.True(t, providerServer.AuthenticateUser("Alice", []byte("AuthenticationToken")), " Authentication should be successful")
}

func TestProviderServer_AuthenticateUser_Fail(t *testing.T) {
	record := ClientRecord{Id: "Alice", Host: "localhost", Port: "1111", PubKey: nil, Token: []byte("AuthenticationToken")}
	providerServer.assignedClients["Alice"] = record
	assert.False(t, providerServer.AuthenticateUser("Alice", []byte("WrongAuthToken")), " Authentication should not be successful")
}

func createInbox(id string, t *testing.T){
	path := filepath.Join("./inboxes", id)
	exists, err := helpers.DirExists(path)
	if err != nil{
		t.Fatal(err)
	}
	if exists {
		os.RemoveAll(path)
		os.MkdirAll(path, 0755)
	} else {
		os.MkdirAll(path, 0755)
	}
}

func createTestMessage(id string, t *testing.T){

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
	createInbox(clientServer.Id, t)
	createTestMessage(clientServer.Id, t)

	signal, err := providerServer.FetchMessages(clientServer.Id)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "SI", signal, " For a non-exisitng inbox id the function should return signal NI")
}


func TestProviderServer_FetchMessages_EmptyInbox(t *testing.T) {
	createInbox("EmptyInbox", t)
	signal, err := providerServer.FetchMessages("EmptyInbox")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "EI", signal, " For an empty inbox id the function should return signal EI")
}

func TestProviderServer_FetchMessages_NoInbox(t *testing.T) {
	signal, err := providerServer.FetchMessages("NonExistingInbox")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "NI", signal, " For a non-exisitng inbox id the function should return signal NI")
}

func TestProviderServer_StoreMessage(t *testing.T) {

	inboxId := "ClientInbox"
	fileId := "12345"
	inboxDir := "./inboxes/" + inboxId
	filePath := inboxDir + "/" + fileId + ".txt"

	err := os.MkdirAll(inboxDir, 0777)
	if err != nil{
		t.Fatal(err)
	}


	message := []byte("Hello world message")
	providerServer.StoreMessage(message, inboxId, fileId)

	_, err = os.Stat(filePath)
	if err != nil{
		t.Fatal(err)
	}
	assert.Nil(t, err, "The file with the message should be created")

	dat, err := ioutil.ReadFile(filePath)
	if err != nil{
		t.Fatal(err)
	}
	assert.Equal(t, message, dat, "Messages should be the same")

}
