package publics

type ClientPubs struct {
	Id string
	Host string
	Port string
	PubKey int64
}

func NewClientPubs(mixId, host, port string, pubKey int64) ClientPubs{
	clientPubs := ClientPubs{mixId, host, port, pubKey}
	return clientPubs
}