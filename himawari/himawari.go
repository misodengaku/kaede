package himawari

func NewClient(host string, port int) (*Client, error) {
	client := Client{
		ServerHost: host,
		ServerPort: port,
	}

	return &client, nil
}
