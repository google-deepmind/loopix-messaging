/*
	Package config implements struct for easy processing and storing of all public information
	of the network participants.
 */

package config

import (
	"github.com/protobuf/proto"
)

func NewMixConfig(mixId, host, port string, pubKey []byte) MixConfig {
	MixConfig := MixConfig{Id: mixId, Host: host, Port: port, PubKey: pubKey}
	return MixConfig
}

func NewClientConfig(clientId, host, port string, pubKey []byte, providerInfo MixConfig) ClientConfig {
	client := ClientConfig{Id: clientId, Host: host, Port: port, PubKey: pubKey, Provider : &providerInfo}
	return client
}

/*
	WrapWithFlag packs the given byte information together with a specified flag into the
	packet.
 */
func WrapWithFlag(flag string, data []byte) ([]byte, error){
	m := GeneralPacket{Flag: flag, Data: data}
	mBytes, err := proto.Marshal(&m)
	if err !=nil {
		return nil, err
	}
	return mBytes, nil
}

type E2EPath struct {
	IngressProvider MixConfig
	Mixes []MixConfig
	EgressProvider MixConfig
	Recipient ClientConfig
}

func (p *E2EPath) Len() int {
	return 3 + len(p.Mixes)
}

