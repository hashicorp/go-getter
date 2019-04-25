package getter

// A ClientOption allows to configure a client
type ClientOption func(*Client) error

// Configure configures a client with options.
func (c *Client) Configure(opts ...ClientOption) error {
	c.Options = opts
	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return err
		}
	}
	// Default decompressor values
	if c.Decompressors == nil {
		c.Decompressors = Decompressors
	}
	// Default detector values
	if c.Detectors == nil {
		c.Detectors = Detectors
	}
	// Default getter values
	if c.Getters == nil {
		c.Getters = Getters
	}

	for _, getter := range c.Getters {
		getter.SetClient(c)
	}
	return nil
}
