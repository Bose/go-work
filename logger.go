package work

// Logger is used by worker to write logs
type Logger interface {
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Errorf(string, ...interface{})
	Debug(...interface{})
	Info(...interface{})
	Error(...interface{})
}
