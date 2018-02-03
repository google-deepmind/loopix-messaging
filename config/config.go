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

func MixConfigToBytes(pubs MixConfig) ([]byte, error) {
	data, err := proto.Marshal(&pubs)

	if err != nil {
		return nil, err
	}
	return data, nil
}

func MixConfigFromBytes(b []byte) (MixConfig, error) {
	var pubs MixConfig
	err := proto.Unmarshal(b, &pubs)
	if err != nil {
		return pubs, err
	}
	return pubs, nil
}

func ClientConfigToBytes(pubs ClientConfig) ([]byte, error) {
	data, err := proto.Marshal(&pubs)

	if err != nil {
		return nil, err
	}
	return data, nil
}

func ClientConfigFromBytes(b []byte) (ClientConfig, error) {
	var pubs ClientConfig
	err := proto.Unmarshal(b, &pubs)
	if err != nil {
		return pubs, err
	}
	return pubs, nil
}

func GeneralPacketToBytes(pkt GeneralPacket) ([]byte, error) {
	data, err := proto.Marshal(&pkt)

	if err != nil {
		return nil, err
	}
	return data, nil
}

func GeneralPacketFromBytes(b []byte) (GeneralPacket, error) {
	var pkt GeneralPacket
	err := proto.Unmarshal(b, &pkt)
	if err != nil {
		return pkt, err
	}
	return pkt, nil
}

func WrapWithFlag(flag string, data []byte) ([]byte, error){
	m := GeneralPacket{Flag: flag, Data: data}
	mBytes, err := GeneralPacketToBytes(m)
	if err !=nil {
		return nil, err
	}
	return mBytes, nil
}

func PullRequestToBytes(r PullRequest) ([]byte, error) {
	data, err := proto.Marshal(&r)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func PullRequestFromBytes(b []byte) (PullRequest, error) {
	var r PullRequest
	err := proto.Unmarshal(b, &r)
	if err != nil {
		return r, err
	}
	return r, nil
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