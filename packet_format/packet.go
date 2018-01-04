package packet_format

import (
	"anonymous-messaging/publics"
	"encoding/json"
	"fmt"
)

type Header struct {
	Meta  MetaData
	Delay float64
}

type MetaData struct {
	NextHopId   string
	NextHopHost string
	NextHopPort string
	StoreFlag   bool
	FinalFlag   bool
}

type Packet struct {
	Message string
	Path    []publics.MixPubs
	Delays  []float64
	Steps   map[string]Header
}

type Packer interface {
	Encode(message string, path []publics.MixPubs, delays []float64)
	Decode(packet Packet)
}

func Encode(message string, path []publics.MixPubs, delays []float64) Packet {
	steps := make(map[string]Header)
	for i := 0; i < len(path); i++ {
		if i+1 >= len(path) {
			mdata := MetaData{NextHopId: "", NextHopHost: "", NextHopPort: "", StoreFlag: false, FinalFlag: false}
			header := Header{mdata, delays[i]}
			steps[path[i].Id] = header
		} else {
			mdata := MetaData{NextHopId: path[i+1].Id, NextHopHost: path[i+1].Host, NextHopPort: path[i+1].Port, StoreFlag: false, FinalFlag: true}
			header := Header{mdata, delays[i]}
			steps[path[i].Id] = header
		}
	}
	p := Packet{Message: message, Delays: delays, Path: path, Steps: steps}
	return p
}

func Decode(packet Packet) Packet {
	return packet
}

func ToString(p Packet) string {
	s, err := json.Marshal(p)

	if err != nil {
		fmt.Println("Error in encoding packet: ", err)
	}
	return string(s)
}

func FromString(s string) Packet {
	var packet Packet
	err := json.Unmarshal([]byte(s), &packet)

	if err != nil {
		fmt.Println("Error in decode: ", err)
	}
	return packet
}

func NewPacket(message string, delays []float64, path []publics.MixPubs, steps map[string]Header) Packet {
	return Packet{Message: message, Delays: delays, Path: path, Steps: steps}
}
