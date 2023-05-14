package errors

import "errors"

var (
	ErrBadCredentials = errors.New("provided credential are wrong")
	ErrTimeout        = errors.New("operation has timeouted")
	ErrFailedIRCSend  = errors.New("failed to send irc message")
)
