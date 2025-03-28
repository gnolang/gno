package cluster

import "context"

type Cluster struct {
	ctx      context.Context
	cancelFn context.CancelFunc

	config Config
}

func New(ctx context.Context, config Config) (*Cluster, error) {
	ctx, cancelFn := context.WithCancel(ctx)

	c := &Cluster{
		ctx:      ctx,
		cancelFn: cancelFn,
		config:   config,
	}

	// TODO implement image building

	return c, nil
}

func (c *Cluster) Shutdown() error {
	// Stop the top-level cluster context
	c.cancelFn()

	// TODO wait on nodes to finish

	return nil
}
