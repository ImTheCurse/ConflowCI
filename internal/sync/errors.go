package sync

import (
	"fmt"
)

type CheckSumError struct {
	message string
}

func (e CheckSumError) Error() string {
	return fmt.Sprintf("checksum error: %s", e.message)
}

type MetadataEncodeError struct {
	message string
}

func (e MetadataEncodeError) Error() string {
	return fmt.Sprintf("metadata encode error: %s", e.message)
}
