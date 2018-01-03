package networker

type NetworkClient interface{
	Send(packet string, host string, port string)
}
