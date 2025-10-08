package ssh

import (
	"errors"
)

var ErrNotSupported = errors.New("Authentication method not supported")
var ErrPrivKetFileNotFound = errors.New("Public key file was not found")
var ErrPrivateKeyParse = errors.New("Private key file could not be parsed")
var ErrEmptyPrivKeyPath = errors.New("Private key path is empty")
