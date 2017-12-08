package anonymous_messaging

import "fmt"

type Provider struct {
	Mix
}

func (p Provider) StorePacket(packet string) {
	fmt.Println(packet)
}
