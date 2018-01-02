package anonymous_messaging

type NetworkClient interface{
	Send(packet string, host string, port string)
}
