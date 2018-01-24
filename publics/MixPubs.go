/*
	Package publics implements struct for easy processing and storing of all public information
	of the network participants.
 */

package publics

func NewMixPubs(mixId, host, port string, pubKey []byte) MixPubs {
	mixPubs := MixPubs{Id: mixId, Host: host, Port: port, PubKey: pubKey}
	return mixPubs
}

func NewClientPubs(clientId, host, port string, pubKey []byte, providerInfo MixPubs) ClientPubs {
	client := ClientPubs{Id: clientId, Host: host, Port: port, PubKey: pubKey, Provider : &providerInfo}
	return client
}
