package anonymous_messaging

type Client struct {
	Id string
	Host string
	IP string
	PubKey int
	PrvKey int
}

type ClientOperations interface {
	EncodeMessage(message string) string
	DecodeMessage(message string) string
}

func (c *Client) EncodeMessage(message string) string {
	return message
}

func (c *Client) DecodeMessage(message string) string {
	return message
}

