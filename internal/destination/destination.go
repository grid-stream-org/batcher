package destination

import "context"

type Destination interface {
	Add(ctx context.Context, data any) error
	Close() error
}
