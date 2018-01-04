/*
	Package public implements struct for easy processing and storing of all public information
	of the network participants.
 */

package publics

type MixPubs struct {
	Id     string
	Host   string
	Port   string
	PubKey int64
}

func NewMixPubs(mixId, host, port string, pubKey int64) MixPubs {
	mixPubs := MixPubs{mixId, host, port, pubKey}
	return mixPubs
}
