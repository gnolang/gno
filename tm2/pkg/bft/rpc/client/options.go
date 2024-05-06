package client

import "time"

type Option func(client *RPCClient)

// WithRequestTimeout sets the request timeout
func WithRequestTimeout(timeout time.Duration) Option {
	return func(client *RPCClient) {
		client.requestTimeout = timeout
	}
}
