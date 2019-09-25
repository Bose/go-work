package work

import (
	"fmt"
	"reflect"
)

// WrapChanel takes a concrete receiving chan in as an interface{}, and wraps it with an interface{} chan
// so you can treat all receiving channels the same way
func WrapChannel(chanToWrap interface{}) (<-chan interface{}, error) {
	t := reflect.TypeOf(chanToWrap)
	if t.Kind() != reflect.Chan {
		return nil, fmt.Errorf("WrapChannel: not a channel")
	}
	chanDirection := t.ChanDir()
	if chanDirection == reflect.RecvDir {
		return nil, fmt.Errorf("WrapChannel: is not a receiving channel")
	}
	// we don't want to buffer this wrapper
	wrapper := make(chan interface{})

	go func() {
		rawChan := reflect.ValueOf(chanToWrap)
		for {
			rawValue, ok := rawChan.Recv()
			if !ok {
				close(wrapper)
				return
			}
			wrapper <- rawValue.Interface()
		}
	}()
	return wrapper, nil
}
