package work

// GetOpts - iterate the inbound Options and return a struct
func GetOpts(opt ...Option) Options {
	opts := getDefaultOptions()
	for _, o := range opt {
		o(opts)
	}
	return opts
}

// Option - how Options are passed as arguments
type Option func(Options)

// Options = how options are represented
type Options map[string]interface{}

func getDefaultOptions() Options {
	return Options{
		optionWithSync: false,
	}
}

const optionWithSync = "optionWithSync"

// WithSync optional synchronous execution
func WithSync(sync bool) Option {
	return func(o Options) {
		o[optionWithSync] = sync
	}
}

const optionWithJob = "optionWithJob"

// WithJob optional Job parameter
func WithJob(j *Job) Option {
	return func(o Options) {
		o[optionWithJob] = j
	}
}

const optionWithChannel = "optionWithChannel"

// WithChannel optional channel parameter
func WithChannel(ch interface{}) Option {
	return func(o Options) {
		o[optionWithChannel] = ch
	}
}

// Get a specific option by name
func (o *Options) Get(name string) (interface{}, bool) {
	if d, ok := (*o)[name]; ok {
		return d, ok
	}
	return nil, false
}
