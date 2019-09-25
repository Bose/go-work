package work

// Args is how parameters are passed to jobs
type Args map[string]interface{}

// Handler is executed a a Work for a given Job.  It also defines the interface
// for handlers that can be used by middleware adapters
type Handler func(j *Job) error

// Adapter defines the adaptor middleware type
type Adapter func(Handler) Handler

// Adapt a handler with provided middlware adapters
func Adapt(h Handler, adapters ...Adapter) Handler {
	for _, adapter := range adapters {
		h = adapter(h)
	}
	return h
}
