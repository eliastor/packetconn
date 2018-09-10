package packetconn

import "errors"

//Errors for convinience
var (
	ErrWrongNetwork    = errors.New("Wrong network type")
	ErrTooManyConns    = errors.New("Too many connections")
	ErrInvalidListener = errors.New("Unable to cast to Listener")
)
