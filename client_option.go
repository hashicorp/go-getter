package getter

// A ClientOption allows to configure a client
type ClientOption func(*Client) error

// Configure configures a client with options.
func (c *Client) Configure(opts ...ClientOption) error {
	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return err
		}
	}
	return nil
}
