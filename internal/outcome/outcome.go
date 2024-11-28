package outcome

type Outcome interface {
	ProjectID() string
	CurrentOutput() float64
	Data() map[string][]any
}
