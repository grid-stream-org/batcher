package destination

type Destination interface {
	Add(data any) error
	Close() error
}
