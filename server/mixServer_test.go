package server

import (
	"fmt"
	"os"
	"testing"
)

var mixServer MixServer

func TestMain(m *testing.M) {
	mixServer = *NewMixServer("MixServer", "localhost", "9998", 1, 0, "../pki/database.db")
	fmt.Println(mixServer)

	os.Exit(m.Run())
}
