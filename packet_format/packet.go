package packet_format

import (
	"encoding/json"
	"fmt"
)

type Packet struct {
	Message string
	Path []string
	Delays []float64
}

func Encode(message string, path []string, delays []float64) Packet{
	p := Packet{Message:message, Delays:delays, Path:path}
	return p
}

func Decode(packet Packet) Packet{
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

	if err !=nil {
		fmt.Println("Error in decode: ", err)
	}
	return packet
}

func NewPacket(message string, delays []float64, path []string) Packet{
	return Packet{Message:message, Delays:delays, Path:path}
}