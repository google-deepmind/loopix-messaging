package server

import (
	"fmt"
	"os"
	"testing"
	"anonymous-messaging/publics"
)

var mixServer MixServer

func TestMain(m *testing.M) {
	pubM, privM := publics.GenerateKeyPair()
	mixServer = *NewMixServer("MixServer", "localhost", "9998", pubM, privM, "../pki/database.db")
	fmt.Println(mixServer)

	os.Exit(m.Run())
}
